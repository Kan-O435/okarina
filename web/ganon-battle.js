const btnEnableAudio = document.getElementById('btn-enable-audio');
const btnTestSound = document.getElementById('btn-test-sound');

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
      });

      btnTestSound.addEventListener('click', () => {
        playNote(60, 100);
        setTimeout(() => stopNote(60), 500);
      });

      initMIDI();

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
