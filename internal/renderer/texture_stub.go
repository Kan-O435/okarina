//go:build !js

package renderer

import "errors"

// Texture はネイティブビルド向けのスタブ。
type Texture struct{}

// WhiteTexture はネイティブビルドでは空のTextureを返す。
func (c *Context) WhiteTexture() *Texture {
	return &Texture{}
}

// NewImageTexture はネイティブビルドでは常にエラーを返す。
func (c *Context) NewImageTexture(data []byte, mimeType string) (*Texture, error) {
	return nil, errors.New("renderer: WebGL is only available in js/wasm builds")
}

// Bind はネイティブビルドでは何もしない。
func (t *Texture) Bind(c *Context) {}
