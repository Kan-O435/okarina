// Package gltf は、GLB(バイナリglTF)ファイルを最小限パースし、
// Meshy AI等で生成した3DモデルをGoの自作WebGLエンジンへ取り込むための土台を提供する。
//
// 現時点では以下に対応範囲を絞っている(スコープ外は今後拡張する):
//   - Parse()は最初のメッシュの最初のプリミティブのみを読む。ParseSkinned()は
//     スキン付きノードが参照する全プリミティブを1つに結合する(体パーツ分割
//     モデル向け)
//   - POSITION/TEXCOORD_0/インデックスに対応。NORMALはまだ読まない(陰影無し)
//   - マテリアルはbaseColorFactorとbaseColorTexture(GLBに埋め込まれた
//     JPEG/PNG)のみ読む。他のテクスチャ(法線・金属度等)は非対応
//   - 頂点数は65535以下(インデックスをuint16に変換するため)
package gltf

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

const (
	glbMagic          = 0x46546C67 // "glTF"
	glbChunkTypeJSON  = 0x4E4F534A // "JSON"
	glbChunkTypeBIN   = 0x004E4942 // "BIN\x00"
	glbHeaderLength   = 12
	glbChunkHeaderLen = 8
)

const (
	componentTypeUnsignedByte  = 5121
	componentTypeUnsignedShort = 5123
	componentTypeUnsignedInt   = 5125
	componentTypeFloat         = 5126
)

// Primitive はパース結果として取り出した、描画に必要な最小限のデータ。
type Primitive struct {
	Positions       []float32 // (x, y, z) の並び
	TexCoords       []float32 // (u, v) の並び。TEXCOORD_0が無い頂点は(0, 0)で埋める
	Indices         []uint16
	BaseColor       [3]float32 // マテリアルのbaseColorFactor(RGB)。無ければglTF仕様どおり白(1,1,1)
	TextureData     []byte     // baseColorTextureの画像バイナリ(JPEG/PNG)。無ければnil
	TextureMimeType string
	Min, Max        [3]float32 // ローカル座標系でのバウンディングボックス(配置・スケール調整に使う)
}

// document はglTFのJSONチャンクのうち、パースに必要な部分だけを表す。
type document struct {
	Meshes      []mesh       `json:"meshes"`
	Nodes       []node       `json:"nodes"`
	Accessors   []accessor   `json:"accessors"`
	BufferViews []bufferView `json:"bufferViews"`
	Materials   []material   `json:"materials"`
	Textures    []texture    `json:"textures"`
	Images      []image      `json:"images"`
}

// texture はtextures配列の要素。どのimageを参照するかだけを読む。
type texture struct {
	Source *int `json:"source"`
}

// image はimages配列の要素。GLBはバイナリチャンクに埋め込まれた画像を
// bufferView経由で参照するため、それだけ読む(外部URI参照は非対応)。
type image struct {
	BufferView *int   `json:"bufferView"`
	MimeType   string `json:"mimeType"`
}

// textureInfo はmaterial.pbrMetallicRoughness.baseColorTexture等の形。
type textureInfo struct {
	Index int `json:"index"`
}

// node はメッシュとスキン(ボーン割り当て)の対応関係だけを読む。
// KayKit等の人型キャラクターは、体パーツ(頭・胴・腕・脚)ごとに別メッシュへ
// 分かれており、スキン付きのノードだけを集めることで「装備品(武器・盾・
// 兜等の付け替えパーツ)を除いた本体だけ」を機械的に判別できる。
type node struct {
	Mesh *int `json:"mesh"`
	Skin *int `json:"skin"`
}

type mesh struct {
	Primitives []primitive `json:"primitives"`
}

type primitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    *int           `json:"indices"`
	Material   *int           `json:"material"`
}

func (p primitive) positionAccessor() (int, bool) {
	idx, ok := p.Attributes["POSITION"]
	return idx, ok
}

func (p primitive) texCoordAccessor() (int, bool) {
	idx, ok := p.Attributes["TEXCOORD_0"]
	return idx, ok
}

type accessor struct {
	BufferView    *int   `json:"bufferView"`
	ByteOffset    int    `json:"byteOffset"`
	ComponentType int    `json:"componentType"`
	Count         int    `json:"count"`
	Type          string `json:"type"`
}

type bufferView struct {
	Buffer     int `json:"buffer"`
	ByteOffset int `json:"byteOffset"`
	ByteLength int `json:"byteLength"`
}

type material struct {
	PBRMetallicRoughness *pbrMetallicRoughness `json:"pbrMetallicRoughness"`
}

type pbrMetallicRoughness struct {
	BaseColorFactor  []float64    `json:"baseColorFactor"`
	BaseColorTexture *textureInfo `json:"baseColorTexture"`
}

// Parse はGLBバイナリをパースし、最初のメッシュの最初のプリミティブを返す。
func Parse(data []byte) (*Primitive, error) {
	jsonChunk, binChunk, err := splitGLBChunks(data)
	if err != nil {
		return nil, err
	}

	var doc document
	if err := json.Unmarshal(jsonChunk, &doc); err != nil {
		return nil, fmt.Errorf("gltf: failed to parse JSON chunk: %w", err)
	}

	if len(doc.Meshes) == 0 || len(doc.Meshes[0].Primitives) == 0 {
		return nil, errors.New("gltf: no mesh primitives found")
	}
	prim := doc.Meshes[0].Primitives[0]

	positionIdx, ok := prim.positionAccessor()
	if !ok {
		return nil, errors.New("gltf: primitive has no POSITION attribute")
	}
	positions, err := readFloat32Accessor(&doc, binChunk, positionIdx)
	if err != nil {
		return nil, err
	}

	if prim.Indices == nil {
		return nil, errors.New("gltf: unindexed primitives are not supported yet")
	}
	indices, err := readIndexAccessor(&doc, binChunk, *prim.Indices)
	if err != nil {
		return nil, err
	}

	texCoords, err := readTexCoordsOrZero(&doc, binChunk, prim, len(positions)/3)
	if err != nil {
		return nil, err
	}

	mat, err := resolveMaterial(&doc, binChunk, prim.Material)
	if err != nil {
		return nil, err
	}

	min, max := computeBounds(positions)

	return &Primitive{
		Positions:       positions,
		TexCoords:       texCoords,
		Indices:         indices,
		BaseColor:       mat.baseColor,
		TextureData:     mat.textureData,
		TextureMimeType: mat.textureMimeType,
		Min:             min,
		Max:             max,
	}, nil
}

// ParseSkinned はGLBバイナリをパースし、スキン(ボーン)が割り当てられた
// ノードが参照するメッシュのプリミティブをすべて1つに結合して返す。
//
// KayKit等の人型キャラクターアセットは、体パーツ(頭・胴・腕・脚)ごとに
// 別メッシュへ分割されており、Parse()のように「最初のメッシュだけ」読むと
// 体の一部しか表示されない。このパッケージはまだスキニング(ボーンに応じた
// 頂点変形)を実装していないため、アニメーションは反映されず常にエクスポート
// 時の姿勢(バインドポーズ)のまま静止表示になる。武器・盾・兜などスキンの
// 割り当てがない装備品ノードは対象外とし、本体パーツのみを結合する。
func ParseSkinned(data []byte) (*Primitive, error) {
	jsonChunk, binChunk, err := splitGLBChunks(data)
	if err != nil {
		return nil, err
	}

	var doc document
	if err := json.Unmarshal(jsonChunk, &doc); err != nil {
		return nil, fmt.Errorf("gltf: failed to parse JSON chunk: %w", err)
	}

	var meshIndices []int
	for _, n := range doc.Nodes {
		if n.Skin != nil && n.Mesh != nil {
			meshIndices = append(meshIndices, *n.Mesh)
		}
	}
	if len(meshIndices) == 0 {
		return nil, errors.New("gltf: no skinned meshes found")
	}

	var positions, texCoords []float32
	var indices []uint16
	mat := defaultMaterialInfo()
	haveMaterial := false

	for _, meshIdx := range meshIndices {
		if meshIdx < 0 || meshIdx >= len(doc.Meshes) {
			return nil, fmt.Errorf("gltf: mesh index %d out of range", meshIdx)
		}
		for _, prim := range doc.Meshes[meshIdx].Primitives {
			positionIdx, ok := prim.positionAccessor()
			if !ok {
				return nil, errors.New("gltf: primitive has no POSITION attribute")
			}
			pos, err := readFloat32Accessor(&doc, binChunk, positionIdx)
			if err != nil {
				return nil, err
			}

			if prim.Indices == nil {
				return nil, errors.New("gltf: unindexed primitives are not supported yet")
			}
			idx, err := readIndexAccessor(&doc, binChunk, *prim.Indices)
			if err != nil {
				return nil, err
			}

			uv, err := readTexCoordsOrZero(&doc, binChunk, prim, len(pos)/3)
			if err != nil {
				return nil, err
			}

			vertexOffset := len(positions) / 3
			if vertexOffset+len(pos)/3 > 0xFFFF {
				return nil, errors.New("gltf: combined mesh has more than 65535 vertices, not supported yet")
			}
			for _, v := range idx {
				indices = append(indices, v+uint16(vertexOffset))
			}
			positions = append(positions, pos...)
			texCoords = append(texCoords, uv...)

			// 体パーツごとにマテリアルが分かれている場合、最初に見つかったものを
			// 代表として使う(テクスチャアトラスが共有されている想定)。
			if !haveMaterial && prim.Material != nil {
				info, err := resolveMaterial(&doc, binChunk, prim.Material)
				if err != nil {
					return nil, err
				}
				mat = info
				haveMaterial = true
			}
		}
	}

	min, max := computeBounds(positions)
	return &Primitive{
		Positions:       positions,
		TexCoords:       texCoords,
		Indices:         indices,
		BaseColor:       mat.baseColor,
		TextureData:     mat.textureData,
		TextureMimeType: mat.textureMimeType,
		Min:             min,
		Max:             max,
	}, nil
}

// materialInfo は、マテリアルから読み取った色・テクスチャ情報をまとめたもの。
type materialInfo struct {
	baseColor       [3]float32
	textureData     []byte
	textureMimeType string
}

// defaultMaterialInfo はマテリアル未指定の場合のデフォルト値(glTF仕様どおり白)。
func defaultMaterialInfo() materialInfo {
	return materialInfo{baseColor: [3]float32{1, 1, 1}}
}

// resolveMaterial は指定インデックスのマテリアルからbaseColorFactorと
// baseColorTextureの画像バイナリを取り出す。materialIdxがnilならデフォルト値を返す。
func resolveMaterial(doc *document, bin []byte, materialIdx *int) (materialInfo, error) {
	info := defaultMaterialInfo()
	if materialIdx == nil || *materialIdx < 0 || *materialIdx >= len(doc.Materials) {
		return info, nil
	}
	mat := doc.Materials[*materialIdx]
	if mat.PBRMetallicRoughness == nil {
		return info, nil
	}
	if len(mat.PBRMetallicRoughness.BaseColorFactor) >= 3 {
		c := mat.PBRMetallicRoughness.BaseColorFactor
		info.baseColor = [3]float32{float32(c[0]), float32(c[1]), float32(c[2])}
	}
	if bct := mat.PBRMetallicRoughness.BaseColorTexture; bct != nil {
		data, mimeType, err := resolveTextureImage(doc, bin, bct.Index)
		if err != nil {
			return info, err
		}
		info.textureData = data
		info.textureMimeType = mimeType
	}
	return info, nil
}

// resolveTextureImage はtextures[textureIdx]が参照するimageのうち、
// バイナリチャンクに埋め込まれた画像データ(JPEG/PNG)を取り出す。
// 外部URI参照の画像はGLBでは通常使われないため非対応。
func resolveTextureImage(doc *document, bin []byte, textureIdx int) (data []byte, mimeType string, err error) {
	if textureIdx < 0 || textureIdx >= len(doc.Textures) {
		return nil, "", fmt.Errorf("gltf: texture index %d out of range", textureIdx)
	}
	tex := doc.Textures[textureIdx]
	if tex.Source == nil {
		return nil, "", errors.New("gltf: texture has no source image")
	}
	if *tex.Source < 0 || *tex.Source >= len(doc.Images) {
		return nil, "", fmt.Errorf("gltf: image index %d out of range", *tex.Source)
	}
	img := doc.Images[*tex.Source]
	if img.BufferView == nil {
		return nil, "", errors.New("gltf: images referenced by URI are not supported (GLB only)")
	}
	if *img.BufferView < 0 || *img.BufferView >= len(doc.BufferViews) {
		return nil, "", fmt.Errorf("gltf: bufferView index %d out of range", *img.BufferView)
	}
	bv := doc.BufferViews[*img.BufferView]
	start, end := bv.ByteOffset, bv.ByteOffset+bv.ByteLength
	if end > len(bin) {
		return nil, "", errors.New("gltf: image data out of bounds")
	}
	out := make([]byte, end-start)
	copy(out, bin[start:end])
	return out, img.MimeType, nil
}

// readTexCoordsOrZero はTEXCOORD_0があれば読み、無ければvertexCount分の
// (0, 0)で埋める(単色テクスチャに対してはUV値が何であれ同じ色になるため、
// テクスチャ無しオブジェクトはこれで問題ない)。
func readTexCoordsOrZero(doc *document, bin []byte, prim primitive, vertexCount int) ([]float32, error) {
	if idx, ok := prim.texCoordAccessor(); ok {
		return readFloat32Accessor(doc, bin, idx)
	}
	return make([]float32, vertexCount*2), nil
}

// computeBounds は(x, y, z)の並びからaxis-alignedバウンディングボックスを求める。
func computeBounds(positions []float32) (min, max [3]float32) {
	if len(positions) < 3 {
		return min, max
	}
	min = [3]float32{positions[0], positions[1], positions[2]}
	max = min
	for i := 0; i+2 < len(positions); i += 3 {
		for axis := 0; axis < 3; axis++ {
			v := positions[i+axis]
			if v < min[axis] {
				min[axis] = v
			}
			if v > max[axis] {
				max[axis] = v
			}
		}
	}
	return min, max
}

// splitGLBChunks はGLBコンテナのヘッダを検証し、JSON/BINチャンクを取り出す。
func splitGLBChunks(data []byte) (jsonChunk, binChunk []byte, err error) {
	if len(data) < glbHeaderLength {
		return nil, nil, errors.New("gltf: file too small to be a valid GLB")
	}
	if magic := binary.LittleEndian.Uint32(data[0:4]); magic != glbMagic {
		return nil, nil, errors.New("gltf: not a GLB file (invalid magic)")
	}

	offset := glbHeaderLength
	for offset+glbChunkHeaderLen <= len(data) {
		chunkLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		chunkType := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		start := offset + glbChunkHeaderLen
		end := start + chunkLength
		if end > len(data) {
			return nil, nil, errors.New("gltf: chunk length exceeds file size")
		}

		switch chunkType {
		case glbChunkTypeJSON:
			jsonChunk = data[start:end]
		case glbChunkTypeBIN:
			binChunk = data[start:end]
		}
		offset = end
	}

	if jsonChunk == nil {
		return nil, nil, errors.New("gltf: missing JSON chunk")
	}
	return jsonChunk, binChunk, nil
}

func readFloat32Accessor(doc *document, bin []byte, accessorIdx int) ([]float32, error) {
	acc, bv, err := lookupAccessor(doc, accessorIdx)
	if err != nil {
		return nil, err
	}
	if acc.ComponentType != componentTypeFloat {
		return nil, fmt.Errorf("gltf: unsupported component type %d for float accessor", acc.ComponentType)
	}

	componentsPerElement, err := componentCount(acc.Type)
	if err != nil {
		return nil, err
	}

	start := bv.ByteOffset + acc.ByteOffset
	count := acc.Count * componentsPerElement
	out := make([]float32, count)
	for i := 0; i < count; i++ {
		off := start + i*4
		if off+4 > len(bin) {
			return nil, errors.New("gltf: accessor data out of bounds")
		}
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(bin[off : off+4]))
	}
	return out, nil
}

func readIndexAccessor(doc *document, bin []byte, accessorIdx int) ([]uint16, error) {
	acc, bv, err := lookupAccessor(doc, accessorIdx)
	if err != nil {
		return nil, err
	}

	start := bv.ByteOffset + acc.ByteOffset
	out := make([]uint16, acc.Count)
	for i := 0; i < acc.Count; i++ {
		switch acc.ComponentType {
		case componentTypeUnsignedByte:
			if start+i >= len(bin) {
				return nil, errors.New("gltf: index data out of bounds")
			}
			out[i] = uint16(bin[start+i])
		case componentTypeUnsignedShort:
			off := start + i*2
			if off+2 > len(bin) {
				return nil, errors.New("gltf: index data out of bounds")
			}
			out[i] = binary.LittleEndian.Uint16(bin[off : off+2])
		case componentTypeUnsignedInt:
			off := start + i*4
			if off+4 > len(bin) {
				return nil, errors.New("gltf: index data out of bounds")
			}
			v := binary.LittleEndian.Uint32(bin[off : off+4])
			if v > 0xFFFF {
				return nil, errors.New("gltf: mesh has more than 65535 vertices, not supported yet")
			}
			out[i] = uint16(v)
		default:
			return nil, fmt.Errorf("gltf: unsupported index component type %d", acc.ComponentType)
		}
	}
	return out, nil
}

func lookupAccessor(doc *document, idx int) (accessor, bufferView, error) {
	if idx < 0 || idx >= len(doc.Accessors) {
		return accessor{}, bufferView{}, fmt.Errorf("gltf: accessor index %d out of range", idx)
	}
	acc := doc.Accessors[idx]
	if acc.BufferView == nil {
		return accessor{}, bufferView{}, errors.New("gltf: accessor without bufferView is not supported")
	}
	if *acc.BufferView < 0 || *acc.BufferView >= len(doc.BufferViews) {
		return accessor{}, bufferView{}, fmt.Errorf("gltf: bufferView index %d out of range", *acc.BufferView)
	}
	return acc, doc.BufferViews[*acc.BufferView], nil
}

func componentCount(accessorType string) (int, error) {
	switch accessorType {
	case "SCALAR":
		return 1, nil
	case "VEC2":
		return 2, nil
	case "VEC3":
		return 3, nil
	case "VEC4":
		return 4, nil
	default:
		return 0, fmt.Errorf("gltf: unsupported accessor type %q", accessorType)
	}
}
