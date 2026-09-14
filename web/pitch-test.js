// オタマトーンの音声をマイクで取得し、音程(周波数)を数値化できるか
// 検証するためのプロトタイプ。ゲーム本体(index.html/main.js)とは独立している。

const micStatusEl = document.getElementById('mic-status');
const btnStartMic = document.getElementById('btn-start-mic');
const pitchDisplayEl = document.getElementById('pitch-display');
const noteDisplayEl = document.getElementById('note-display');
const volumeDisplayEl = document.getElementById('volume-display');
const volumeMeterEl = document.getElementById('volume-meter');
const thresholdSlider = document.getElementById('threshold-slider');
const thresholdValueEl = document.getElementById('threshold-value');

const AudioContextClass = window.AudioContext || window.webkitAudioContext;
const NOTE_NAMES = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'];

// メーター表示に使う音量の範囲(dB)。この範囲外は振り切れとして扱う。
const METER_MIN_DB = -60;
const METER_MAX_DB = 0;

let audioCtx = null;
let analyser = null;
let buffer = null;

thresholdValueEl.textContent = thresholdSlider.value + ' dB';
thresholdSlider.addEventListener('input', () => {
  thresholdValueEl.textContent = thresholdSlider.value + ' dB';
});

// computeRMS は波形データの実効値(音量の大きさ)を計算する。
function computeRMS(buf) {
  let sum = 0;
  for (let i = 0; i < buf.length; i++) {
    sum += buf[i] * buf[i];
  }
  return Math.sqrt(sum / buf.length);
}

// rmsToDb はRMSをデシベルに変換する(0dB = 振幅1相当、値が小さいほど負に大きくなる)。
function rmsToDb(rms) {
  return 20 * Math.log10(Math.max(rms, 1e-8));
}

function frequencyToNoteName(freq) {
  const noteNumber = 12 * Math.log2(freq / 440) + 69;
  const rounded = Math.round(noteNumber);
  const name = NOTE_NAMES[((rounded % 12) + 12) % 12];
  const octave = Math.floor(rounded / 12) - 1;
  return name + octave;
}

// 自己相関法で波形の基本周波数(ピッチ)を推定する。
// 音量による足切り(音量が小さければ無視する)は呼び出し側(update)で行う。
function detectPitch(buf, sampleRate) {
  const size = buf.length;

  // 振幅が閾値を下回る位置で波形の前後をトリムし、無音部分の影響を減らす。
  const threshold = 0.2;
  let start = 0;
  for (let i = 0; i < size / 2; i++) {
    if (Math.abs(buf[i]) < threshold) {
      start = i;
      break;
    }
  }
  let end = size - 1;
  for (let i = 1; i < size / 2; i++) {
    if (Math.abs(buf[size - i]) < threshold) {
      end = size - i;
      break;
    }
  }

  const trimmed = buf.slice(start, end);
  const n = trimmed.length;
  if (n < 2) {
    return -1;
  }

  const correlation = new Float32Array(n);
  for (let lag = 0; lag < n; lag++) {
    let sum = 0;
    for (let i = 0; i < n - lag; i++) {
      sum += trimmed[i] * trimmed[i + lag];
    }
    correlation[lag] = sum;
  }

  // 最初の下り坂を抜けた後の最大ピークを、基本周期の候補として探す。
  let d = 0;
  while (d < n - 1 && correlation[d] > correlation[d + 1]) {
    d++;
  }
  let maxValue = -1;
  let maxLag = -1;
  for (let i = d; i < n; i++) {
    if (correlation[i] > maxValue) {
      maxValue = correlation[i];
      maxLag = i;
    }
  }
  if (maxLag <= 0) {
    return -1;
  }

  // 放物線補間でピーク位置を細かく補正する。
  let period = maxLag;
  const x1 = correlation[maxLag - 1] || 0;
  const x2 = correlation[maxLag];
  const x3 = correlation[maxLag + 1] || 0;
  const a = (x1 + x3 - 2 * x2) / 2;
  const b = (x3 - x1) / 2;
  if (a !== 0) {
    period = maxLag - b / (2 * a);
  }

  return sampleRate / period;
}

function update() {
  analyser.getFloatTimeDomainData(buffer);

  const rms = computeRMS(buffer);
  const db = rmsToDb(rms);
  const thresholdDb = Number(thresholdSlider.value);

  volumeDisplayEl.textContent = db.toFixed(1) + ' dB';
  const meterPercent = Math.min(
    Math.max((db - METER_MIN_DB) / (METER_MAX_DB - METER_MIN_DB), 0),
    1
  ) * 100;
  volumeMeterEl.style.width = meterPercent + '%';
  volumeMeterEl.classList.toggle('above-threshold', db >= thresholdDb);

  if (db < thresholdDb) {
    pitchDisplayEl.textContent = '-- Hz';
    noteDisplayEl.textContent = '(音量が小さいため無視)';
    requestAnimationFrame(update);
    return;
  }

  const freq = detectPitch(buffer, audioCtx.sampleRate);

  if (freq > 0 && freq < 5000) {
    pitchDisplayEl.textContent = freq.toFixed(1) + ' Hz';
    noteDisplayEl.textContent = frequencyToNoteName(freq);
  } else {
    pitchDisplayEl.textContent = '-- Hz';
    noteDisplayEl.textContent = '--';
  }

  requestAnimationFrame(update);
}

btnStartMic.addEventListener('click', async () => {
  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    micStatusEl.textContent = 'このブラウザはマイク入力に対応していません';
    return;
  }

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    audioCtx = new AudioContextClass();
    const source = audioCtx.createMediaStreamSource(stream);
    analyser = audioCtx.createAnalyser();
    analyser.fftSize = 2048;
    buffer = new Float32Array(analyser.fftSize);
    source.connect(analyser);

    micStatusEl.textContent = 'マイク入力を受信中';
    requestAnimationFrame(update);
  } catch (err) {
    micStatusEl.textContent = 'マイクの取得に失敗しました: ' + err;
  }
});
