// Package assets は、3Dモデル等のバイナリアセットをgo:embedでコンパイル時に
// バイナリへ埋め込む。ブラウザ側で別途fetch()する手間を省くための簡易的な方法で、
// アセット数が増えてきたらHTTP fetch方式への切り替えを検討する。
package assets

import "embed"

//go:embed models/temple-body.glb
var TempleBody []byte

// LinkKnight は、3フィールド(神殿/草原/ガノン)共通で使うプレイヤーキャラクター
// LinkのGLB。もとはKayKit Adventurers(CC0)のKnightモデルだったが、Sketchfabの
// ファンアートモデル「Ocarina of Time Link」(作者: projectmgame、CC
// Attribution、https://sketchfab.com/3d-models/ocarina-of-time-link-c62717add333410987482d44959e56c7、
// 元はSuper Smash Bros. Brawl改造版「Project M」のLinkの代替コスチューム)に
// 差し替えている。GanonBoss/GanonBossBattleと同じく任天堂の実際のLinkの
// デザインをそのまま再現したファンアートであり、リスクを理解した上で
// ユーザーの指示により採用している。詳細は
// docs/licenses/sketchfab-projectmgame-link-CC-BY.txt参照。スキンは無く、
// パーツごとに別メッシュ・別テクスチャへ分かれているため、GanonBoss等と同じ
// LoadGLBParts/CombinedGroundTransformで読み込む(internal/renderer/demo.goの
// linkObject参照)。
//
//go:embed models/link-knight.glb
var LinkKnight []byte

// Trees は、フィールドに配置する木(CC0、Gobkit Nature Kit)のGLB群。
// 本数が多いので個別変数ではなくembed.FSでまとめて埋め込む。
//
//go:embed models/trees/*.glb
var Trees embed.FS

// DoorTexture は、隠し扉の見た目に使うテクスチャ画像(CC0、ambientCG Door002の
// プレビュー画像、扉の外側が透過PNG)。3Dモデル自体はまだ無いため、
// 板(quad)にこの画像を貼って「扉らしい見た目」にする。
//
//go:embed textures/door.png
var DoorTexture []byte

// SongOfTimeSheetTexture は、「時の歌」(ラ→レ→ファ→ラ→レ→ファ)の楽譜を
// 表すバナー画像(背景透過PNG、ChatGPT生成)。ゲーム画面のHUDとして
// 常時オーバーレイ表示する(internal/renderer.BuildSongSheetHUD参照)。
//
//go:embed textures/song-of-time-sheet.png
var SongOfTimeSheetTexture []byte

// GrasslandCastle は、草原フィールドの道の先に置く「白い城」のGLB
// (Tripo3Dで生成・リメッシュ済み、テクスチャ無しの単色モデル)。
//
//go:embed models/grassland-castle.glb
var GrasslandCastle []byte

// HaloTexture は、城の後ろに浮かべる白い後光(放射状グラデーション、
// 中心が不透明な白・外側が透明)の板に貼るテクスチャ。自作の手続き生成画像。
//
//go:embed textures/halo.png
var HaloTexture []byte

// CloudTexture は、空に浮かべる雲の板に貼るテクスチャ(不定形の白い塊、
// 外側が透明)。自作の手続き生成画像。色はObject.Colorで紫に着色する。
//
//go:embed textures/cloud.png
var CloudTexture []byte

// GanonBattleScene は、ガノンフィールドの背景となる、荒れ果てた戦場跡の
// ジオラマのGLB(Tripo3Dで生成・リメッシュ・テクスチャ生成済み)。横長・
// 低めの「舞台」状の形をしている。
//
//go:embed models/ganon-battle.glb
var GanonBattleScene []byte

// GanonHall は、ガノンフィールドの背景デザイン案として比較中の「玉座の間」
// (Tripo3Dで生成、テクスチャ付き)。GanonBattleSceneより縦横比が立方体に
// 近く(幅:高さ:奥行 ≈ 1.4:1:0.9)、コンパクトな形をしている。
//
//go:embed models/ganon-hall.glb
var GanonHall []byte

// GanonBoss は、玉座の間案(GanonBackgroundHall)の背景の手前に配置する
// ボス役のGLB。Sketchfabのファンアートモデル「Zelda Ocarina of time Fan
// Art: Ganon」(作者: totidoki、CC Attribution、
// https://sketchfab.com/3d-models/zelda-ocarina-of-time-fan-art-ganon-868ad08b11e245078410ae8aa70a7068)の
// 「converted」形式(609KB、人型のガノンドロフ姿)。これは任天堂の実際の
// ガノン(時のオカリナ)のデザインをそのまま再現したものであり、以前
// 使っていたKayKit Barbarian(着想を得たオリジナルデザイン)とは異なり、
// 任天堂の知的財産権に関するリスクを理解した上でユーザーの指示により
// 採用している。詳細はdocs/licenses/sketchfab-totidoki-ganon-CC-BY.txt
// 参照。スキンは無く、パーツごとに別メッシュ・別テクスチャへ分かれて
// いるため、LoadGLBParts/CombinedGroundTransformで読み込む
// (internal/renderer/ganon.goのganonBossObject参照)。
//
//go:embed models/ganon-boss.glb
var GanonBoss []byte

// GanonBossBattle は、戦場跡案(GanonBackgroundBattle)の背景の手前に
// 配置するボス役のGLB。同じSketchfabモデルの、作者のオリジナルFBXから
// 変換したより高品質な版(緑色・角のある獣形態のガノン)。元のテクスチャは
// 4096px等の非圧縮PNGで合計約66MBあったため、1024px以下・JPEG(quality
// 80程度)に圧縮して約3.8MBまで削減している(このエンジンはbaseColor
// Textureしか読まないため、法線・金属度等のテクスチャも同様に圧縮して
// いるが実際には使われない)。GanonBossとはバウンディングボックスの
// 縦横比が大きく異なる別モデルのため、専用のganonBossBattleZ/Y/Height
// で配置する。
//
//go:embed models/ganon-boss-battle.glb
var GanonBossBattle []byte

// Horse は、草原フィールドで「馬の歌」を演奏すると呼び出される馬のGLB
// (CC0、Quaternius "Ultimate Animated Animal Pack"、docs/assets/horse/
// README.md参照)。テクスチャ無し、マテリアルの単色のみで構成される
// 低ポリスタイル。アニメーション・スキニングは未実装のため静止ポーズで表示する。
//
//go:embed models/horse.glb
var Horse []byte

// HorseSongSheetTexture は、「馬の歌」(レ→シ→ラ→レ→シ→ラ)の楽譜を
// 表すバナー画像(背景透過PNG、ChatGPT生成)。草原フィールドで障害物に
// 近づいた時だけHUDとして表示する(internal/renderer.BuildHorseSongSheetHUD、
// game.IsNearGrasslandObstacle参照)。
//
//go:embed textures/horse-song-sheet.png
var HorseSongSheetTexture []byte

// TitleOcarina は、タイトル画面の背景でくるくる回るオカリナのGLB
// (テクスチャ付き、Tripo3D等のセグメンテーション機能で3パーツに分かれて
// 出力されたモデル)。internal/renderer.BuildTitleSceneがLoadGLBPartsで
// 読み込み、パーツごとに同じ回転Transformを適用して描画する。
//
//go:embed models/title-ocarina.glb
var TitleOcarina []byte
