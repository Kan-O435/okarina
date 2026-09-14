//go:build !js

package renderer

// Mesh はネイティブビルド向けのスタブ。
type Mesh struct{}

// NewMesh はネイティブビルドでは何もしない空のMeshを返す。
func (c *Context) NewMesh(positions, texcoords []float32, indices []uint16) *Mesh {
	return &Mesh{}
}

// Draw はネイティブビルドでは何もしない。
func (m *Mesh) Draw(c *Context, program *Program) {}
