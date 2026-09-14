package renderer

// このファイルは、手打ちで用意する基本形状(平面・直方体)の頂点データを生成する。
// 地面・道・プールのような「ほぼ平ら」なものはquad、階段・塔・建物プレース
// ホルダーのような「厚みのある」ものはboxを使う。

// zeroUVs は全頂点のUV座標を(0, 0)で埋めた配列を返す。1x1の単色テクスチャ
// (Context.WhiteTexture)はUV値によらず同じ色を返すため、実テクスチャ画像を
// 持たない自作オブジェクト(地面・プレースホルダー等)はこれで問題ない。
func zeroUVs(vertexCount int) []float32 {
	return make([]float32, vertexCount*2)
}

// quadIndices は4頂点(v0..v3)で構成される矩形を三角形2枚として描画するための
// 共通インデックス。quadVertices/linkVertices どちらの頂点順にも対応する。
func quadIndices() []uint16 {
	return []uint16{0, 1, 2, 0, 2, 3}
}

// quadVertices はXZ平面上、高さyに配置した矩形(中心がX/Z原点)の頂点を返す。
// 地面・道・プールなど、水平に置くオブジェクトに使う。
func quadVertices(halfWidth, halfDepth, y float32) []float32 {
	return []float32{
		-halfWidth, y, -halfDepth,
		halfWidth, y, -halfDepth,
		halfWidth, y, halfDepth,
		-halfWidth, y, halfDepth,
	}
}

// boxIndices はboxVerticesの8頂点(v0..v7)から直方体の6面(12三角形)を描画する
// ためのインデックス。
func boxIndices() []uint16 {
	return []uint16{
		0, 1, 2, 0, 2, 3, // 奥面 (-Z)
		5, 4, 7, 5, 7, 6, // 手前面 (+Z)
		4, 0, 3, 4, 3, 7, // 左面 (-X)
		1, 5, 6, 1, 6, 2, // 右面 (+X)
		4, 5, 1, 4, 1, 0, // 底面 (-Y)
		3, 2, 6, 3, 6, 7, // 天面 (+Y)
	}
}

// boxVertices は、底面の中心をX/Z原点・Y=0とし、Y=heightまで積み上がる直方体の
// 頂点を返す。Translate(x, 0, z) だけで地面の上に設置できるよう、
// 底が原点に来る座標系にしている(階段・塔・建物プレースホルダー用)。
func boxVertices(halfWidth, height, halfDepth float32) []float32 {
	x, y, z := halfWidth, height, halfDepth
	return []float32{
		-x, 0, -z, // v0
		x, 0, -z, // v1
		x, y, -z, // v2
		-x, y, -z, // v3
		-x, 0, z, // v4
		x, 0, z, // v5
		x, y, z, // v6
		-x, y, z, // v7
	}
}
