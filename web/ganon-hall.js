const btnEnableAudio = document.getElementById('btn-enable-audio');
const btnTestSound = document.getElementById('btn-test-sound');

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
