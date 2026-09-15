const btnEnableAudio = document.getElementById('btn-enable-audio');

if (!WebAssembly) {
  console.error("このブラウザはWebAssemblyに対応していません");
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("title.wasm"), go.importObject)
    .then((result) => {
      console.log("タイトル画面: Go WASM 起動完了");
      go.run(result.instance);

      btnEnableAudio.addEventListener('click', () => {
        ensureAudioContext();
      });

      initMIDI();
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}
