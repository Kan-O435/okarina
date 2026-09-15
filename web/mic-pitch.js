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

// startMicPitch は、マイクの取得〜ピッチ検出ループの開始までを行う。
// ボタンクリック(手動)からも、ページ読み込み時の自動開始(下記)からも
// 呼べるよう関数として切り出している。
async function startMicPitch() {
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
}

btnStartMicPitch.addEventListener('click', startMicPitch);

// このオリジン(同じホスト・ポート)へのマイクアクセスをブラウザがすでに
// 「許可」済みの場合、ページ遷移のたびに手動でボタンを押さなくても
// 自動的にマイク入力を開始する。ブラウザの許可自体はオリジン単位で
// 記憶されるため本来ページ遷移のたびに聞かれ直すことは無いはずだが、
// (ローカル開発でポート番号が毎回変わっている場合はオリジンが変わって
// しまい許可が引き継がれない点に注意)、少なくともこちら側で「毎回
// ボタンを押す」手間は無くすためのもの。Permissions APIの'microphone'に
// 対応していないブラウザ(Safari等)では何もせず、従来通りボタンでの
// 開始のみをサポートする。
if (navigator.permissions && navigator.permissions.query) {
  navigator.permissions.query({ name: 'microphone' })
    .then((status) => {
      if (status.state === 'granted') {
        startMicPitch();
      }
    })
    .catch(() => {
      // 'microphone'照会に対応していない場合はここに来る。ボタンでの
      // 手動開始にフォールバックする。
    });
}
