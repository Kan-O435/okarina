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

## 今後の予定

- [ ] #2: GoをWASMとして実行できるようにする
- [ ] #3: GoとJavaScript間のBridgeを作成する(MIDIイベントをGoへ渡す)
