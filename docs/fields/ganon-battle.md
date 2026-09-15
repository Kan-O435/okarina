# フィールド情報: ガノンフィールド(戦場跡案)— 最終形態撃破カットシーン

## 概要

`cmd/ganon-battle`(戦場跡案、`internal/renderer.GanonBackgroundBattle`)フィールドに、
Ganon最終形態(`assets.GanonBossBattle`、緑色・角のある獣形態)を倒すカットシーンを実装した。

演出の流れ:

```
(トリガー: デバッグボタン、または「嵐の歌」を演奏)
      ↓
嵐が発生する(嵐雲が湧き上がり、空が暗くなる)
      ↓
ガノンに雷が落ちる(画面全体+ガノン直撃、複数回明滅)
      ↓
ガノンが爆発して消える(煙+岩片+閃光、本体を実際に画面外へ隠す)
      ↓
画面が白くフェードする
      ↓
ending.html(新規、静的ページ)へページ遷移
```

この文書は、上記カットシーンの実装のためにGo側(`internal/renderer`, `internal/game`,
`internal/bridge`, `cmd/ganon-battle`)に加えた変更をまとめたもの。メロディ判定
(「嵐の歌」、`music.GanonBattleMelodyName`)と本家BGM再生は別セッションの実装
(`feature/ganon-battle-melody`ブランチ、PR #62)によるもので、本ドキュメントの主眼は
撃破演出そのもの。

## エンジン側の拡張(`internal/renderer`)

雷の閃光・最後の白フェードには、通常の3Dオブジェクト(`Object`、`Color`はvec3のみで
アルファの概念が無い)では表現できない「画面全体を任意の不透明度で覆う」処理が必要だった。
そのため以下を追加している:

- **`shaders.go`**: 共通フラグメントシェーダーに`uniform float uAlpha`を追加し、
  `gl_FragColor`のアルファに掛け合わせるようにした。WebGLのuniform初期値は0.0のため、
  既存の`Scene.Render`/`HUD.Render`では明示的に`uAlpha = 1.0`を設定するよう変更している
  (これを怠ると全描画が透明になる)。
- **`shader_js.go` / `shader_stub.go`**: `Program.SetUniformVec3`と同じ形で
  `SetUniformFloat`を追加(js版は`uniform1f`呼び出し、ネイティブビルド向けstubはno-op)。
- **`overlay.go`(新規)**: `WhiteFadeOverlay`型。`HUD`とほぼ同じ構造(スクリーン座標の
  クアッド+オルソ投影)だが、バナー画像ではなくキャンバス全体を覆う白いクアッドを、
  外部から指定した可変アルファで毎フレーム描画し直せる点が異なる。雷の閃光にも最後の
  白フェードにも同じ型を使い回している。

## ganon-battle固有の演出パーツ(`internal/renderer/ganon.go`)

### ボスを直接隠す仕組み

`ganon-hall`(玉座の間)の第一形態撃破演出はボスの3Dオブジェクトを非表示にする仕組みを
持たず、煙玉でカメラから隠すだけで成立していた(カメラが固定のため)。今回はユーザーの
要望で「実際に消える」ようにする必要があったため、以下を追加した:

- **`buildGanonSceneObjects`**: `BuildGanonScene`の中身を共通化した非公開ヘルパー。
  地面・背景・ボス・Linkを積み上げつつ、ボスが`scene.Objects`内で占める範囲
  (`bossStart`, `bossEnd`)も計算する。
- **`BuildGanonSceneWithBossRange`(新規、公開)**: `BuildGanonScene`と同じくSceneを
  組み立てるが、追加でボスのインデックス範囲を返す。既存の`BuildGanonScene`(戦場跡・
  玉座の間の両方から呼ばれる)には一切手を加えず、新しい関数を追加するだけに留めている。
  `cmd/ganon-battle/main.go`は爆発フェーズの冒頭でこの範囲のオブジェクトを
  `Translate(0, -500, 0)`で画面外へ退避させ、「爆発して消える」を実際に消すことで
  表現している。

### 嵐雲

- **`ganonBattleStormCloudColor`**: 元は明るい水色のアセットを暗い青灰色に着色。
- **`ganonBattleStormCloudPlacements`**(10個): `assets.GanonStormCloud`(7種類の
  フラットカラー低ポリ雲)から選び、X: -30〜30・Z: -30〜-62に散らして上空全体を覆う。
  戦場跡カメラの画角(垂直35.3°、水平視線)は狭いため、Yをそのまま高くすると雲が
  フレーム上端よりさらに上に出て一切映らなくなる不具合があった。各雲のYは、
  そのZ位置での画面上端のワールドY(`ganonBattleCameraY + 距離*tan(35.3°/2)`)から
  雲の高さの半分を引いた値にすることで、雲の下半分だけが画面に映り上半分は上端で
  切れる「垂れ込めた雲」の見た目になるよう修正した。
- **`GanonBattleStormCloudObjects`**: `gltf.ParseParts`+`Model.GroundTransform`で
  読み込み・配置する(`demo.go`の`templeCloudObjects`と同じパターン)。

### 雷

- **`GanonBattleLightningObject`**: `assets.GanonLightning`(3種類の分岐雷)から
  最も縦長のパーツを選び、Ganonの頭上(`GanonBattleLightningStrikeY`、
  `GanonBattleBossZ`)に着弾させる1本。
- **`ganonBattleSkyLightningPlacements`**(7本): 着弾用の1本だけでは「背景の後ろだけ
  光っている」ように見えたため、3段階の奥行きに追加で配置している:
  1. 遠景(Z: -55〜-60、空いっぱいに伸びる)
  2. フィールド中景(Z: -28〜-33、手前の廃墟の中に落ちる)
  3. Ganon直撃(Z: -37〜-39、着弾用の1本の左右にもう2本)
- **`GanonBattleSkyLightningObjects`**: 上記をまとめて配置する。

### 爆発

- **`GanonBattleExplosionSmokePlacements`**(8個): `GanonHallSmokePuffPlacement`
  (玉座の間の第一形態撃破演出と同じ構造体)を流用し、`GanonHallSmokeObject`
  (煙玉のテンプレート)と組み合わせる。玉座の間より大きめ・多めにしてある。
- **`GanonBattleExplosionDebrisPlacements`**(8個): `GanonHallCollapseRockPlacement`/
  `GanonHallCollapseRockObject`(玉座の間の崩落演出と同じ)を流用。X/Zは各岩片が
  外側へ飛んでいく先を表し、実際の「中心から外側へ飛んでから落ちる」放物線の動きは
  `cmd/ganon-battle/main.go`側で計算する。

## トリガー・ブリッジ配線(`internal/game/game.go`, `internal/bridge/bridge_js.go`)

`ganonHallCollapseTrigger`(玉座の間)と全く同じコールバックパターンで追加:

- `game.SetGanonBattleDefeatTrigger(f func())` / `game.TriggerGanonBattleDefeat()`:
  `cmd/ganon-battle/main.go`がフェーズ管理の開始関数を登録し、ブリッジ
  (`goDefeatGanonFinalForm`)またはメロディ認識(`onMelodyRecorded`、別セッション実装)
  から呼ばれる。
- `game.SetPlayGanonBattleThunderSoundFunc` / `game.PlayGanonBattleThunderSound()`:
  雷が落ちた瞬間の雷鳴用フック。
- `bridge_js.go`の`goDefeatGanonFinalForm`: デバッグボタンから`TriggerGanonBattleDefeat()`
  を呼ぶだけの薄いラッパー(`goDefeatGanonFirstForm`と同じ仮実装パターン)。

## カットシーンのフェーズ管理(`cmd/ganon-battle/main.go`)

`cmd/ganon-hall/main.go`の`ganonHallPhase`と同じ考え方で、`ganonBattleDefeatPhase`
というステートマシンを`ctx.RunLoop`内に持たせている。

| フェーズ | 長さ | 内容 |
|---|---|---|
| `ganonBattleDefeatIdle` | - | 通常操作(トリガー待ち) |
| `ganonBattleDefeatStorming` | 2.5秒 | 嵐雲をスケール0→1でsmoothstep湧き上がらせ、`ctx.ClearColor`を通常色→暗い嵐色へ線形補間 |
| `ganonBattleDefeatLightning` | 0.9秒 | 着弾用+画面全体の雷(計8本)を出現させ、`ganonBattleLightningFlickers`で定義した3回の三角形状の閃光を白オーバーレイで明滅させる。雷鳴(`PlayGanonBattleThunderSound`)もここで再生 |
| `ganonBattleDefeatExploding` | 1.4秒 | フェーズ冒頭で`BuildGanonSceneWithBossRange`のボス範囲を画面外へ退避(閃光の裏に隠す)。同時に大きな閃光(`ganonBattleExplosionFlashFraction`)、煙玉8個・岩片8個のアニメーションを開始 |
| `ganonBattleDefeatFading` | 1.2秒 | 白オーバーレイのalphaを0→1、到達したら`ctx.Navigate("ending.html")` |

ページ読み込み直後のwipe岩(ganon-hallからの遷移演出)は上記ステートマシンと独立して
動作し続ける(既存実装のまま変更していない)。

## `web/ending.html`(新規)

静的ページ(WASM/MIDI/canvas無し)。`style.css`のダークテーマを流用し、中央に
締めのメッセージをフェードインで表示するだけの最小限の実装。

## `web/ganon-battle.html` / `web/ganon-battle.js`

- デバッグボタン`<button id="btn-defeat-ganon-final">ガノン(最終形態)を倒した(テスト)</button>`
  を追加し、クリックで`window.goDefeatGanonFinalForm()`を呼ぶ。
- `playGanonBattleThunderSound()`: 雷鳴効果音。新規の音声ファイルは追加せず、
  `audio.js`が既に持っているホワイトノイズバッファ(`noiseBuffer`)を使い回し、
  ローパスフィルタ(1800Hz→120Hzへの急速な減衰)とゲインエンベロープでWeb Audio APIに
  よりその場で合成している。

## アセット取得ログ

| # | パーツ | 取得元 | 形式/サイズ | 補足 |
|---|---|---|---|---|
| 1 | 雷(`GanonLightning`) | Sketchfab「3 Pack of Storm Lightning」(作者: Incg5764、CC Attribution) | GLB, 42kB(3パーツ、分岐した低ポリ雷) | `docs/licenses/sketchfab-incg5764-stormlightning-CC-BY.txt` |
| 2 | 嵐雲(`GanonStormCloud`) | Sketchfab「Low Poly Cloud 3D Model Game Asset UI Element」(作者: B1Blender、CC Attribution) | GLB, 171kB(7パーツ、フラットカラー低ポリ雲) | `docs/licenses/sketchfab-b1blender-lowpolycloud-CC-BY.txt` |

保存先: `internal/assets/models/ganon-lightning.glb`, `internal/assets/models/storm-cloud.glb`
(いずれも`go:embed`で`internal/assets.GanonLightning`/`GanonStormCloud`として埋め込み)。

## 関連

- 玉座の間の第一形態撃破演出(`cmd/ganon-hall/main.go`)が今回の実装の直接のテンプレート。
  煙玉・岩片・wipe岩のオブジェクト/localTransformの仕組みはすべてそちらを再利用している。
- メロディ判定(「嵐の歌」レ→ファ→レ→レ→ファ→レ、`music.GanonBattleMelodyName`)・
  確認音・本家BGM(`ganon-battle-song-of-storms.mp3`)・嵐の歌の楽譜HUDは、
  `feature/ganon-battle-melody`ブランチ(PR #62)で別途実装されたもので、
  `game.TriggerGanonBattleDefeat()`を呼ぶことでこのカットシーンに接続している。
