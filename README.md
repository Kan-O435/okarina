# okarina

ハックツハッカソン参加プロジェクト。
MIDIキーボード(Akai MPK Mini MK3)で演奏したメロディに応じて、
ゲーム世界の状態(昼夜・天候・ギミックなど)が変化する
「時のオカリナ」リメイク企画。

## プロジェクト構成

```
/
├── cmd/
│   └── game/        # エントリーポイント (main.go)
├── internal/
│   ├── game/         # ゲーム全体のループ・状態管理
│   ├── midi/          # MIDI入力の受信・メロディ判定
│   ├── music/         # メロディ(合言葉曲)の定義・照合ロジック
│   ├── world/          # フィールド状態(昼夜・天候等)の管理
│   └── renderer/        # 描画処理
├── web/                  # ブラウザ向け静的ファイル(HTML/JS, wasm_exec.js等)
├── assets/               # 画像・音声等のアセット
├── go.mod
└── README.md
```

## 動作確認 (ネイティブ実行)

```bash
go run ./cmd/game
```

`Game initialized` と `game.Run() called` がコンソールに表示されれば正常。

## 動作確認 (ブラウザ / WASM)

### 1. WASMビルド

```bash
GOOS=js GOARCH=wasm go build -o web/game.wasm ./cmd/game
```

### 2. wasm_exec.js の配置(Goのバージョンを上げた時などは再実行)

Go 1.24以降は配置場所が `lib/wasm/` に変わっている(それ以前は `misc/wasm/`)。

```bash
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js
```

### 3. ローカルサーバーで配信

`fetch()` で `.wasm` を読み込むため、`file://` では動かない。
`web/` 直下でHTTPサーバーを立てる。

```bash
cd web
python3 -m http.server 8000
```

### 4. ブラウザで確認

`http://localhost:8000/index.html` を開き、DevToolsのConsoleに

```
Game initialized
game.Run() called
```

が表示されればOK。

### 5. JS-Go Bridgeの確認

画面上のボタンを操作すると、JavaScript ⇔ Go WASM の相互呼び出しを確認できる。

- 「JS→Go: goPing() を呼ぶ」: JSからGoの関数を呼び出し、戻り値を受け取る
- 「JS→Go: MIDIイベントを送る(テスト)」: MIDIイベント相当のデータ(note, velocity, isNoteOn, timestamp)をGoへ渡す

いずれもConsoleと画面下部のログにメッセージが表示されればOK。

## 今後の予定

- [x] Goプロジェクトを作成する
- [x] GoをWASMとして実行できるようにする
- [x] GoとJavaScript間のBridgeを作成する
- [ ] MIDIデバイスを検出する
- [ ] MIDI Note ON/OFFを取得する
- [ ] MIDIイベントをGoへ渡す(Web MIDI APIとの結合)
