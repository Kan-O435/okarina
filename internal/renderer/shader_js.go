//go:build js

package renderer

import (
	"errors"
	"syscall/js"

	"github.com/Kan-O435/okarina/internal/vecmath"
)

const (
	glVertexShader   = 0x8B31
	glFragmentShader = 0x8B30
	glCompileStatus  = 0x8B81
	glLinkStatus     = 0x8B82
)

// Program はコンパイル・リンク済みのWebGLシェーダープログラムをラップする。
type Program struct {
	handle js.Value
}

// NewProgram は頂点シェーダーとフラグメントシェーダーのGLSLソースから
// Programをコンパイル・リンクして生成する。
func (c *Context) NewProgram(vertexSrc, fragmentSrc string) (*Program, error) {
	vs, err := c.compileShader(glVertexShader, vertexSrc)
	if err != nil {
		return nil, err
	}
	fs, err := c.compileShader(glFragmentShader, fragmentSrc)
	if err != nil {
		return nil, err
	}

	program := c.gl.Call("createProgram")
	c.gl.Call("attachShader", program, vs)
	c.gl.Call("attachShader", program, fs)
	c.gl.Call("linkProgram", program)

	if !c.gl.Call("getProgramParameter", program, glLinkStatus).Bool() {
		infoLog := c.gl.Call("getProgramInfoLog", program).String()
		return nil, errors.New("renderer: program link failed: " + infoLog)
	}

	return &Program{handle: program}, nil
}

func (c *Context) compileShader(shaderType int, src string) (js.Value, error) {
	shader := c.gl.Call("createShader", shaderType)
	c.gl.Call("shaderSource", shader, src)
	c.gl.Call("compileShader", shader)

	if !c.gl.Call("getShaderParameter", shader, glCompileStatus).Bool() {
		infoLog := c.gl.Call("getShaderInfoLog", shader).String()
		return js.Value{}, errors.New("renderer: shader compile failed: " + infoLog)
	}
	return shader, nil
}

// Use はこのプログラムを以降の描画で使用するシェーダーとして有効化する。
func (p *Program) Use(c *Context) {
	c.gl.Call("useProgram", p.handle)
}

// SetUniformMat4 はmat4型のuniform変数に値を設定する。Use済みであること。
func (p *Program) SetUniformMat4(c *Context, name string, m vecmath.Mat4) {
	loc := c.gl.Call("getUniformLocation", p.handle, name)
	data := make([]float32, 16)
	for i, v := range m {
		data[i] = float32(v)
	}
	c.gl.Call("uniformMatrix4fv", loc, false, float32ArrayOf(data))
}

// SetUniformVec3 はvec3型のuniform変数に値を設定する。Use済みであること。
func (p *Program) SetUniformVec3(c *Context, name string, v vecmath.Vec3) {
	loc := c.gl.Call("getUniformLocation", p.handle, name)
	c.gl.Call("uniform3f", loc, v.X, v.Y, v.Z)
}

// SetUniformSampler はsampler2D型のuniform変数に、バインド先のテクスチャ
// ユニット番号(gl.TEXTURE0からのオフセット)を設定する。Use済みであること。
func (p *Program) SetUniformSampler(c *Context, name string, textureUnit int) {
	loc := c.gl.Call("getUniformLocation", p.handle, name)
	c.gl.Call("uniform1i", loc, textureUnit)
}

// SetUniformFloat はfloat型のuniform変数に値を設定する。Use済みであること。
func (p *Program) SetUniformFloat(c *Context, name string, v float64) {
	loc := c.gl.Call("getUniformLocation", p.handle, name)
	c.gl.Call("uniform1f", loc, v)
}
