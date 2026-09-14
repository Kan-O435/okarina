// オカリナの発音を担当する。Web Audio APIでMIDIノートを音にする。
//
// ただのサイン波だと電子音っぽくなるため、オカリナらしさを出すために
// 「息のノイズ(アタック時のシュッという音、鳴っている間もわずかに残る)」
// 「かすかな倍音」「ビブラート(息の揺れ)」を加えている。
//
// ブラウザの自動再生ポリシーにより、AudioContextはユーザー操作(クリック等)を
// 経てからでないと音が鳴らない場合があるため、ensureAudioContext()は
// ボタンクリック等のイベントハンドラ内から呼ぶ必要がある。

const AudioContextClass = window.AudioContext || window.webkitAudioContext;

let audioCtx = null;
let noiseBuffer = null;
const activeNotes = new Map(); // MIDIノート番号 -> 再生中のノード一式

function ensureAudioContext() {
  if (!AudioContextClass) {
    return null;
  }
  if (!audioCtx) {
    audioCtx = new AudioContextClass();
    noiseBuffer = createNoiseBuffer(audioCtx);
  }
  if (audioCtx.state === 'suspended') {
    audioCtx.resume();
  }
  return audioCtx;
}

// createNoiseBuffer は息のノイズに使うホワイトノイズのバッファを作る。
function createNoiseBuffer(ctx) {
  const durationSeconds = 1;
  const buffer = ctx.createBuffer(1, ctx.sampleRate * durationSeconds, ctx.sampleRate);
  const data = buffer.getChannelData(0);
  for (let i = 0; i < data.length; i++) {
    data[i] = Math.random() * 2 - 1;
  }
  return buffer;
}

// noteToFrequency はMIDIノート番号を周波数(Hz)に変換する(平均律、A4=440Hz基準)。
function noteToFrequency(note) {
  return 440 * Math.pow(2, (note - 69) / 12);
}

// playNote はMIDIノート番号に対応する音を鳴らし始める。
// オカリナは単音楽器なので、同じノートが既に鳴っていれば鳴らし直す。
function playNote(note, velocity) {
  const ctx = ensureAudioContext();
  if (!ctx) {
    return;
  }
  stopNote(note);

  const now = ctx.currentTime;
  const freq = noteToFrequency(note);
  const volume = Math.min(velocity / 127, 1) * 0.3;

  // 基音(サイン波)
  const oscillator = ctx.createOscillator();
  oscillator.type = 'sine';
  oscillator.frequency.value = freq;

  // かすかな倍音(丸さを保ちつつ少し厚みを出す)
  const overtone = ctx.createOscillator();
  overtone.type = 'sine';
  overtone.frequency.value = freq * 2;
  const overtoneGain = ctx.createGain();
  overtoneGain.gain.value = volume * 0.08;

  // ビブラート(息の揺れ)。基音・倍音の周波数を微妙に揺らす。
  const vibrato = ctx.createOscillator();
  vibrato.type = 'sine';
  vibrato.frequency.value = 5.5;
  const vibratoDepth = ctx.createGain();
  vibratoDepth.gain.value = freq * 0.006;
  vibrato.connect(vibratoDepth);
  vibratoDepth.connect(oscillator.frequency);
  vibratoDepth.connect(overtone.frequency);

  // 息のノイズ。アタック時は強め、鳴っている間はごく弱く残す。
  const noiseSource = ctx.createBufferSource();
  noiseSource.buffer = noiseBuffer;
  noiseSource.loop = true;
  const noiseFilter = ctx.createBiquadFilter();
  noiseFilter.type = 'bandpass';
  noiseFilter.frequency.value = freq;
  noiseFilter.Q.value = 1;
  const noiseGain = ctx.createGain();
  noiseGain.gain.setValueAtTime(volume * 0.6, now);
  noiseGain.gain.linearRampToValueAtTime(volume * 0.05, now + 0.15);
  noiseSource.connect(noiseFilter);
  noiseFilter.connect(noiseGain);

  const gainNode = ctx.createGain();
  gainNode.gain.setValueAtTime(0, now);
  gainNode.gain.linearRampToValueAtTime(volume, now + 0.03);

  oscillator.connect(gainNode);
  overtone.connect(overtoneGain);
  overtoneGain.connect(gainNode);
  noiseGain.connect(gainNode);
  gainNode.connect(ctx.destination);

  oscillator.start(now);
  overtone.start(now);
  vibrato.start(now);
  noiseSource.start(now);

  activeNotes.set(note, { oscillator, overtone, vibrato, noiseSource, gainNode });
}

// stopNote はMIDIノート番号に対応する音を止める。
function stopNote(note) {
  const active = activeNotes.get(note);
  if (!active || !audioCtx) {
    return;
  }
  activeNotes.delete(note);

  const { oscillator, overtone, vibrato, noiseSource, gainNode } = active;
  const now = audioCtx.currentTime;
  gainNode.gain.cancelScheduledValues(now);
  gainNode.gain.setValueAtTime(gainNode.gain.value, now);
  gainNode.gain.linearRampToValueAtTime(0, now + 0.08);

  oscillator.stop(now + 0.1);
  overtone.stop(now + 0.1);
  vibrato.stop(now + 0.1);
  noiseSource.stop(now + 0.1);
}
