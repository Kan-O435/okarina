//go:build js

package renderer

import (
	"errors"
	"syscall/js"
)

const (
	glTexture2D        = 0x0DE1
	glRGBA              = 0x1908
	glUnsignedByte      = 0x1401
	glTextureMinFilter  = 0x2801
	glTextureMagFilter  = 0x2800
	glLinear            = 0x2601
	glClampToEdge       = 0x812F
	glTextureWrapS      = 0x2802
	glTextureWrapT      = 0x2803
	glTexture0          = 0x84C0
	glUnpackFlipYWebgl  = 0x9240
	glNearest           = 0x2600
)

// Texture はWebGLテクスチャオブジェクトをラップする。
type Texture struct {
	handle js.Value
}

// WhiteTexture は1x1の白テクスチャを返す(初回のみ生成し、以後は使い回す)。
// 実テクスチャを持たないオブジェクト(自作の地面・プレースホルダー等)に使う。
// texColor(白) * uColor = uColor になるため、実質uColorだけで色が決まる。
func (c *Context) WhiteTexture() *Texture {
	if c.whiteTexture == nil {
		c.whiteTexture = c.newSolidTexture(255, 255, 255, 255)
	}
	return c.whiteTexture
}

func (c *Context) newSolidTexture(r, g, b, a byte) *Texture {
	tex := c.gl.Call("createTexture")
	c.gl.Call("bindTexture", glTexture2D, tex)
	c.gl.Call("texImage2D", glTexture2D, 0, glRGBA, 1, 1, 0, glRGBA, glUnsignedByte,
		uint8ArrayOf([]byte{r, g, b, a}))
	c.applyTextureParams()
	return &Texture{handle: tex}
}

// NewImageTexture は画像バイナリ(JPEG/PNG等)をブラウザの画像デコーダで非同期に
// デコードし、WebGLテクスチャを生成する。
//
// 呼び出し中はGoのゴルーチンがチャネル受信でブロックするが、Go WASMの
// スケジューラはゴルーチンがブロックすると実行をブラウザのイベントループへ
// 戻すため、その間にPromiseの.then()コールバックが実行され、ページが
// フリーズすることはない(net/httpのWASM版fetch実装と同じパターン)。
func (c *Context) NewImageTexture(data []byte, mimeType string) (*Texture, error) {
	blobParts := js.Global().Get("Array").New(1)
	blobParts.SetIndex(0, uint8ArrayOf(data))
	options := js.Global().Get("Object").New()
	options.Set("type", mimeType)
	blob := js.Global().Get("Blob").New(blobParts, options)

	type decodeResult struct {
		bitmap js.Value
		err    error
	}
	done := make(chan decodeResult, 1)

	var onFulfilled, onRejected js.Func
	onFulfilled = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer onFulfilled.Release()
		defer onRejected.Release()
		done <- decodeResult{bitmap: args[0]}
		return nil
	})
	onRejected = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer onFulfilled.Release()
		defer onRejected.Release()
		msg := "unknown error"
		if len(args) > 0 {
			msg = args[0].String()
		}
		done <- decodeResult{err: errors.New("renderer: image decode failed: " + msg)}
		return nil
	})

	js.Global().Call("createImageBitmap", blob).Call("then", onFulfilled, onRejected)

	result := <-done
	if result.err != nil {
		return nil, result.err
	}

	tex := c.gl.Call("createTexture")
	c.gl.Call("bindTexture", glTexture2D, tex)
	c.gl.Call("pixelStorei", glUnpackFlipYWebgl, true)
	c.gl.Call("texImage2D", glTexture2D, 0, glRGBA, glRGBA, glUnsignedByte, result.bitmap)
	c.applyTextureParams()
	result.bitmap.Call("close")

	return &Texture{handle: tex}, nil
}

// NewPixelArtTexture はNewImageTextureと同じく画像バイナリをデコードして
// テクスチャを生成するが、拡大時の補間をNEAREST(最近傍)にする。ドット絵の
// スプライト(story.goの炎アニメーション等)を、LINEAR補間でぼかさず
// くっきりした見た目のまま拡大表示するために使う。
func (c *Context) NewPixelArtTexture(data []byte, mimeType string) (*Texture, error) {
	tex, err := c.NewImageTexture(data, mimeType)
	if err != nil {
		return nil, err
	}
	tex.Bind(c)
	c.gl.Call("texParameteri", glTexture2D, glTextureMinFilter, glNearest)
	c.gl.Call("texParameteri", glTexture2D, glTextureMagFilter, glNearest)
	return tex, nil
}

// NewRGBATexture は、既にデコード済みのRGBA8ピクセルデータ(width*height*4
// バイト、行は画像の上から下、アルファは非事前乗算)から直接WebGLテクスチャを
// 作る。NewImageTexture/NewPixelArtTextureとは異なりブラウザの画像デコーダを
// 経由しないため、Go側で先にデコードしたデータ(image/gifでデコードした
// アニメーションGIFの各フレーム等、createImageBitmapでは最初の1コマしか
// 取り出せない形式)をそのままテクスチャ化したい場合に使う
// (internal/renderer/story.goのStoryBackground参照)。
func (c *Context) NewRGBATexture(pix []byte, width, height int) *Texture {
	tex := c.gl.Call("createTexture")
	c.gl.Call("bindTexture", glTexture2D, tex)
	c.gl.Call("pixelStorei", glUnpackFlipYWebgl, true)
	c.gl.Call("texImage2D", glTexture2D, 0, glRGBA, width, height, 0, glRGBA, glUnsignedByte, uint8ArrayOf(pix))
	c.applyTextureParams()
	return &Texture{handle: tex}
}

func (c *Context) applyTextureParams() {
	c.gl.Call("texParameteri", glTexture2D, glTextureMinFilter, glLinear)
	c.gl.Call("texParameteri", glTexture2D, glTextureMagFilter, glLinear)
	c.gl.Call("texParameteri", glTexture2D, glTextureWrapS, glClampToEdge)
	c.gl.Call("texParameteri", glTexture2D, glTextureWrapT, glClampToEdge)
}

// Bind はこのテクスチャをテクスチャユニット0にバインドする。
func (t *Texture) Bind(c *Context) {
	c.gl.Call("activeTexture", glTexture0)
	c.gl.Call("bindTexture", glTexture2D, t.handle)
}
