// Package assets は、3Dモデル等のバイナリアセットをgo:embedでコンパイル時に
// バイナリへ埋め込む。ブラウザ側で別途fetch()する手間を省くための簡易的な方法で、
// アセット数が増えてきたらHTTP fetch方式への切り替えを検討する。
package assets

import _ "embed"

//go:embed models/temple-body.glb
var TempleBody []byte

//go:embed models/link-knight.glb
var LinkKnight []byte
