package renderer

import "github.com/Kan-O435/okarina/internal/vecmath"

// Object は描画する1つのMeshと、その配置(モデル行列)・テクスチャ・色を
// まとめたもの。Textureがnilの場合はContext.WhiteTexture()を使う
// (実質Colorだけで色が決まる)。
type Object struct {
	Mesh      *Mesh
	Texture   *Texture
	Transform vecmath.Mat4
	Color     vecmath.Vec3
}

// Scene は描画に使うシェーダー、カメラ(Projection*View済みの行列)、
// 描画対象のObject一覧をまとめたもの。
type Scene struct {
	Program        *Program
	ViewProjection vecmath.Mat4
	Objects        []Object
}

// Render はScene内の全Objectを描画する。
func (s *Scene) Render(c *Context) {
	c.Clear()
	s.Program.Use(c)
	s.Program.SetUniformSampler(c, "uTexture", 0)

	for _, obj := range s.Objects {
		texture := obj.Texture
		if texture == nil {
			texture = c.WhiteTexture()
		}
		texture.Bind(c)

		mvp := s.ViewProjection.Mul(obj.Transform)
		s.Program.SetUniformMat4(c, "uMVP", mvp)
		s.Program.SetUniformVec3(c, "uColor", obj.Color)
		s.Program.SetUniformFloat(c, "uAlpha", 1.0)
		obj.Mesh.Draw(c, s.Program)
	}
}
