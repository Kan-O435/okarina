package renderer

import "github.com/Kan-O435/okarina/internal/vecmath"

// WhiteFadeOverlay は、HUDと同じくスクリーン座標に固定表示する2D
// オーバーレイだが、HUD(常に不透明なバナー画像)とは異なり、キャンバス
// 全体を覆う単色(白)のクアッドを、外部から指定した不透明度(alpha)で
// 毎フレーム描画し直せる点が異なる。ganon-battleフィールドのGanon最終
// 形態撃破カットシーン(internal/renderer/ganon.go)で、雷が落ちた瞬間の
// 一瞬の閃光と、演出の最後に画面を白く覆ってからページ遷移する場面の
// 両方で使い回す。
type WhiteFadeOverlay struct {
	Program *Program
	Mesh    *Mesh
}

// BuildWhiteFadeOverlay は、キャンバス全体を覆う白いクアッドを1つ組み立てる。
// widthPx/heightPxはcanvasの実サイズ(Context.CanvasSize())。
func BuildWhiteFadeOverlay(c *Context, widthPx, heightPx int) (*WhiteFadeOverlay, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	mesh := c.NewMesh(fullscreenQuadVertices(widthPx, heightPx), quadUVsCropped(0, 1, 1, 0), quadIndices())

	return &WhiteFadeOverlay{Program: program, Mesh: mesh}, nil
}

// fullscreenQuadVertices は、hudQuadVertices(バナー用、canvas幅の一部だけ)
// とは異なり、キャンバス全体(0,0)〜(canvasWidth,canvasHeight)を覆う
// スクリーン座標(原点は左上、Y下向き、ピクセル単位)の頂点を返す。
// 頂点順は他のquad系関数と同じく「左下・右下・右上・左上」。
func fullscreenQuadVertices(canvasWidth, canvasHeight int) []float32 {
	w := float32(canvasWidth)
	h := float32(canvasHeight)
	return []float32{
		0, h, 0,
		w, h, 0,
		w, 0, 0,
		0, 0, 0,
	}
}

// Render はキャンバス全体を白でalpha(0〜1)の不透明度で覆う。深度テストを
// 一時的に無効化し、常に3Dシーンより手前に表示されるようにする。呼び出し側
// (cmd/ganon-battle/main.go)が、3Dシーンを描画した後に毎フレーム呼ぶ。
// alpha<=0の間は呼ばなくてよい(完全に透明で描画する意味が無いため)。
func (o *WhiteFadeOverlay) Render(c *Context, canvasWidth, canvasHeight int, alpha float64) {
	c.DisableDepthTest()

	ortho := vecmath.Ortho(0, float64(canvasWidth), float64(canvasHeight), 0, -1, 1)

	o.Program.Use(c)
	o.Program.SetUniformSampler(c, "uTexture", 0)
	c.WhiteTexture().Bind(c)
	o.Program.SetUniformMat4(c, "uMVP", ortho)
	o.Program.SetUniformVec3(c, "uColor", vecmath.NewVec3(1, 1, 1))
	o.Program.SetUniformFloat(c, "uAlpha", alpha)
	o.Mesh.Draw(c, o.Program)

	c.EnableDepthTest()
}
