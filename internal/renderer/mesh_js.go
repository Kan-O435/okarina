//go:build js

package renderer

import "syscall/js"

const (
	glArrayBuffer        = 0x8892
	glElementArrayBuffer = 0x8893
	glStaticDraw         = 0x88E4
	glFloat              = 0x1406
	glUnsignedShort      = 0x1403
	glTriangles          = 0x0004
)

// Mesh はGPU上にアップロード済みの頂点バッファ・UVバッファ・インデックス
// バッファをラップする。
type Mesh struct {
	vertexBuffer   js.Value
	texcoordBuffer js.Value
	indexBuffer    js.Value
	indexCount     int
}

// NewMesh は頂点座標(x, y, z の並び)・UV座標(u, v の並び)・インデックスから
// Meshを生成し、GPUへアップロードする。
func (c *Context) NewMesh(positions, texcoords []float32, indices []uint16) *Mesh {
	vertexBuffer := c.gl.Call("createBuffer")
	c.gl.Call("bindBuffer", glArrayBuffer, vertexBuffer)
	c.gl.Call("bufferData", glArrayBuffer, float32ArrayOf(positions), glStaticDraw)

	texcoordBuffer := c.gl.Call("createBuffer")
	c.gl.Call("bindBuffer", glArrayBuffer, texcoordBuffer)
	c.gl.Call("bufferData", glArrayBuffer, float32ArrayOf(texcoords), glStaticDraw)

	indexBuffer := c.gl.Call("createBuffer")
	c.gl.Call("bindBuffer", glElementArrayBuffer, indexBuffer)
	c.gl.Call("bufferData", glElementArrayBuffer, uint16ArrayOf(indices), glStaticDraw)

	return &Mesh{
		vertexBuffer:   vertexBuffer,
		texcoordBuffer: texcoordBuffer,
		indexBuffer:    indexBuffer,
		indexCount:     len(indices),
	}
}

// Draw はこのMeshを、programの"aPosition"/"aTexCoord"属性にバインドして描画する。
// programはあらかじめ Use() 済みであること。
func (m *Mesh) Draw(c *Context, program *Program) {
	c.gl.Call("bindBuffer", glArrayBuffer, m.vertexBuffer)
	posLoc := c.gl.Call("getAttribLocation", program.handle, "aPosition")
	c.gl.Call("enableVertexAttribArray", posLoc)
	c.gl.Call("vertexAttribPointer", posLoc, 3, glFloat, false, 0, 0)

	c.gl.Call("bindBuffer", glArrayBuffer, m.texcoordBuffer)
	uvLoc := c.gl.Call("getAttribLocation", program.handle, "aTexCoord")
	c.gl.Call("enableVertexAttribArray", uvLoc)
	c.gl.Call("vertexAttribPointer", uvLoc, 2, glFloat, false, 0, 0)

	c.gl.Call("bindBuffer", glElementArrayBuffer, m.indexBuffer)
	c.gl.Call("drawElements", glTriangles, m.indexCount, glUnsignedShort, 0)
}
