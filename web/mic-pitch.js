// オタマトーンの音をマイクで拾い、ピッチ(周波数)をGo/WASM側に送って
// Link(プレイヤー)を前後に移動させる。ピッチ検出そのもの(pitch.js)は
// pitch-test.htmlの検証結果を流用している。

const micPitchStatusEl = document.getElementById('mic-pitch-status');
const btnStartMicPitch = document.getElementById('btn-start-mic-pitch');
const micPitchValueEl = document.getElementById('mic-pitch-value');
const micDirectionEl = document.getElementById('mic-direction');
const micThresholdSlider = document.getElementById('mic-threshold-slider');
const micThresholdValueEl = document.getElementById('mic-threshold-value');

const DIRECTION_LABELS = {
  forward: '前進 ↑',
  backward: '後退 ↓',
  idle: '停止',
};

let micAudioCtx = null;
let micAnalyser = null;
let micBuffer = null;

micThresholdValueEl.textContent = micThresholdSlider.value + ' dB';
micThresholdSlider.addEventListener('input', () => {
  micThresholdValueEl.textContent = micThresholdSlider.value + ' dB';
});

function micPitchFrame(timestamp) {
  micAnalyser.getFloatTimeDomainData(micBuffer);

  const rms = computeRMS(micBuffer);
  const db = rmsToDb(rms);
  const thresholdDb = Number(micThresholdSlider.value);

  let freq = -1;
  if (db >= thresholdDb) {
    const detected = detectPitch(micBuffer, micAudioCtx.sampleRate);
    if (detected > 0 && detected < 5000) {
      freq = detected;
    }
  }

  if (window.goOnPitchDetected) {
    window.goOnPitchDetected(freq);
  }

  if (freq > 0) {
    micPitchValueEl.textContent = freq.toFixed(1) + ' Hz (' + frequencyToNoteName(freq) + ')';
  } else {
    micPitchValueEl.textContent = '-- Hz';
  }

  if (window.goGetPlayerDirection) {
    const direction = window.goGetPlayerDirection();
    micDirectionEl.textContent = DIRECTION_LABELS[direction] || direction;
  }

  requestAnimationFrame(micPitchFrame);
}

btnStartMicPitch.addEventListener('click', async () => {
  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    micPitchStatusEl.textContent = 'このブラウザはマイク入力に対応していません';
    return;
  }

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    micAudioCtx = new (window.AudioContext || window.webkitAudioContext)();
    const source = micAudioCtx.createMediaStreamSource(stream);
    micAnalyser = micAudioCtx.createAnalyser();
    micAnalyser.fftSize = 2048;
    micBuffer = new Float32Array(micAnalyser.fftSize);
    source.connect(micAnalyser);

    micPitchStatusEl.textContent = 'マイク入力を受信中(オタマトーンの音でLinkが前後に動きます)';
    btnStartMicPitch.disabled = true;
    requestAnimationFrame(micPitchFrame);
  } catch (err) {
    micPitchStatusEl.textContent = 'マイクの取得に失敗しました: ' + err;
  }
});
