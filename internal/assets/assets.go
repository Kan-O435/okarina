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
