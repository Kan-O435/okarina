const statusEl = document.getElementById('status');

if (!WebAssembly) {
  statusEl.textContent = "このブラウザはWebAssemblyに対応していません";
  statusEl.className = "error";
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("ganon-hall.wasm"), go.importObject)
    .then((result) => {
      statusEl.textContent = "ガノンフィールド(玉座の間): Go WASM 起動完了";
      go.run(result.instance);
    })
    .catch((err) => {
      statusEl.textContent = "WASMのロードに失敗しました: " + err;
      statusEl.className = "error";
      console.error(err);
    });
}
