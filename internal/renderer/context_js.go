//go:build js

package renderer

import (
	"errors"
	"syscall/js"
)

// syscall/js経由ではgl定数(gl.COLOR_BUFFER_BIT等)を直接参照できないため、
// 使用する定数はここで値を定義しておく。
const (
	glColorBufferBit   = 0x00004000
	glDepthBufferBit   = 0x00000100
	glDepthTest        = 0x0B71
	glBlend            = 0x0BE2
	glSrcAlpha         = 0x0302
	glOneMinusSrcAlpha = 0x0303
)

// Context は取得済みのWebGL1レンダリングコンテキストをラップする。
// 今後のシェーダー・メッシュ描画は、このContextを起点に実装していく。
type Context struct {
	canvas       js.Value
	gl           js.Value
	whiteTexture *Texture // WhiteTexture()で遅延生成してキャッシュする
}

// NewContext はDOM上のcanvas要素(id指定)からWebGL1コンテキストを取得する。
// canvas要素が見つからない場合、またはブラウザがWebGLに対応していない場合はエラーを返す。
func NewContext(canvasID string) (*Context, error) {
	canvas := js.Global().Get("document").Call("getElementById", canvasID)
	if canvas.IsNull() || canvas.IsUndefined() {
		return nil, errors.New("renderer: canvas element not found: " + canvasID)
	}

	gl := canvas.Call("getContext", "webgl")
	if gl.IsNull() || gl.IsUndefined() {
		return nil, errors.New("renderer: failed to get WebGL1 context (browser may not support WebGL)")
	}

	return &Context{canvas: canvas, gl: gl}, nil
}

// ClearColor は画面クリア時の色を設定する。各成分は0.0〜1.0。
func (c *Context) ClearColor(r, g, b, a float64) {
	c.gl.Call("clearColor", r, g, b, a)
}

// Clear はカラーバッファ・深度バッファをクリアする。
func (c *Context) Clear() {
	c.gl.Call("clear", glColorBufferBit|glDepthBufferBit)
}

// EnableDepthTest は深度テストを有効にする。
// 手前・奥にあるオブジェクトの重なりを正しく描画するために必要。
func (c *Context) EnableDepthTest() {
	c.gl.Call("enable", glDepthTest)
}

// DisableDepthTest は深度テストを無効にする。HUDのような、3Dシーンより
// 必ず手前に(奥行きを無視して)描画したい要素を描く前に呼ぶ。
func (c *Context) DisableDepthTest() {
	c.gl.Call("disable", glDepthTest)
}

// EnableBlend はアルファブレンディングを有効にする。
// 扉画像のような、背景が透過(アルファ<1)なテクスチャを正しく合成するために使う。
func (c *Context) EnableBlend() {
	c.gl.Call("enable", glBlend)
	c.gl.Call("blendFunc", glSrcAlpha, glOneMinusSrcAlpha)
}

// Viewport はWebGLの描画範囲を設定する。canvasのサイズと合わせて呼ぶ。
func (c *Context) Viewport(width, height int) {
	c.gl.Call("viewport", 0, 0, width, height)
}

// CanvasSize はcanvas要素の描画サイズ(width, height)を返す。
func (c *Context) CanvasSize() (int, int) {
	return c.canvas.Get("width").Int(), c.canvas.Get("height").Int()
}

// RunLoop はrequestAnimationFrameを使い、毎フレームcallbackを呼び出し続ける。
// callbackには前フレームからの経過時間(秒、dt)が渡される(1フレーム目は0)。
func (c *Context) RunLoop(callback func(dt float64)) {
	var frame js.Func
	lastTime := 0.0
	frame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := args[0].Float()
		dt := 0.0
		if lastTime > 0 {
			dt = (now - lastTime) / 1000
		}
		lastTime = now

		callback(dt)

		js.Global().Call("requestAnimationFrame", frame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", frame)
}

// Navigate はブラウザを指定したURLへ遷移させる(window.location.href = url)。
// フィールドをまたぐページ遷移(例: 扉を抜けた先の次のフィールドへ移動)に使う。
func (c *Context) Navigate(url string) {
	js.Global().Get("window").Get("location").Set("href", url)
}

// GL は生のWebGLコンテキスト(js.Value)を返す。
// シェーダーコンパイルやメッシュ描画など、Contextにまだラップされていない
// 低レベルなWebGL API呼び出しが必要になった際に使用する。
func (c *Context) GL() js.Value {
	return c.gl
}
