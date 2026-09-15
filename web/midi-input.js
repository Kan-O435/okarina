// Web MIDI APIからのノート入力を扱う共通ロジック。神殿フィールド(main.js)・
// 草原フィールド(grassland.js)の両方から使う。
// MIDIノートが来たら、その場でオカリナの音を鳴らし(audio.jsのplayNote/
// stopNote)、Go/WASM側(旋律認識)にも転送する(window.goOnMIDIEvent)。

const midiStatusEl = document.getElementById('midi-status');

function appendMidiStatus(text) {
  if (!midiStatusEl) {
    return;
  }
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

  if (isNoteOn) {
    playNote(note, velocity);
  } else {
    stopNote(note);
  }

  if (window.goOnMIDIEvent) {
    window.goOnMIDIEvent(note, velocity, isNoteOn, event.timeStamp);
  }
}
