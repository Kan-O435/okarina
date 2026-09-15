const bridgeLogEl = document.getElementById('bridge-log');
const btnPing = document.getElementById('btn-ping');
const btnMidi = document.getElementById('btn-midi');
const btnOpenDoor = document.getElementById('btn-open-door');
const btnEnableAudio = document.getElementById('btn-enable-audio');
const btnTestSound = document.getElementById('btn-test-sound');

function appendLog(text) {
  const line = document.createElement('div');
  line.textContent = text;
  bridgeLogEl.appendChild(line);
}

// initMIDI/onMIDIMessage等はmidi-input.jsで定義(神殿・草原フィールド共通)。

if (!WebAssembly) {
  console.error("このブラウザはWebAssemblyに対応していません");
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("game.wasm"), go.importObject)
    .then((result) => {
      console.log("Go WASM 起動完了");
      go.run(result.instance);

      btnPing.disabled = false;
      btnMidi.disabled = false;
      btnOpenDoor.disabled = false;

      btnPing.addEventListener('click', () => {
        const reply = window.goPing();
        appendLog("JS → Go → 戻り値: " + reply);
      });

      btnMidi.addEventListener('click', () => {
        window.goOnMIDIEvent(60, 100, true, performance.now());
        appendLog("JS → Go: MIDIイベント(note=60, velocity=100, on=true)を送信しました。Consoleも確認してください。");
      });

      btnOpenDoor.addEventListener('click', () => {
        window.goOpenDoor();
        appendLog("JS → Go: goOpenDoor() を呼びました。隠し扉が開くはずです。");
      });

      btnEnableAudio.addEventListener('click', () => {
        ensureAudioContext();
        appendLog("オーディオを有効化しました");
      });

      btnTestSound.addEventListener('click', () => {
        playNote(60, 100);
        setTimeout(() => stopNote(60), 500);
        appendLog("テスト再生: note=60 を0.5秒鳴らしました");
      });

      initMIDI();
    })
    .catch((err) => {
      console.error("WASMのロードに失敗しました:", err);
    });
}
