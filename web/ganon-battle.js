const btnEnableAudio = document.getElementById('btn-enable-audio');
const btnTestSound = document.getElementById('btn-test-sound');

// ganonBattleSongBuffer は、嵐の歌を正しく演奏した後に流す本家のBGM
// (ゼルダの伝説 時のオカリナ「嵐の歌」、無音部分を除いて5秒に切り詰め
// 済み)を事前にデコードしたバッファ。web/grassland.jsの馬のジャンプ音・
// web/ganon-hall.jsの崩落音と同じ方式(事前デコード+自己修復
// フォールバック)。
let ganonBattleSongBuffer = null;
let ganonBattleSongLoadPromise = null;

// loadGanonBattleSong は、BGMファイルの取得・デコードを開始する
// (すでに開始・完了済みなら何もしない)。デコード済みのAudioBufferで
// 解決するPromiseを返す。
function loadGanonBattleSong() {
  if (ganonBattleSongBuffer) {
    return Promise.resolve(ganonBattleSongBuffer);
  }
  if (ganonBattleSongLoadPromise) {
    return ganonBattleSongLoadPromise;
  }
  const ctx = ensureAudioContext();
  if (!ctx) {
    return Promise.resolve(null);
  }
  ganonBattleSongLoadPromise = fetch('assets/audio/ganon-battle-song-of-storms.mp3')
    .then((res) => res.arrayBuffer())
    .then((data) => ctx.decodeAudioData(data))
    .then((buffer) => {
      ganonBattleSongBuffer = buffer;
      return buffer;
    })
    .catch((err) => {
      console.error('ganon battle song: failed to load/decode', err);
      ganonBattleSongLoadPromise = null; // 失敗時は次回呼び出しでもう一度試す
      return null;
    });
  return ganonBattleSongLoadPromise;
}

// playGanonBattleSongOfStorms はGoから嵐の歌の確認音の後に呼ばれる。
// 事前にデコード済みのAudioBufferがあれば即座に再生する。「音を有効に
// する」ボタンのクリックがWASM読み込み完了前に発生した等の理由でまだ
// 読み込めていない場合は、ここで読み込みを開始し、終わり次第再生する
// (この場合に限り、初回の再生に読み込み分の遅延が乗る)。
function playGanonBattleSongOfStorms() {
  const ctx = ensureAudioContext();
  if (!ctx) {
    return;
  }
  if (ganonBattleSongBuffer) {
    playAudioBuffer(ctx, ganonBattleSongBuffer);
    return;
  }
  loadGanonBattleSong().then((buffer) => {
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
  console.error("このブラウザはWebAssemblyに対応していません");
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("ganon-battle.wasm"), go.importObject)
    .then((result) => {
      console.log("ガノンフィールド(戦場跡): Go WASM 起動完了");
      go.run(result.instance);

      btnEnableAudio.addEventListener('click', () => {
        ensureAudioContext();
        loadGanonBattleSong();
      });

      btnTestSound.addEventListener('click', () => {
        playNote(60, 100);
        setTimeout(() => stopNote(60), 500);
      });

      initMIDI();
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}
