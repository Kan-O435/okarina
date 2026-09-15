// Package assets は、3Dモデル等のバイナリアセットをgo:embedでコンパイル時に
// バイナリへ埋め込む。ブラウザ側で別途fetch()する手間を省くための簡易的な方法で、
// アセット数が増えてきたらHTTP fetch方式への切り替えを検討する。
package assets

import _ "embed"

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

// TempleTree は、神殿・草原フィールド共通で使う木のGLB(Sketchfab
// 「Stylize Tree Lowpoly」、作者: uday、CC Attribution、
// https://sketchfab.com/3d-models/stylize-tree-lowpoly-dcf20a5a86784331a34657e94442511f)。
// 以前使っていたCC0(Gobkit Nature Kit)の木より作り込まれた見た目の
// 単色針葉樹モデルに、両フィールドとも差し替えている(ガノンフィールドには
// 木を置いていない)。詳細はdocs/licenses/sketchfab-uday-stylize-tree-CC-BY.txt
// 参照。スキンは無い単一メッシュだが、Sketchfabのconverted形式でルート
// ノードに軸補正の変換行列が入っているため、これを焼き込むgltf.ParseParts
// 経由で読み込む(internal/renderer/demo.goのtreeObjects、grassland.goの
// grasslandTreeObjects参照。どちらもparts[0]を使う)。
//
//go:embed models/temple-tree.glb
var TempleTree []byte

// TempleCloud は、神殿・草原フィールド共通で空に浮かべる雲のGLB(Sketchfab
// 「Stylized Clouds Pack - Vol 09」、作者: PolyOne Studio、CC Attribution、
// https://sketchfab.com/3d-models/stylized-clouds-pack-vol-09-d9c1ff67f80841c6b1d8229dca5495a5)。
// 16種類のブロック調の雲メッシュが1つのGLBにまとまっており、そのうち
// いくつかを選んで上空に手動配置する(internal/renderer/demo.goの
// templeCloudObjects、grassland.goのgrasslandTempleCloudObjects参照。
// 草原フィールドでは、以前使っていたCloudTexture(平面の板)の雲をこちらに
// 差し替えている。CloudTextureは草原の城の周りのモヤ用として引き続き使う)。
// スキンは無くパーツごとに別メッシュへ分かれているため、GanonBoss等と
// 同じLoadGLBParts/CombinedGroundTransformで読み込む。詳細は
// docs/licenses/sketchfab-polyonestudio-clouds-CC-BY.txt参照。
//
//go:embed models/temple-cloud.glb
var TempleCloud []byte

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

// CloudTexture は、柔らかい白い塊(外側が透明)の板に貼る自作の手続き生成
// テクスチャ。草原フィールドでは、城の周りに軽く漂わせるモヤに使う
// (internal/renderer/grassland.goのgrasslandSkyObjects参照。城の後光と
// 同様、Object.Colorで着色する)。以前は同じ板に雲としても使っていたが、
// 神殿・草原とも雲はTempleCloud(立体的なメッシュ)に差し替えている。
//
//go:embed textures/cloud.png
var CloudTexture []byte

// GrasslandGroundTexture は、草原フィールドの地面に貼る草のテクスチャ
// (オリーブ〜黄緑〜土色がまだらに混ざった、乾いた草原らしい見た目)。
// 自作の手続き生成画像(複数スケールのvalue noiseを重ねて生成)。
// 地面は非常に広い1枚のquad(internal/renderer/grassland.goの
// grasslandGroundObject参照)にタイリングせずそのまま貼るため、
// 実際にプレイヤーが動き回る範囲(中心付近)では模様が大きく・
// ゆるやかに見える。
//
//go:embed textures/grassland-ground.png
var GrasslandGroundTexture []byte

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

// RockDebris は、玉座の間(GanonBackgroundHall)が崩れる演出で降らせる岩の
// GLB。Sketchfabの「Stone Pack」(作者: ashkan.fancy、CC Attribution、
// https://sketchfab.com/3d-models/stone-pack-f3e0a67b9ca243b09119177649f21e17)
// のうち、大小の岩(Big/Mid/Small、計19パーツ)だけを抜き出したもの
// (装飾用のルーン岩・柱パーツは崩落演出には不要なため除外)。元のGLBは
// テクスチャ込みで約11.6MBあったため、baseColorTextureとして実際に
// 使われる3枚のみ残し、512px・JPEG(quality 80程度)に圧縮して約2.1MBまで
// 削減している。パーツごとに別メッシュ・別テクスチャへ分かれているため、
// GanonBoss等と同様LoadRockDebrisParts(gltf.ParseParts)で読み込む。
// 詳細はdocs/licenses/sketchfab-ashkanfancy-stonepack-CC-BY.txt参照。
//
//go:embed models/rock-debris.glb
var RockDebris []byte

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

// GanonHallMelodySheetTexture は、玉座の間でGanonの第一形態を倒す合図
// 「光のプレリュード」(レ→ラ→レ→ラ→シ→レ)の楽譜を表すバナー画像
// (背景透過PNG、ユーザー提供)。崩落演出が始まる前だけHUDとして表示する
// (internal/renderer.BuildGanonHallMelodySheetHUD、
// game.GanonHallMelodyPlayed参照)。
//
//go:embed textures/ganon-hall-melody-sheet.png
var GanonHallMelodySheetTexture []byte

// GanonBattleMelodySheetTexture は、戦場跡でGanonの最終形態を倒す合図
// 「嵐の歌」(レ→ファ→レ→レ→ファ→レ)の楽譜を表すバナー画像
// (背景透過PNG、ユーザー提供)。撃破演出が始まる前だけHUDとして表示する
// (internal/renderer.BuildGanonBattleMelodySheetHUD、
// game.GanonBattleMelodyPlayed参照)。
//
//go:embed textures/ganon-battle-melody-sheet.png
var GanonBattleMelodySheetTexture []byte

// GanonLightning は、ganon-battleフィールド(戦場跡案)でGanonの最終形態を
// 倒すカットシーンの雷に使うGLB(Sketchfab「3 Pack of Storm Lightning」、
// 作者: Incg5764、CC Attribution、
// https://sketchfab.com/3d-models/3-pack-of-storm-lightning-8acf90c132754b6bb9ab2b60ebd9465f)。
// 分岐した雷のメッシュが3種類まとまっており、そのうち縦長で1本に近い
// 見た目のものを1つ選んでGanonの頭上に落とす
// (internal/renderer/ganon.goのGanonBattleLightningObject参照)。スキンは
// 無くパーツごとに別メッシュへ分かれているため、TempleCloud等と同じ
// gltf.ParseParts経由で読み込む。詳細は
// docs/licenses/sketchfab-incg5764-stormlightning-CC-BY.txt参照。
//
//go:embed models/ganon-lightning.glb
var GanonLightning []byte

// GanonStormCloud は、ganon-battleフィールドでGanonの最終形態を倒す
// カットシーンの冒頭、アリーナ上空に集まる嵐雲のGLB(Sketchfab「Low Poly
// Cloud 3D Model Game Asset UI Element」、作者: B1Blender、CC Attribution、
// https://sketchfab.com/3d-models/low-poly-cloud-3d-model-game-asset-ui-element-16eb4c30016a4a4dafd70f2daf75c501)。
// 元は明るい水色だが、暗い青灰色に着色して不穏な嵐の雲として使う
// (internal/renderer/ganon.goのGanonBattleStormCloudObjects参照。
// 神殿・草原の雲(TempleCloud)とは別モデル)。7種類のフラットカラーの
// 雲メッシュがまとまっており、いくつかを選んで手動配置する。スキンは
// 無くパーツごとに別メッシュへ分かれているため、TempleCloud等と同じ
// gltf.ParseParts経由で読み込む。詳細は
// docs/licenses/sketchfab-b1blender-lowpolycloud-CC-BY.txt参照。
//
//go:embed models/storm-cloud.glb
var GanonStormCloud []byte

// Zelda は、エンディング画面でLinkと向かい合って立つゼルダ姫のGLB
// (ユーザー提供、Ocarina of Time風ファンアートモデル)。LinkKnightと同じく
// スキンは無くパーツごとに別メッシュ・別テクスチャへ分かれているため、
// LoadGLBParts/CombinedGroundTransformで読み込む
// (internal/renderer/ending.go参照)。
//
//go:embed models/zelda.glb
var Zelda []byte

// FlowerFieldTexture は、エンディング画面の地面に貼る花畑のテクスチャ
// (緑の草地に白・黄・ピンク・紫の小さな花を散らした模様)。自作の手続き
// 生成画像(internal/renderer/grassland.goのGrasslandGroundTexture等と
// 同様のアプローチ)。
//
//go:embed textures/flower-field.png
var FlowerFieldTexture []byte

// PetalTexture は、エンディング画面で舞い散る花びら1枚分のテクスチャ
// (背景透過PNG、先端に切れ込みのある桜の花びらのシルエット)。自作の
// 手続き生成画像。1枚のメッシュ・テクスチャを全ての花びらのObjectで
// 使い回す(internal/renderer/ending.go参照)。
//
//go:embed textures/petal.png
var PetalTexture []byte
