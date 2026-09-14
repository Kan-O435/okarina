// オカリナの発音を担当する。Web Audio APIでMIDIノートを単純な波形で鳴らす。
//
// ブラウザの自動再生ポリシーにより、AudioContextはユーザー操作(クリック等)を
// 経てからでないと音が鳴らない場合があるため、ensureAudioContext()は
// ボタンクリック等のイベントハンドラ内から呼ぶ必要がある。

const AudioContextClass = window.AudioContext || window.webkitAudioContext;

let audioCtx = null;
const activeOscillators = new Map(); // MIDIノート番号 -> { oscillator, gainNode }

function ensureAudioContext() {
  if (!AudioContextClass) {
    return null;
  }
  if (!audioCtx) {
    audioCtx = new AudioContextClass();
  }
  if (audioCtx.state === 'suspended') {
    audioCtx.resume();
  }
  return audioCtx;
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

  const oscillator = ctx.createOscillator();
  const gainNode = ctx.createGain();

  oscillator.type = 'sine';
  oscillator.frequency.value = noteToFrequency(note);

  const volume = Math.min(velocity / 127, 1) * 0.3;
  const now = ctx.currentTime;
  gainNode.gain.setValueAtTime(0, now);
  gainNode.gain.linearRampToValueAtTime(volume, now + 0.02);

  oscillator.connect(gainNode);
  gainNode.connect(ctx.destination);
  oscillator.start();

  activeOscillators.set(note, { oscillator, gainNode });
}

// stopNote はMIDIノート番号に対応する音を止める。
function stopNote(note) {
  const active = activeOscillators.get(note);
  if (!active || !audioCtx) {
    return;
  }
  activeOscillators.delete(note);

  const { oscillator, gainNode } = active;
  const now = audioCtx.currentTime;
  gainNode.gain.cancelScheduledValues(now);
  gainNode.gain.setValueAtTime(gainNode.gain.value, now);
  gainNode.gain.linearRampToValueAtTime(0, now + 0.05);
  oscillator.stop(now + 0.06);
}
