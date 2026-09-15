if (!WebAssembly) {
  console.error("このブラウザはWebAssemblyに対応していません");
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("ending.wasm"), go.importObject)
    .then((result) => {
      console.log("エンディング画面: Go WASM 起動完了");
      go.run(result.instance);
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}
