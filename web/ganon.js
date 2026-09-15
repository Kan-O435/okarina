const statusEl = document.getElementById('status');
const btnVariantBattle = document.getElementById('btn-variant-battle');
const btnVariantHall = document.getElementById('btn-variant-hall');
const variantStatusEl = document.getElementById('ganon-variant-status');

const VARIANT_LABELS = {
  battle: '現行案: 荒れ果てた戦場跡のジオラマを表示中',
  hall: '旧案: 玉座の間を表示中',
};

function setActiveVariantButton(variant) {
  btnVariantBattle.classList.toggle('variant-active', variant === 'battle');
  btnVariantHall.classList.toggle('variant-active', variant === 'hall');
  variantStatusEl.textContent = VARIANT_LABELS[variant] || '';
}

btnVariantBattle.addEventListener('click', () => {
  if (window.goSwitchGanonBackground) {
    window.goSwitchGanonBackground('battle');
    setActiveVariantButton('battle');
  }
});

btnVariantHall.addEventListener('click', () => {
  if (window.goSwitchGanonBackground) {
    window.goSwitchGanonBackground('hall');
    setActiveVariantButton('hall');
  }
});

if (!WebAssembly) {
  statusEl.textContent = "このブラウザはWebAssemblyに対応していません";
  statusEl.className = "error";
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("ganon.wasm"), go.importObject)
    .then((result) => {
      statusEl.textContent = "ガノンフィールド: Go WASM 起動完了";
      go.run(result.instance);
    })
    .catch((err) => {
      statusEl.textContent = "WASMのロードに失敗しました: " + err;
      statusEl.className = "error";
      console.error(err);
    });
}
