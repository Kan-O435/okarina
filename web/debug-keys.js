// デバッグ用: オタマトーン/マイクが無くても、矢印キー(←/→)でLinkを
// 前後に動かして動作確認できるようにする。本番の入力経路(ピッチ検出)とは
// 独立しており、キーを離すとIdleに戻る。
// → : 前進, ← : 後退

let debugLeftPressed = false;
let debugRightPressed = false;

function updateDebugDirection() {
  if (!window.goSetDebugDirection) {
    return;
  }
  let direction = 'idle';
  if (debugRightPressed && !debugLeftPressed) {
    direction = 'forward';
  } else if (debugLeftPressed && !debugRightPressed) {
    direction = 'backward';
  }
  window.goSetDebugDirection(direction);
}

window.addEventListener('keydown', (event) => {
  if (event.key === 'ArrowRight') {
    debugRightPressed = true;
    updateDebugDirection();
  } else if (event.key === 'ArrowLeft') {
    debugLeftPressed = true;
    updateDebugDirection();
  }
});

window.addEventListener('keyup', (event) => {
  if (event.key === 'ArrowRight') {
    debugRightPressed = false;
    updateDebugDirection();
  } else if (event.key === 'ArrowLeft') {
    debugLeftPressed = false;
    updateDebugDirection();
  }
});

// ウィンドウがフォーカスを失うとkeyupが飛んで来ず押しっぱなし扱いのまま
// 残ることがあるため、フォーカスが外れたら念のため止める。
window.addEventListener('blur', () => {
  debugLeftPressed = false;
  debugRightPressed = false;
  updateDebugDirection();
});
