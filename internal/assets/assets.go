// Package assets は、3Dモデル等のバイナリアセットをgo:embedでコンパイル時に
// バイナリへ埋め込む。ブラウザ側で別途fetch()する手間を省くための簡易的な方法で、
// アセット数が増えてきたらHTTP fetch方式への切り替えを検討する。
package assets

import "embed"

//go:embed models/temple-body.glb
var TempleBody []byte

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
