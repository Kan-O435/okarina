//go:build js

package renderer

import (
	"errors"
	"syscall/js"
)

// syscall/js経由ではgl定数(gl.COLOR_BUFFER_BIT等)を直接参照できないため、
// 使用する定数はここで値を定義しておく。
const (
	glColorBufferBit = 0x00004000
	glDepthBufferBit = 0x00000100
	glDepthTest      = 0x0B71
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

// Viewport はWebGLの描画範囲を設定する。canvasのサイズと合わせて呼ぶ。
func (c *Context) Viewport(width, height int) {
	c.gl.Call("viewport", 0, 0, width, height)
}

// CanvasSize はcanvas要素の描画サイズ(width, height)を返す。
func (c *Context) CanvasSize() (int, int) {
	return c.canvas.Get("width").Int(), c.canvas.Get("height").Int()
}

// GL は生のWebGLコンテキスト(js.Value)を返す。
// シェーダーコンパイルやメッシュ描画など、Contextにまだラップされていない
// 低レベルなWebGL API呼び出しが必要になった際に使用する。
func (c *Context) GL() js.Value {
	return c.gl
}
