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

// endingFanfareBuffer/LoadPromiseは、このページに入った際に流すファン
// ファーレ(assets/audio/ending-fanfare.mp3)を事前にデコードしたバッファ。
// web/grassland.js等の効果音と同じ方式(事前デコード+自己修復
// フォールバック)。
let endingFanfareBuffer = null;
let endingFanfareLoadPromise = null;
let endingFanfarePlayed = false;

// loadEndingFanfare は、ファンファーレファイルの取得・デコードを開始する
// (すでに開始・完了済みなら何もしない)。デコード済みのAudioBufferで
// 解決するPromiseを返す。
function loadEndingFanfare() {
  if (endingFanfareBuffer) {
    return Promise.resolve(endingFanfareBuffer);
  }
  if (endingFanfareLoadPromise) {
    return endingFanfareLoadPromise;
  }
  endingFanfareLoadPromise = fetch('assets/audio/ending-fanfare.mp3')
    .then((res) => res.arrayBuffer())
    .then((data) => endingAudioCtx.decodeAudioData(data))
    .then((buffer) => {
      endingFanfareBuffer = buffer;
      return buffer;
    })
    .catch((err) => {
      console.error('ending fanfare: failed to load/decode', err);
      endingFanfareLoadPromise = null; // 失敗時は次回呼び出しでもう一度試す
      return null;
    });
  return endingFanfareLoadPromise;
}

// playEndingFanfare はGoから、このページに入った直後に呼ばれる。ブラウザの
// 自動再生ポリシーにより、ユーザー操作を経ていないAudioContextは無音のまま
// (state: 'suspended')になる場合があるため、resume()を試みてから再生する。
// それでも鳴らせなかった場合に備え、ページ内の最初のクリック/タップ/キー
// 入力で改めて再生を試みるフォールバックも用意している
// (setupEndingFanfareAutoplayFallback参照)。
function playEndingFanfare() {
  if (endingFanfarePlayed) {
    return;
  }
  Promise.all([endingAudioCtx.resume().catch(() => {}), loadEndingFanfare()]).then(([, buffer]) => {
    if (buffer && !endingFanfarePlayed && endingAudioCtx.state === 'running') {
      endingFanfarePlayed = true;
      const source = endingAudioCtx.createBufferSource();
      source.buffer = buffer;
      source.connect(endingAudioCtx.destination);
      source.start(0);
    }
  });
}

// setupEndingFanfareAutoplayFallback は、自動再生がブロックされた場合に
// 備え、ページ内の最初のクリック/タップ/キー入力で改めて再生を試みる。
function setupEndingFanfareAutoplayFallback() {
  const retry = () => {
    if (endingFanfarePlayed) {
      return;
    }
    playEndingFanfare();
  };
  ['pointerdown', 'keydown'].forEach((type) => {
    document.addEventListener(type, retry, { once: true });
  });
}

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

    endingMicStatusEl.textContent = 'マイク入力を受信中';
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

// ファンファーレの再生自体はGo側(game.PlayEndingFanfare、
// cmd/ending/main.go)がこのページに入った直後に呼び出すが、自動再生が
// ブロックされた場合のフォールバックはここで用意しておく。
setupEndingFanfareAutoplayFallback();

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
