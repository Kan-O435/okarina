//go:build !js

package renderer

import "errors"

// Context はネイティブビルド向けのスタブ。WebGLはブラウザ環境でのみ利用できるため、
// go build ./... やネイティブ実行(go run ./cmd/game)が通るように何もしない実装を提供する。
type Context struct{}

// NewContext はネイティブビルドでは常にエラーを返す。
func NewContext(canvasID string) (*Context, error) {
	return nil, errors.New("renderer: WebGL is only available in js/wasm builds")
}

// ClearColor はネイティブビルドでは何もしない。
func (c *Context) ClearColor(r, g, b, a float64) {}

// Clear はネイティブビルドでは何もしない。
func (c *Context) Clear() {}
