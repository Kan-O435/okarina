const statusEl = document.getElementById('status');
const btnEnableAudio = document.getElementById('btn-enable-audio');
const btnTestSound = document.getElementById('btn-test-sound');

// horseJumpBuffer は、馬がジャンプした際にいななく効果音(mixkit-scared-
// horse-neighing-85、無音部分を除いて1.9秒に切り詰め済み)を事前に
// デコードしたバッファ。ジャンプが起きてからfetch/decodeAudioDataすると
// その分再生が遅れてしまうため、ensureAudioContext()と同じタイミング
// (「音を有効にする」ボタン)であらかじめ読み込み・デコードしておく。
// horseJumpLoadPromiseは、読み込み中(あるいは読み込み済み)のPromiseを
// 保持し、playHorseJumpSound側からも同じ読み込み処理を使い回せるように
// するためのもの。
let horseJumpBuffer = null;
let horseJumpLoadPromise = null;

// loadHorseJumpSound は、効果音ファイルの取得・デコードを開始する
// (すでに開始・完了済みなら何もしない)。デコード済みのAudioBufferで
// 解決するPromiseを返す。
function loadHorseJumpSound() {
  if (horseJumpBuffer) {
    return Promise.resolve(horseJumpBuffer);
  }
  if (horseJumpLoadPromise) {
    return horseJumpLoadPromise;
  }
  const ctx = ensureAudioContext();
  if (!ctx) {
    return Promise.resolve(null);
  }
  horseJumpLoadPromise = fetch('assets/audio/horse-jump-neigh.mp3')
    .then((res) => res.arrayBuffer())
    .then((data) => ctx.decodeAudioData(data))
    .then((buffer) => {
      horseJumpBuffer = buffer;
      return buffer;
    })
    .catch((err) => {
      console.error('horse jump sound: failed to load/decode', err);
      horseJumpLoadPromise = null; // 失敗時は次回呼び出しでもう一度試す
      return null;
    });
  return horseJumpLoadPromise;
}

// playHorseJumpSound はGoから馬のジャンプ時に呼ばれる。事前にデコード
// 済みのAudioBufferがあればAudioBufferSourceNodeで即座に再生する
// (<audio>要素のplay()を都度呼ぶ場合に比べて再生開始の遅延がほぼ無い)。
// 「音を有効にする」ボタンのクリックがWASM読み込み完了前に発生した等の
// 理由でまだ読み込めていない場合は、ここで読み込みを開始し、終わり次第
// 再生する(この場合に限り、初回の再生に読み込み分の遅延が乗る)。
function playHorseJumpSound() {
  const ctx = ensureAudioContext();
  if (!ctx) {
    return;
  }
  if (horseJumpBuffer) {
    playAudioBuffer(ctx, horseJumpBuffer);
    return;
  }
  loadHorseJumpSound().then((buffer) => {
    if (buffer) {
      playAudioBuffer(ctx, buffer);
    }
  });
}

function playAudioBuffer(ctx, buffer) {
  const source = ctx.createBufferSource();
  source.buffer = buffer;
  source.connect(ctx.destination);
  source.start(0);
}

if (!WebAssembly) {
  statusEl.textContent = "このブラウザはWebAssemblyに対応していません";
  statusEl.className = "error";
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("grassland.wasm"), go.importObject)
    .then((result) => {
      statusEl.textContent = "草原フィールド: Go WASM 起動完了";
      go.run(result.instance);

      btnEnableAudio.addEventListener('click', () => {
        ensureAudioContext();
        loadHorseJumpSound();
      });

      btnTestSound.addEventListener('click', () => {
        playNote(60, 100);
        setTimeout(() => stopNote(60), 500);
      });

      initMIDI();
    })
    .catch((err) => {
      statusEl.textContent = "WASMのロードに失敗しました: " + err;
      statusEl.className = "error";
      console.error(err);
    });
}
