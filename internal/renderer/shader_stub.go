//go:build !js

package renderer

import (
	"errors"

	"github.com/Kan-O435/okarina/internal/vecmath"
)

// Program はネイティブビルド向けのスタブ。
type Program struct{}

// NewProgram はネイティブビルドでは常にエラーを返す。
func (c *Context) NewProgram(vertexSrc, fragmentSrc string) (*Program, error) {
	return nil, errors.New("renderer: WebGL is only available in js/wasm builds")
}

// Use はネイティブビルドでは何もしない。
func (p *Program) Use(c *Context) {}

// SetUniformMat4 はネイティブビルドでは何もしない。
func (p *Program) SetUniformMat4(c *Context, name string, m vecmath.Mat4) {}

// SetUniformVec3 はネイティブビルドでは何もしない。
func (p *Program) SetUniformVec3(c *Context, name string, v vecmath.Vec3) {}
