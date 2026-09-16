// エンディング画面のマイク入力の音量しきい値(dB)。他フィールドのような
// 移動操作の微調整は不要な演出用途のため、スライダーは設けず固定値に
// している(web/mic-pitch.jsのデフォルト値と同じ)。
const ENDING_MIC_THRESHOLD_DB = -30;

const btnStartEndingMic = document.getElementById('btn-start-ending-mic');
const endingMicStatusEl = document.getElementById('ending-mic-status');

// endingAudioCtx は、マイク入力の解析(startEndingMic)で使うAudioContext。
// ページを開いた瞬間にできるだけ早く使えるよう、モジュール読み込み時点
// (ボタンクリックを待たず)に生成しておく。
const endingAudioCtx = new (window.AudioContext || window.webkitAudioContext)();

let endingMicAnalyser = null;
let endingMicBuffer = null;

// endingMicFrame は、マイクの波形からピッチ(周波数)を検出し、
// window.goOnEndingPitchDetected経由でGo側(花びらを舞わせる演出、
// game.OnEndingPitchDetected参照)へ渡す処理を毎フレーム繰り返す。
// ピッチ検出そのもの(computeRMS/rmsToDb/detectPitch)はpitch.js
// (他フィールドと共通)を使う。
function endingMicFrame() {
  endingMicAnalyser.getFloatTimeDomainData(endingMicBuffer);

  const rms = computeRMS(endingMicBuffer);
  const db = rmsToDb(rms);

  let freq = -1;
  if (db >= ENDING_MIC_THRESHOLD_DB) {
    const detected = detectPitch(endingMicBuffer, endingAudioCtx.sampleRate);
    if (detected > 0 && detected < 5000) {
      freq = detected;
    }
  }

  if (window.goOnEndingPitchDetected) {
    window.goOnEndingPitchDetected(freq);
  }

  requestAnimationFrame(endingMicFrame);
}

async function startEndingMic() {
  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    endingMicStatusEl.textContent = 'このブラウザはマイク入力に対応していません';
    return;
  }

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    const source = endingAudioCtx.createMediaStreamSource(stream);
    endingMicAnalyser = endingAudioCtx.createAnalyser();
    endingMicAnalyser.fftSize = 2048;
    endingMicBuffer = new Float32Array(endingMicAnalyser.fftSize);
    source.connect(endingMicAnalyser);

    endingMicStatusEl.textContent = 'マイク入力を受信中(音程を大きく上下させると花びらが舞います)';
    btnStartEndingMic.disabled = true;
    requestAnimationFrame(endingMicFrame);
  } catch (err) {
    endingMicStatusEl.textContent = 'マイクの取得に失敗しました: ' + err;
  }
}

btnStartEndingMic.addEventListener('click', startEndingMic);

// ページに入った時点で、ボタンを押さなくても最初からマイクを有効にする
// (getUserMedia自体はユーザー操作無しでも呼べるため、ブラウザの権限
// ダイアログが出るだけで済む。既に拒否されている等で失敗した場合は
// catch内でステータス表示され、ボタンから手動で再試行できる)。
startEndingMic();

if (!WebAssembly) {
  console.error("このブラウザはWebAssemblyに対応していません");
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("ending.wasm"), go.importObject)
    .then((result) => {
      console.log("エンディング画面: Go WASM 起動完了");
      go.run(result.instance);
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}
