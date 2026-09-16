const btnEnableAudio = document.getElementById('btn-enable-audio');

// storyEvilLaughBuffer/LoadPromiseは、このページに入った際に流すガノンの
// 高笑い(assets/audio/ganon-evil-laugh.mp3)を事前にデコードしたバッファ。
// web/grassland.js等の効果音と同じ方式(事前デコード+自己修復
// フォールバック)。
let storyEvilLaughBuffer = null;
let storyEvilLaughLoadPromise = null;
let storyEvilLaughPlayed = false;

// loadStoryEvilLaugh は、高笑いファイルの取得・デコードを開始する
// (すでに開始・完了済みなら何もしない)。デコード済みのAudioBufferで
// 解決するPromiseを返す。
function loadStoryEvilLaugh() {
  if (storyEvilLaughBuffer) {
    return Promise.resolve(storyEvilLaughBuffer);
  }
  if (storyEvilLaughLoadPromise) {
    return storyEvilLaughLoadPromise;
  }
  const ctx = ensureAudioContext();
  if (!ctx) {
    return Promise.resolve(null);
  }
  storyEvilLaughLoadPromise = fetch('assets/audio/ganon-evil-laugh.mp3')
    .then((res) => res.arrayBuffer())
    .then((data) => ctx.decodeAudioData(data))
    .then((buffer) => {
      storyEvilLaughBuffer = buffer;
      return buffer;
    })
    .catch((err) => {
      console.error('story evil laugh: failed to load/decode', err);
      storyEvilLaughLoadPromise = null; // 失敗時は次回呼び出しでもう一度試す
      return null;
    });
  return storyEvilLaughLoadPromise;
}

// playStoryEvilLaugh はGoから、このページに入った直後に呼ばれる。ブラウザの
// 自動再生ポリシーにより、ユーザー操作を経ていないAudioContextは無音のまま
// (state: 'suspended')になる場合があるため、resume()を試みてから再生する。
// それでも鳴らせなかった場合に備え、ページ内の最初のクリック/タップ/キー
// 入力で改めて再生を試みるフォールバックも用意している
// (setupStoryEvilLaughAutoplayFallback参照)。
function playStoryEvilLaugh() {
  if (storyEvilLaughPlayed) {
    return;
  }
  const ctx = ensureAudioContext();
  if (!ctx) {
    return;
  }
  Promise.all([ctx.resume().catch(() => {}), loadStoryEvilLaugh()]).then(([, buffer]) => {
    if (buffer && !storyEvilLaughPlayed && ctx.state === 'running') {
      storyEvilLaughPlayed = true;
      const source = ctx.createBufferSource();
      source.buffer = buffer;
      source.connect(ctx.destination);
      source.start(0);
    }
  });
}

// setupStoryEvilLaughAutoplayFallback は、自動再生がブロックされた場合に
// 備え、ページ内の最初のクリック/タップ/キー入力で改めて再生を試みる。
function setupStoryEvilLaughAutoplayFallback() {
  const retry = () => {
    if (storyEvilLaughPlayed) {
      return;
    }
    playStoryEvilLaugh();
  };
  ['pointerdown', 'keydown'].forEach((type) => {
    document.addEventListener(type, retry, { once: true });
  });
}

if (!WebAssembly) {
  console.error("このブラウザはWebAssemblyに対応していません");
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("story.wasm"), go.importObject)
    .then((result) => {
      console.log("ストーリー画面: Go WASM 起動完了");
      go.run(result.instance);

      btnEnableAudio.addEventListener('click', () => {
        ensureAudioContext();
      });

      // デバッグ用: MIDIキーボード/オカリナが無くても、Oキーで「ド(C)」を
      // 弾いたのと同じ効果(神殿フィールドへ遷移)を試せるようにする。
      window.addEventListener('keydown', (event) => {
        if ((event.key === 'o' || event.key === 'O') && window.goDebugTriggerTitleStart) {
          window.goDebugTriggerTitleStart();
        }
      });

      initMIDI();
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}

// 高笑いの再生自体はGo側(game.PlayStoryEvilLaugh、cmd/story/main.go)が
// このページに入った直後に呼び出すが、自動再生がブロックされた場合の
// フォールバックはここで用意しておく。
setupStoryEvilLaughAutoplayFallback();
