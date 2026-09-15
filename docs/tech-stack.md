# 技術スタック(実装ベース)

`CLAUDE.md`の「5. 技術スタック」は開発方針として最初に決めたものだが、この文書は
**実際にリポジトリで使われている技術を、実装済みのコードから棚卸ししたもの**。
方針と実態がずれている箇所(末尾の「CLAUDE.mdとの差分」参照)も含めて記録する。

## コア

| 項目 | 技術 | 備考 |
|---|---|---|
| 言語 | Go 1.22 (`go.mod`) | ゲームロジック・自作3Dエンジンの両方をGoで統一 |
| 実行環境 | WebAssembly(`GOOS=js GOARCH=wasm`) | フィールド(ページ)ごとに別バイナリ(`cmd/game`, `cmd/grassland`, `cmd/ganon-hall`, `cmd/ganon-battle`, `cmd/title`)をビルドする構成。共通ロジックは`internal/`配下のパッケージに切り出して各`cmd`から参照する |
| Go/JS連携 | `syscall/js` + 標準の`wasm_exec.js`(Goが生成するグルーコード) | `internal/bridge`パッケージがJS⇄Goの相互呼び出しを仲介する層。`//go:build js`と`//go:build !js`でファイルを分け、ネイティブ(`go build`/`go test`)ではstub実装に差し替わる |

## 3Dレンダリング(自作エンジン、`internal/renderer`)

既存の3Dエンジン(Three.js等)は使わず、WebGL1を直接叩く薄いラッパーを自作している。

| 項目 | 技術/実装 | 備考 |
|---|---|---|
| グラフィックスAPI | WebGL1(`canvas.getContext("webgl")`) | WebGL2は未使用 |
| シェーダー | GLSL ES 1.00 | 全オブジェクト・HUD・オーバーレイ共通の1組の頂点/フラグメントシェーダー(`shaders.go`)。`uMVP`(モデル・ビュー・プロジェクション)、`uColor`(vec3の単色乗算)、`uAlpha`(float、オーバーレイ用)の3 uniformのみ |
| モデル形式 | glTF 2.0 / GLB | パーサーは自作(`internal/gltf`)。POSITION/TEXCOORD_0とインデックス、`baseColorFactor`/`baseColorTexture`、ノードのワールド変換行列の焼き込み(`ParseParts`)に対応。法線・PBRマテリアル(メタリック/ラフネス等)・アニメーション・スキニングは未対応 |
| テクスチャ | `createImageBitmap`によるJPEG/PNGの非同期デコード→WebGLテクスチャ化 | Goのゴルーチン+チャネルでPromiseの完了を待つ(`net/http`のWASM実装と同じ手法) |
| アセット組み込み | `go:embed` | 3Dモデル(.glb)・テクスチャ(.png)をコンパイル時にバイナリへ埋め込み、ブラウザ側の別fetchを不要にしている(`internal/assets`) |
| 数学 | 自作(`internal/vecmath`) | Vec3, Mat4(Translate/Scale/RotateX/Y/Z, Perspective, Ortho, LookAt)。Quaternionは未実装 |

## ブラウザ側(`web/`)

| 項目 | 技術 | 備考 |
|---|---|---|
| 言語 | **JavaScript(素のES、`.js`)** | ビルドツール・バンドラ・npm/node_modulesは無し。`<script>`タグで直接読み込む |
| MIDI入力 | Web MIDI API(`web/midi-input.js`) | ノートオン/オフ・ベロシティ・タイムスタンプを取得しGo側へ橋渡し |
| ピッチ検出 | Web Audio API(`AnalyserNode`)+ 自己相関法(`web/pitch.js`) | オタマトーン(マイク入力)の音程からプレイヤー移動方向を判定する。RMS→dB変換でしきい値以下は無視 |
| 効果音・BGM | Web Audio API(`web/audio.js`ほか各フィールド専用JS) | 2方式を併用: (1) 事前fetchしたmp3を`decodeAudioData`して`AudioBufferSourceNode`で再生、(2) ホワイトノイズ+`BiquadFilterNode`+`GainNode`によるその場合成(雷鳴など、音声ファイルを追加しない効果音) |
| ページ遷移 | `window.location.href`書き換え(`Context.Navigate`経由) | フィールドごとに別ページ(別.html/.wasm)なので、SPA的なルーティングではなく実際のページ遷移 |

## アセット調達

| 種別 | 調達元 | 備考 |
|---|---|---|
| キャラクター・敵・小物モデル | Sketchfab(CC0 / CC Attribution) | `docs/licenses/`に出典・ライセンスを個別記録。利用のたびにダウンロード→`internal/assets/models/`に配置→`go:embed` |
| 背景・建物モデル | Tripo3D(AI生成)+ 手動リトポロジー | 神殿本体・草原の城・ガノンフィールド背景など。頂点数が自作ローダーの上限(uint16インデックス=65,535頂点)を超えないようリメッシュしてから利用 |
| 一部CC0モデル | ambientCG, Quaternius, KayKit | 扉テクスチャ、馬モデル等 |
| 手続き生成テクスチャ | Go(`image`/`image/color`/`image/png`標準パッケージ) | 草原の地面テクスチャ(多重スケールのvalue noise)、後光・モヤ用のグラデーション画像など。PillowやImageMagick等の外部ツールは使わずGoで自作 |

## テスト・検証

| 項目 | 技術 | 備考 |
|---|---|---|
| ユニットテスト | Go標準の`testing`パッケージ、`go test ./...` | `internal/{game,gltf,music,player,renderer,vecmath,world}`に配置。JSに依存する部分は`!js`ビルドタグのstub経由でネイティブ実行できるようにしてある |
| 手動検証ツール | `cmd/gltfcheck` | ブラウザを開かずにGLBファイルが自作ローダーで読み込めるか確認する小さなCLI |
| ブラウザ動作確認 | Claude in Chrome(このプロジェクトの開発で使用) | 実装後に必ず実機(ブラウザ)で見た目・挙動を確認する運用 |

## ビルド・配信

| 項目 | 技術 | 備考 |
|---|---|---|
| WASMビルド | `GOOS=js GOARCH=wasm go build -o web/<field>.wasm ./cmd/<field>` | フィールドごとに手動実行(Makefile等の自動化は無し) |
| ローカルサーバー | `python -m http.server`(`web/`ディレクトリを静的配信) | Node.js製の開発サーバーやHMRは使わない |
| デプロイ | 未整備(ローカル開発のみ) | 本番デプロイ先・CI/CDパイプラインは今のところ無い |

## バージョン管理・開発フロー

| 項目 | 技術 | 備考 |
|---|---|---|
| VCS | Git / GitHub(`Kan-O435/okarina`) | フィーチャーブランチ+Pull Requestで`dev`へマージし、`dev`から`main`へ、という運用 |
| 並行開発 | 同一ローカル作業ディレクトリを複数セッション(複数のClaude Codeセッション/開発者)が共有 | ファイル単位で「直前に読んでから編集する」「既存行を極力変更せず追記する」ことでコンフリクトを避ける運用が定着している |

## CLAUDE.mdとの差分

方針(`CLAUDE.md`)からは以下の点で実態がずれている:

- **ブラウザ側言語**: `CLAUDE.md`は「TypeScript」と明記しているが、実際は素のJavaScript
  (`.js`)のみで、TypeScriptコンパイラ・型定義・`tsconfig.json`等は一切導入されていない。
- **DB**: `CLAUDE.md`通り、MVPでは未導入(将来的な想定のみ)。
