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

      // 楽譜HUD(WebGL側)に添えた煽り文は、楽譜HUD本体と全く同じ
      // タイミング(game.SetMelodyHUDVisibleFuncで登録された条件)で
      // 表示・非表示を切り替える。
      const melodyHintEl = document.getElementById('ganon-battle-melody-hint');
      function updateMelodyHintVisibility() {
        melodyHintEl.hidden = !(window.goMelodyHUDVisible && window.goMelodyHUDVisible());
        requestAnimationFrame(updateMelodyHintVisibility);
      }
      requestAnimationFrame(updateMelodyHintVisibility);

      // 「特定の演奏でGanonの最終形態を倒す」処理はまだ無いため、動作確認用の
      // ボタンから直接goDefeatGanonFinalForm()を呼ぶ仮実装
      // (web/ganon-hall.htmlのbtn-defeat-ganonと同様)。
      document.getElementById('btn-defeat-ganon-final').addEventListener('click', () => {
        window.goDefeatGanonFinalForm();
      });
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}

// playGanonBattleThunderSound は、Ganon最終形態撃破演出で雷が落ちた瞬間に
// Go側(internal/bridge/bridge_js.goのcallPlayGanonBattleThunderSound)から
// 呼ばれる雷鳴の効果音。新しい音声ファイルは追加せず、audio.jsが
// ensureAudioContext()内で作っているホワイトノイズ(noiseBuffer)を
// 使い回して合成する。ローパスフィルタで周波数を急激に下げていく短い
// バーストと、それに続く長めの減衰でざっくり「バリバリ…ゴロゴロ」という
// 雷鳴らしさを出している(オカリナの息ノイズほど作り込む必要はないため
// シンプルな2段構成に留めている)。
function playGanonBattleThunderSound() {
  const ctx = ensureAudioContext();
  if (!ctx || !noiseBuffer) {
    return;
  }
  const now = ctx.currentTime;

  const source = ctx.createBufferSource();
  source.buffer = noiseBuffer;
  source.loop = true;

  const filter = ctx.createBiquadFilter();
  filter.type = 'lowpass';
  filter.frequency.setValueAtTime(1800, now);
  filter.frequency.exponentialRampToValueAtTime(120, now + 1.2);

  const gain = ctx.createGain();
  gain.gain.setValueAtTime(0.0001, now);
  gain.gain.exponentialRampToValueAtTime(0.9, now + 0.03);
  gain.gain.exponentialRampToValueAtTime(0.0001, now + 1.4);

  source.connect(filter);
  filter.connect(gain);
  gain.connect(ctx.destination);

  source.start(now);
  source.stop(now + 1.5);
}
