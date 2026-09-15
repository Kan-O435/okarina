package renderer

import (
	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// songSheetAspect は楽譜画像(2016x780px、ChatGPT生成のバナー)の横:縦比。
// HUDのサイズ計算に使う。
const songSheetAspect = 2016.0 / 780.0

// HUD は、3Dシーンとは独立してスクリーン座標(ピクセル、Y下向き)に固定
// 表示する2Dオーバーレイ(時の歌の楽譜など)を描画するための最小限の仕組み。
type HUD struct {
	Program *Program
	Mesh    *Mesh
	Texture *Texture
}

// BuildSongSheetHUD は「時の歌」の楽譜バナーを、canvas上部中央に配置する
// HUDとして組み立てる。widthPx/heightPxはcanvasの実サイズ
// (Context.CanvasSize())。
func BuildSongSheetHUD(c *Context, widthPx, heightPx int) (*HUD, error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	texture, err := c.NewImageTexture(assets.SongOfTimeSheetTexture, "image/png")
	if err != nil {
		return nil, err
	}

	// スクリーン座標系(Y下向き)のクアッドに対し、V座標を反転させて渡す
	// (画像が上下逆さまに表示されるのを防ぐため)。
	mesh := c.NewMesh(songSheetQuadVertices(widthPx, heightPx), quadUVsCropped(0, 1, 1, 0), quadIndices())

	return &HUD{Program: program, Mesh: mesh, Texture: texture}, nil
}

// songSheetQuadVertices は、canvas上部中央に楽譜バナーを配置するための
// スクリーン座標(原点は左上、Y下向き、ピクセル単位)の頂点を返す。
// 頂点順は他のquad系関数と同じく「左下・右下・右上・左上」。
func songSheetQuadVertices(canvasWidth, canvasHeight int) []float32 {
	const widthRatio = 0.4 // canvas幅に対する楽譜の横幅の割合
	const marginTop = 12   // canvas上端からの余白(px)

	width := float32(canvasWidth) * widthRatio
	height := width / songSheetAspect
	x0 := (float32(canvasWidth) - width) / 2
	x1 := x0 + width
	y0 := float32(marginTop)
	y1 := y0 + height

	return []float32{
		x0, y1, 0,
		x1, y1, 0,
		x1, y0, 0,
		x0, y0, 0,
	}
}

// Render はHUDを画面座標にオーバーレイ描画する。深度テストを一時的に
// 無効化することで、常に3Dシーンより手前に表示されるようにする。
// 呼び出し側(cmd/game/main.go)が、3Dシーンを描画した後に毎フレーム呼ぶ。
func (h *HUD) Render(c *Context, canvasWidth, canvasHeight int) {
	c.DisableDepthTest()

	ortho := vecmath.Ortho(0, float64(canvasWidth), float64(canvasHeight), 0, -1, 1)

	h.Program.Use(c)
	h.Program.SetUniformSampler(c, "uTexture", 0)
	h.Texture.Bind(c)
	h.Program.SetUniformMat4(c, "uMVP", ortho)
	h.Program.SetUniformVec3(c, "uColor", vecmath.NewVec3(1, 1, 1))
	h.Mesh.Draw(c, h.Program)

	c.EnableDepthTest()
}
