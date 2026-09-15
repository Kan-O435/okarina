# AGENTS.md

## Project Overview

本プロジェクトは、**MIDI楽器の演奏を入力として、3Dゲーム世界に変化を起こすブラウザゲーム**である。

コンセプトは「ゼルダの伝説 時のオカリナ」のような、**音楽によって世界に変化が起こる3Dゲーム体験**から着想を得ている。

ただし、本プロジェクトはゼルダの完全な再現・クローンを目的としない。

MVPでは、以下の一連の体験を成立させることを最重要目標とする。

```text
MIDI楽器
    ↓
MIDI Note Input
    ↓
メロディ認識
    ↓
ゲームイベント
    ↓
自作3Dゲームエンジン
    ↓
3D世界の変化
```

---

# 1. MVPの目的

MVPで最も重要なのは、**MIDI演奏によって3D世界が変化すること**である。

以下の3つのギミックのみを実装する。

1. 天候変化
2. 昼夜変化
3. 隠し扉の開閉

### MIDIとギミックの対応

```text
Melody A
    ↓
Weather Change

Melody B
    ↓
Day / Night Change

Melody C
    ↓
Secret Door Open
```

この3つがMIDI演奏によって正常に動作すれば、MVPのコアは完成とする。

---

# 2. MVPの範囲

## 実装するもの

* MIDIデバイス接続
* MIDI Note On / Note Off取得
* MIDIノート列の取得
* 定義済みメロディの認識
* 3Dフィールド
* プレイヤー
* プレイヤー移動
* 三人称カメラ
* 3Dモデル表示
* 基本的なアニメーション
* 簡易衝突判定
* 天候変更
* 昼夜変更
* 隠し扉
* 扉の開閉アニメーション
* 簡易パーティクル
* ゲームイベントシステム

---

# 3. MVPで実装しないもの

以下はMVPの対象外とする。

* 敵
* 戦闘
* ボス
* NPC
* アイテム
* インベントリ
* ダンジョン
* 大規模マップ
* クエスト
* セーブ機能
* マルチプレイ
* ログイン
* ユーザー登録
* ランキング
* ショップ
* 複雑な物理演算
* 高度なAI
* 完全なPBR
* 高度なポストプロセス
* 完全なゲームエディタ
* 完全な3Dエンジン
* Three.js相当の汎用ライブラリ

時間に余裕がある場合のみ拡張する。

---

# 4. 技術方針

## 基本方針

**3Dエンジンを自作する。**

Three.jsなどの既存3Dエンジン・3Dレンダリングライブラリには依存しない。

ただし、ブラウザが提供するWebGL APIは利用する。

---

# 5. 技術スタック

| 項目        | 技術           |
| --------- | ------------ |
| ゲームエンジン   | 自作           |
| メイン言語     | Go           |
| 実行環境      | WebAssembly  |
| 3D API    | WebGL        |
| シェーダー     | GLSL         |
| ブラウザ側     | TypeScript   |
| MIDI      | Web MIDI API |
| 3Dモデル     | GLB / glTF   |
| ゲームループ    | Go           |
| Animation | 自作           |
| Collision | 自作・簡易実装      |
| Particle  | 自作・簡易実装      |
| DB        | MVPでは使用しない   |

---

# 6. アーキテクチャ

全体構成は以下とする。

```text
Browser
│
├── TypeScript
│   ├── Web MIDI API
│   ├── Canvas
│   ├── Keyboard Input
│   ├── WASM Boot
│   └── Browser API Bridge
│
└── Go / WebAssembly
    │
    ├── Game Runtime
    │   ├── Game Loop
    │   ├── Scene
    │   ├── Entity
    │   ├── Transform
    │   ├── Camera
    │   ├── Input
    │   └── Game State
    │
    ├── 3D Engine
    │   ├── Math
    │   ├── Renderer
    │   ├── Mesh
    │   ├── Material
    │   ├── Texture
    │   ├── Animation
    │   └── Lighting
    │
    ├── Asset System
    │   └── GLB / glTF
    │
    └── Game Logic
        ├── MIDI
        ├── Melody Recognition
        ├── Weather
        ├── Day/Night
        └── Secret Door

WebGL
 ↓
GPU
 ↓
Canvas
```

---

# 7. Goの責務

Goはゲームと3Dエンジンの中心として使用する。

## 3Dエンジン

* Vec2
* Vec3
* Vec4
* Matrix4
* Quaternion
* Scene
* Entity
* Transform
* Camera
* Mesh
* Material
* Texture
* Renderer
* Shader
* Animation
* Skeleton
* Bone
* Particle
* Collision

## ゲーム

* Game Loop
* Player
* Player Movement
* Game State
* Weather
* Day/Night
* Secret Door
* Event System
* Melody Recognition

---

# 8. TypeScriptの責務

TypeScriptは**ブラウザ固有の処理に限定する**。

主な責務：

* Web MIDI API
* キーボード入力
* Canvas管理
* WASMのロード
* ブラウザイベント
* Go/WASMとのBridge

ゲームロジックをTypeScript側に分散させない。

---

# 9. MIDIシステム

MIDI入力はWeb MIDI APIを利用する。

```text
MIDI Device
    ↓
Web MIDI API
    ↓
TypeScript
    ↓
Go/WASM
    ↓
MIDI Manager
```

取得する情報：

* Note Number
* Velocity
* Note On
* Note Off
* Timestamp

MIDI音声をマイクから解析する方式は採用しない。

**MIDI信号そのものを利用する。**

---

# 10. Melody Recognition

あらかじめ定義したメロディと、ユーザーが演奏したMIDIノート列を比較する。

例：

```text
Melody A

C → D → E → G

60 → 62 → 64 → 67
```

正しいメロディが入力された場合、

```text
MelodyRecognized
```

イベントを発生させる。

メロディ認識と3D描画処理は直接結合しない。

---

# 11. Event System

システム間の通信にはEvent Systemを使用する。

例：

```text
MelodyRecognized
        ↓
Game Event
        ↓
┌───────┼────────┐
↓       ↓        ↓
Weather Time  SecretDoor
```

イベント例：

```text
WeatherChange
DayNightChange
SecretDoorOpen
```

これにより、

```text
MIDI
```

と

```text
Renderer
```

を直接依存させない。

---

# 12. 3D Engine

自作3Dエンジンは、今回のゲームに必要な機能だけ実装する。

## Math

```text
Vec2
Vec3
Vec4
Matrix4
Quaternion
```

## Scene

```text
Scene
 ├── Entity
 ├── Transform
 └── Camera
```

## Renderer

WebGLを利用する。

```text
Go
 ↓
WebAssembly
 ↓
WebGL
 ↓
GLSL
 ↓
GPU
```

---

# 13. Renderer

最低限の描画機能を実装する。

必要な機能：

* WebGL初期化
* Shader生成
* Vertex Buffer
* Index Buffer
* Texture
* Material
* Depth Test
* Camera
* Model Matrix
* View Matrix
* Projection Matrix
* Draw Call

MVPでは高度なレンダリング機能は不要。

---

# 14. GLB / glTF

3Dモデルは基本的にGLB形式を使用する。

```text
GLB
├── Mesh
├── Material
├── Texture
├── Skeleton
└── Animation
```

自作Asset Loaderで読み込む。

ただし、glTF仕様を完全実装する必要はない。

**今回使用するアセットに必要な仕様だけを実装する。**

---

# 15. Animation

Animation Systemも自作する。

```text
Animation
├── AnimationClip
├── Animator
├── Skeleton
├── Bone
└── Keyframe
```

例えばPlayer：

```text
Idle
Walk
Ocarina
```

MIDIによるメロディ認識後、

```text
Melody Recognized
        ↓
Animator.Play("Ocarina")
```

などのアニメーションを実行する。

---

# 16. Player

MVPでは最低限のプレイヤー操作を実装する。

必要機能：

* 移動
* 回転
* 重力
* 地面との衝突
* カメラ追従

高度なアクションは不要。

---

# 17. Camera

三人称カメラを使用する。

```text
Player
   ↑
Camera
   ↓
World
```

Cameraが管理するもの：

* Position
* Rotation
* FOV
* Near
* Far
* Aspect Ratio

---

# 18. Collision

物理エンジンは作らない。

MVPでは簡易Collisionを実装する。

候補：

* AABB
* Sphere
* Raycast

必要な衝突だけ実装する。

主な対象：

```text
Player
Ground
Wall
Door
```

---

# 19. Weather System

MVPでは2種類のみ。

```text
Sunny
Rainy
```

雨は簡易Particle Systemで表現する。

```text
Rain
 ↓
Particle System
 ↓
Rain Particles
```

---

# 20. Day/Night System

MVPでは2状態のみ。

```text
Day
Night
```

状態変更に応じて、

* Lighting
* Environment
* Sky
* World Color

などを変更する。

---

# 21. Secret Door

隠し扉を3D Entityとして実装する。

状態：

```text
Closed
Opening
Opened
```

メロディ認識成功時：

```text
Melody C
 ↓
SecretDoorOpen
 ↓
Animator.Play("Open")
 ↓
Opened
```

---

# 22. Particle System

MVPでは汎用的なParticle Engineを作る必要はない。

最低限、

```text
Particle
├── Position
├── Velocity
├── Lifetime
└── Size
```

を実装する。

主な用途は雨。

---

# 23. Game Loop

ゲームループはGo側で管理する。

```text
Game Loop

Input
 ↓
Update
 ↓
Collision
 ↓
Game Logic
 ↓
Animation
 ↓
Render
 ↓
Next Frame
```

基本的な処理単位は、

```text
Update(deltaTime)
Render()
```

とする。

---

# 24. Runtime State

MVPではDBを使用しない。

以下はメモリ上のRuntime Stateとして管理する。

```text
Weather
DayNight
SecretDoor
Player Position
Player Rotation
Current Animation
MIDI Input
Melody Recognition
```

---

# 25. DB

MVPではDB・バックエンドを導入しない。

メロディについても、最初はGo内の定義またはJSON等の静的データで管理してよい。

例：

```text
Melody A
60, 62, 64, 67
→ Weather Change
```

DBが必要になった場合のみ後から追加する。

---

# 26. 推奨ディレクトリ構成

```text
project/
│
├── cmd/
│   └── wasm/
│       └── main.go
│
├── engine/
│   ├── math/
│   ├── renderer/
│   ├── scene/
│   ├── camera/
│   ├── mesh/
│   ├── material/
│   ├── texture/
│   ├── shader/
│   ├── animation/
│   ├── collision/
│   ├── particle/
│   └── asset/
│
├── game/
│   ├── player/
│   ├── weather/
│   ├── daynight/
│   ├── secretdoor/
│   ├── midi/
│   ├── melody/
│   └── event/
│
├── assets/
│   ├── models/
│   ├── textures/
│   ├── shaders/
│   └── audio/
│
└── web/
    ├── index.html
    ├── main.ts
    └── midi.ts
```

---

# 27. 実装方針

## 最重要

**3Dエンジンを完成させてからゲームを作る、という開発方法は禁止する。**

以下のように進める。

```text
ゲーム機能を作る
      ↓
必要なEngine機能を特定
      ↓
Engineに実装
      ↓
ゲームで利用
      ↓
次の機能
```

---

# 28. 最初に作るVertical Slice

最初からWeather、Day/Night、Doorを全部実装しない。

まず以下を一本につなげる。

```text
Go/WASM
 ↓
WebGL
 ↓
Cube
 ↓
Camera
 ↓
GLB
 ↓
Animation
 ↓
Player
 ↓
MIDI
 ↓
Melody Recognition
 ↓
Game Event
 ↓
3D World Change
```

この流れが成立した時点で、技術的な実現可能性を確認する。

---

# 29. Engine開発における原則

### 原則1：必要最小限

今回のゲームに必要ない機能は作らない。

### 原則2：汎用化しすぎない

Three.jsやUnityのような汎用ゲームエンジンを作ろうとしない。

### 原則3：ゲームを優先する

エンジン開発そのものが目的にならないようにする。

### 原則4：実際に使う機能から作る

ゲーム側で必要になった機能をEngineに追加する。

### 原則5：動作確認を小さく行う

各Engine機能は実際のゲーム上で確認する。

---

# 30. Codexへの実装ルール

Codexは本プロジェクトの実装を支援する。

ただし、以下を厳守する。

## 実装前

必ず、

1. 現在のプロジェクト構成を確認
2. 関連コードを確認
3. 既存の設計を確認
4. 変更範囲を明確化

してから実装する。

## 実装時

* 既存コードを勝手に大規模変更しない
* 不要な依存ライブラリを追加しない
* Three.js等を勝手に導入しない
* MVP外の機能を勝手に実装しない
* エンジンを過剰に汎用化しない
* 1つの変更を小さく保つ
* 実装後に動作確認する

## 判断に迷った場合

勝手に仕様を拡張せず、既存の要件を優先する。

---

# 31. 技術選定の理由

Goを採用する主な理由は、Goが3Dゲーム開発における唯一の最適解だからではない。

今回の目的は、

```text
MIDI
 ↓
Melody Recognition
 ↓
Game Logic
 ↓
自作3D Runtime
 ↓
WebGL
 ↓
GPU
```

という処理系を自分たちで構築することである。

GoをWebAssemblyへコンパイルすることで、ゲームロジックと自作3DランタイムをGoで統一しながらブラウザ上で実行できる。

TypeScriptはブラウザ固有の処理に限定し、ゲームの主要部分をGo/WASMへ集約する。

この構成自体を本プロジェクトの技術的特徴とする。

---

# 32. 技術的なゴール

最終的に以下を実現する。

```text
MIDI Ocarina
      ↓
Web MIDI API
      ↓
TypeScript
      ↓
Go / WebAssembly
      ↓
Melody Recognition
      ↓
Game Event
      ↓
自作3D Engine
      ↓
WebGL
      ↓
GPU
      ↓
3D Game World
```

ユーザーが正しいメロディを演奏すると、

```text
Melody A → 天候が変わる

Melody B → 昼夜が変わる

Melody C → 隠し扉が開く
```

という体験を実現する。

---

# 33. MVP完成条件

以下をすべて満たした場合、MVP完成とする。

* [ ] ブラウザからMIDIデバイスを接続できる
* [ ] MIDI Noteを取得できる
* [ ] Go/WASMへMIDI情報を渡せる
* [ ] Go側でメロディを認識できる
* [ ] 自作3D Engineで3D空間を描画できる
* [ ] GLBモデルを表示できる
* [ ] プレイヤーを操作できる
* [ ] カメラがプレイヤーを追従する
* [ ] 基本アニメーションが再生できる
* [ ] 天候を変更できる
* [ ] 昼夜を変更できる
* [ ] 隠し扉を開けられる
* [ ] 扉のアニメーションが再生される
* [ ] MIDI演奏から3D世界の変化まで一連の流れが動作する

**最後の項目が最重要である。**

---

# 34. 優先順位

開発優先度は以下とする。

```text
最優先
↓
MIDI Input
↓
Go/WASM
↓
WebGL Renderer
↓
3D Model
↓
Player
↓
Animation
↓
Melody Recognition
↓
Game Event
↓
Weather / DayNight / Door
↓
演出・ポリッシュ
↓
追加機能
```

見た目の作り込みよりも、**MIDI → ゲームイベント → 3D世界の変化**というコア体験を優先する。
