const btnEnableAudio = document.getElementById('btn-enable-audio');
const btnTestSound = document.getElementById('btn-test-sound');

// ganonHallCollapseBuffer は、玉座の間の崩落演出(岩が降り始める瞬間)で
// 鳴らす効果音(stu9-heavy-break-363163、無音部分を除いて約1.9秒に
// 切り詰め済み)を事前にデコードしたバッファ。web/grassland.jsの
// 馬のジャンプ音と同じ方式(事前デコード+自己修復フォールバック)。
let ganonHallCollapseBuffer = null;
let ganonHallCollapseLoadPromise = null;

// loadGanonHallCollapseSound は、効果音ファイルの取得・デコードを開始する
// (すでに開始・完了済みなら何もしない)。デコード済みのAudioBufferで
// 解決するPromiseを返す。
function loadGanonHallCollapseSound() {
  if (ganonHallCollapseBuffer) {
    return Promise.resolve(ganonHallCollapseBuffer);
  }
  if (ganonHallCollapseLoadPromise) {
    return ganonHallCollapseLoadPromise;
  }
  const ctx = ensureAudioContext();
  if (!ctx) {
    return Promise.resolve(null);
  }
  ganonHallCollapseLoadPromise = fetch('assets/audio/ganon-hall-collapse.mp3')
    .then((res) => res.arrayBuffer())
    .then((data) => ctx.decodeAudioData(data))
    .then((buffer) => {
      ganonHallCollapseBuffer = buffer;
      return buffer;
    })
    .catch((err) => {
      console.error('ganon hall collapse sound: failed to load/decode', err);
      ganonHallCollapseLoadPromise = null; // 失敗時は次回呼び出しでもう一度試す
      return null;
    });
  return ganonHallCollapseLoadPromise;
}

// playGanonHallCollapseSound はGoから崩落演出の開始時に呼ばれる。事前に
// デコード済みのAudioBufferがあれば即座に再生する。「音を有効にする」
// ボタンのクリックがWASM読み込み完了前に発生した等の理由でまだ読み込めて
// いない場合は、ここで読み込みを開始し、終わり次第再生する(この場合に
// 限り、初回の再生に読み込み分の遅延が乗る)。
function playGanonHallCollapseSound() {
  const ctx = ensureAudioContext();
  if (!ctx) {
    return;
  }
  if (ganonHallCollapseBuffer) {
    playAudioBuffer(ctx, ganonHallCollapseBuffer);
    return;
  }
  loadGanonHallCollapseSound().then((buffer) => {
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
  WebAssembly.instantiateStreaming(fetch("ganon-hall.wasm"), go.importObject)
    .then((result) => {
      console.log("ガノンフィールド(玉座の間): Go WASM 起動完了");
      go.run(result.instance);

      btnEnableAudio.addEventListener('click', () => {
        ensureAudioContext();
        loadGanonHallCollapseSound();
      });

      btnTestSound.addEventListener('click', () => {
        playNote(60, 100);
        setTimeout(() => stopNote(60), 500);
      });

      // デバッグ用: 演奏無しでも、Oキーで光のプレリュードを正しく演奏した
      // のと同じ効果(確認音→崩落演出)を試せるようにする。
      window.addEventListener('keydown', (event) => {
        if ((event.key === 'o' || event.key === 'O') && window.goDebugTriggerGanonHallMelody) {
          window.goDebugTriggerGanonHallMelody();
        }
      });

      initMIDI();
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}
