const statusEl = document.getElementById('status');
const btnEnableAudio = document.getElementById('btn-enable-audio');

if (!WebAssembly) {
  statusEl.textContent = "このブラウザはWebAssemblyに対応していません";
  statusEl.className = "error";
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("title.wasm"), go.importObject)
    .then((result) => {
      statusEl.textContent = "タイトル画面: Go WASM 起動完了";
      go.run(result.instance);

      btnEnableAudio.addEventListener('click', () => {
        ensureAudioContext();
      });

      initMIDI();
    })
    .catch((err) => {
      statusEl.textContent = "WASMのロードに失敗しました: " + err;
      statusEl.className = "error";
      console.error(err);
    });
}
