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

// Mesh はGPU上にアップロード済みの頂点バッファ・インデックスバッファをラップする。
type Mesh struct {
	vertexBuffer js.Value
	indexBuffer  js.Value
	indexCount   int
}

// NewMesh は頂点座標(x, y, z の並び)とインデックスからMeshを生成し、GPUへアップロードする。
func (c *Context) NewMesh(positions []float32, indices []uint16) *Mesh {
	vertexBuffer := c.gl.Call("createBuffer")
	c.gl.Call("bindBuffer", glArrayBuffer, vertexBuffer)
	c.gl.Call("bufferData", glArrayBuffer, float32ArrayOf(positions), glStaticDraw)

	indexBuffer := c.gl.Call("createBuffer")
	c.gl.Call("bindBuffer", glElementArrayBuffer, indexBuffer)
	c.gl.Call("bufferData", glElementArrayBuffer, uint16ArrayOf(indices), glStaticDraw)

	return &Mesh{vertexBuffer: vertexBuffer, indexBuffer: indexBuffer, indexCount: len(indices)}
}

// Draw はこのMeshを、programの"aPosition"属性にバインドして描画する。
// programはあらかじめ Use() 済みであること。
func (m *Mesh) Draw(c *Context, program *Program) {
	c.gl.Call("bindBuffer", glArrayBuffer, m.vertexBuffer)
	loc := c.gl.Call("getAttribLocation", program.handle, "aPosition")
	c.gl.Call("enableVertexAttribArray", loc)
	c.gl.Call("vertexAttribPointer", loc, 3, glFloat, false, 0, 0)

	c.gl.Call("bindBuffer", glElementArrayBuffer, m.indexBuffer)
	c.gl.Call("drawElements", glTriangles, m.indexCount, glUnsignedShort, 0)
}
