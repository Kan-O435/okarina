package renderer

import (
	"github.com/Kan-O435/okarina/internal/assets"
	"github.com/Kan-O435/okarina/internal/vecmath"
)

// このファイルは、タイトル画面の背景でくるくる回るオカリナのSceneを
// 組み立てる。ロゴ(HTML側のimg要素)がcanvasの手前に重なって表示される
// ため、この3Dシーンはロゴの背景として見える。

// titleOcarinaHeight は、オカリナ(internal/assets.TitleOcarina)をこの
// 高さになるようスケールする。titleCameraDistanceはカメラからオカリナ
// 中心までの距離。
const (
	titleOcarinaHeight  = 3.5
	titleCameraDistance = 6.0
)

// BuildTitleScene は、タイトル画面のSceneを組み立てる。戻り値の
// localTransformは、ワールド位置・回転を含まないモデル原点補正+スケール
// のみの変換で、呼び出し側(cmd/title/main.go)が毎フレーム回転角を進めて
// Transformを再構成するために使う(Scene.Objectsの全要素が同じオカリナの
// パーツなので、共通のlocalTransformを使い回せる)。ocarinaPartsは、読み
// 込んだオカリナのパーツ(Mesh・Textureを含む)をそのまま返したもので、
// 呼び出し側が同じMesh/Textureを使い回して背景に小さなオカリナを
// 大量に追加する(BuildTitleBackgroundOcarinas参照)ためのもの。
func BuildTitleScene(c *Context) (scene *Scene, localTransform vecmath.Mat4, ocarinaParts []*Model, err error) {
	program, err := c.NewProgram(basicVertexShaderSrc, basicFragmentShaderSrc)
	if err != nil {
		return nil, vecmath.Mat4{}, nil, err
	}

	width, height := c.CanvasSize()
	aspect := float64(width) / float64(height)
	projection := vecmath.Perspective(vecmath.Radians(45), aspect, 0.1, 100)
	eyeY := titleOcarinaHeight / 2
	view := vecmath.LookAt(
		vecmath.NewVec3(0, eyeY, titleCameraDistance),
		vecmath.NewVec3(0, eyeY, 0),
		vecmath.NewVec3(0, 1, 0),
	)

	parts, err := c.LoadGLBParts(assets.TitleOcarina)
	if err != nil {
		return nil, vecmath.Mat4{}, nil, err
	}
	// x=0, z=0で求めることで、ワールド位置を含まない「ローカル」変換
	// (モデル原点補正+スケールのみ)になる(Model.GroundTransform参照)。
	localTransform = CombinedGroundTransform(parts, 0, 0, titleOcarinaHeight)

	objects := make([]Object, len(parts))
	for i, p := range parts {
		objects[i] = Object{Mesh: p.Mesh, Texture: p.Texture, Transform: localTransform, Color: p.Color}
	}

	return &Scene{
		Program:        program,
		ViewProjection: projection.Mul(view),
		Objects:        objects,
	}, localTransform, parts, nil
}

// BuildTitleBackgroundOcarinaObjects は、ocarinaParts(BuildTitleSceneが
// 読み込んだメインのオカリナと同じMesh/Texture)を再利用して、指定した
// 高さの小さなオカリナ1体分のObjectを組み立てる。Mesh/Textureを共有する
// ため、GLBの再読み込み(テクスチャの再アップロード等)は発生しない。
// 戻り値のlocalTransformは、メインのオカリナと同様にワールド位置・回転を
// 含まない変換で、呼び出し側が毎フレーム位置・回転を組み合わせて使う。
func BuildTitleBackgroundOcarinaObjects(ocarinaParts []*Model, height float64) (objects []Object, localTransform vecmath.Mat4) {
	localTransform = CombinedGroundTransform(ocarinaParts, 0, 0, height)
	objects = make([]Object, len(ocarinaParts))
	for i, p := range ocarinaParts {
		objects[i] = Object{Mesh: p.Mesh, Texture: p.Texture, Transform: localTransform, Color: p.Color}
	}
	return objects, localTransform
}
