const statusEl = document.getElementById('status');
const bridgeLogEl = document.getElementById('bridge-log');
const midiStatusEl = document.getElementById('midi-status');
const btnPing = document.getElementById('btn-ping');
const btnMidi = document.getElementById('btn-midi');

function appendLog(text) {
  const line = document.createElement('div');
  line.textContent = text;
  bridgeLogEl.appendChild(line);
}

function appendMidiStatus(text) {
  const line = document.createElement('div');
  line.textContent = text;
  midiStatusEl.appendChild(line);
}

function initMIDI() {
  if (!navigator.requestMIDIAccess) {
    appendMidiStatus("このブラウザはWeb MIDI APIに対応していません");
    return;
  }
  navigator.requestMIDIAccess().then(onMIDISuccess, onMIDIFailure);
}

function onMIDIFailure(err) {
  appendMidiStatus("MIDIアクセスの取得に失敗しました: " + err);
}

function onMIDISuccess(midiAccess) {
  const inputs = midiAccess.inputs;
  if (inputs.size === 0) {
    appendMidiStatus("MIDI入力デバイスが見つかりません(MPK Mini等を接続してください)");
  }
  for (const input of inputs.values()) {
    appendMidiStatus("MIDI入力デバイスを検出: " + input.name);
    input.onmidimessage = onMIDIMessage;
  }

  midiAccess.onstatechange = (event) => {
    const port = event.port;
    appendMidiStatus("MIDIデバイスの状態変化: " + port.name + " (" + port.state + ")");
    if (port.type === "input" && port.state === "connected") {
      port.onmidimessage = onMIDIMessage;
    }
  };
}

// MIDIメッセージのstatusバイト上位4bitがコマンド種別。
// 0x9(Note On, velocity>0), 0x8(Note Off) または velocity=0のNote Onのみ扱う。
function onMIDIMessage(event) {
  const [status, note, velocity] = event.data;
  const command = status >> 4;
  const isNoteOn = command === 0x9 && velocity > 0;
  const isNoteOff = command === 0x8 || (command === 0x9 && velocity === 0);

  if (!isNoteOn && !isNoteOff) {
    return;
  }

  appendMidiStatus((isNoteOn ? "Note ON  " : "Note OFF ") + "note=" + note + " velocity=" + velocity);

  if (window.goOnMIDIEvent) {
    window.goOnMIDIEvent(note, velocity, isNoteOn, event.timeStamp);
  }
}

if (!WebAssembly) {
  statusEl.textContent = "このブラウザはWebAssemblyに対応していません";
  statusEl.className = "error";
} else {
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("game.wasm"), go.importObject)
    .then((result) => {
      statusEl.textContent = "Go WASM 起動完了 (Consoleを確認してください)";
      go.run(result.instance);

      btnPing.disabled = false;
      btnMidi.disabled = false;

      btnPing.addEventListener('click', () => {
        const reply = window.goPing();
        appendLog("JS → Go → 戻り値: " + reply);
      });

      btnMidi.addEventListener('click', () => {
        window.goOnMIDIEvent(60, 100, true, performance.now());
        appendLog("JS → Go: MIDIイベント(note=60, velocity=100, on=true)を送信しました。Consoleも確認してください。");
      });

      initMIDI();
    })
    .catch((err) => {
      statusEl.textContent = "WASMのロードに失敗しました: " + err;
      statusEl.className = "error";
      console.error(err);
    });
}
