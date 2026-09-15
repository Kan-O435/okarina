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

// computeRMS / rmsToDb / detectPitch / frequencyToNoteName は pitch.js で定義。

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
