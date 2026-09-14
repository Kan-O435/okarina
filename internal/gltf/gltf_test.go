package gltf

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"
)

// buildTestGLB は、外部ファイルやネットワークに頼らず検証するため、
// 三角形1枚だけのGLBバイナリをその場で組み立てる。
func buildTestGLB(t *testing.T, withMaterial bool) []byte {
	t.Helper()

	positions := []float32{
		0, 0, 0,
		1, 0, 0,
		0, 1, 0,
	}
	indices := []uint16{0, 1, 2}

	posBytes := make([]byte, len(positions)*4)
	for i, f := range positions {
		binary.LittleEndian.PutUint32(posBytes[i*4:], math.Float32bits(f))
	}
	idxBytes := make([]byte, len(indices)*2)
	for i, v := range indices {
		binary.LittleEndian.PutUint16(idxBytes[i*2:], v)
	}

	bin := append(append([]byte{}, posBytes...), idxBytes...)
	for len(bin)%4 != 0 {
		bin = append(bin, 0)
	}

	indicesAccessorIdx := 1
	prim := primitive{
		Attributes: map[string]int{"POSITION": 0},
		Indices:    &indicesAccessorIdx,
	}
	if withMaterial {
		materialIdx := 0
		prim.Material = &materialIdx
	}

	bv0, bv1 := 0, 1
	doc := document{
		Meshes: []mesh{{Primitives: []primitive{prim}}},
		Accessors: []accessor{
			{BufferView: &bv0, ComponentType: componentTypeFloat, Count: 3, Type: "VEC3"},
			{BufferView: &bv1, ComponentType: componentTypeUnsignedShort, Count: 3, Type: "SCALAR"},
		},
		BufferViews: []bufferView{
			{Buffer: 0, ByteOffset: 0, ByteLength: len(posBytes)},
			{Buffer: 0, ByteOffset: len(posBytes), ByteLength: len(idxBytes)},
		},
	}
	if withMaterial {
		doc.Materials = []material{
			{PBRMetallicRoughness: &pbrMetallicRoughness{BaseColorFactor: []float64{0.1, 0.2, 0.3, 1.0}}},
		}
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("failed to marshal test document: %v", err)
	}
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ')
	}

	var buf bytes.Buffer
	writeUint32 := func(v uint32) {
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			t.Fatalf("failed to write uint32: %v", err)
		}
	}

	totalLength := uint32(glbHeaderLength + glbChunkHeaderLen + len(jsonBytes) + glbChunkHeaderLen + len(bin))

	writeUint32(glbMagic)
	writeUint32(2) // version
	writeUint32(totalLength)

	writeUint32(uint32(len(jsonBytes)))
	writeUint32(glbChunkTypeJSON)
	buf.Write(jsonBytes)

	writeUint32(uint32(len(bin)))
	writeUint32(glbChunkTypeBIN)
	buf.Write(bin)

	return buf.Bytes()
}

// buildTestGLBWithTexture は、TEXCOORD_0とbaseColorTexture(埋め込み画像)を
// 持つ三角形1枚だけのGLBバイナリを組み立てる。画像データは本物のJPEG/PNGでは
// なく、バイトがそのまま取り出せることだけを検証するためのダミーバイト列。
func buildTestGLBWithTexture(t *testing.T) []byte {
	t.Helper()

	positions := []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}
	texCoords := []float32{0, 0, 1, 0, 0, 1}
	indices := []uint16{0, 1, 2}
	imageBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x01, 0x02, 0x03} // ダミー(本物のJPEGではない)

	posBytes := make([]byte, len(positions)*4)
	for i, f := range positions {
		binary.LittleEndian.PutUint32(posBytes[i*4:], math.Float32bits(f))
	}
	uvBytes := make([]byte, len(texCoords)*4)
	for i, f := range texCoords {
		binary.LittleEndian.PutUint32(uvBytes[i*4:], math.Float32bits(f))
	}
	idxBytes := make([]byte, len(indices)*2)
	for i, v := range indices {
		binary.LittleEndian.PutUint16(idxBytes[i*2:], v)
	}

	bin := append(append([]byte{}, posBytes...), uvBytes...)
	bin = append(bin, idxBytes...)
	bin = append(bin, imageBytes...)
	for len(bin)%4 != 0 {
		bin = append(bin, 0)
	}

	indicesAccessorIdx := 2
	materialIdx := 0
	textureIdx := 0
	imageIdx := 0

	doc := document{
		Meshes: []mesh{{Primitives: []primitive{{
			Attributes: map[string]int{"POSITION": 0, "TEXCOORD_0": 1},
			Indices:    &indicesAccessorIdx,
			Material:   &materialIdx,
		}}}},
		Accessors: []accessor{
			{BufferView: intPtr(0), ComponentType: componentTypeFloat, Count: 3, Type: "VEC3"},
			{BufferView: intPtr(1), ComponentType: componentTypeFloat, Count: 3, Type: "VEC2"},
			{BufferView: intPtr(2), ComponentType: componentTypeUnsignedShort, Count: 3, Type: "SCALAR"},
		},
		BufferViews: []bufferView{
			{Buffer: 0, ByteOffset: 0, ByteLength: len(posBytes)},
			{Buffer: 0, ByteOffset: len(posBytes), ByteLength: len(uvBytes)},
			{Buffer: 0, ByteOffset: len(posBytes) + len(uvBytes), ByteLength: len(idxBytes)},
			{Buffer: 0, ByteOffset: len(posBytes) + len(uvBytes) + len(idxBytes), ByteLength: len(imageBytes)},
		},
		Materials: []material{{
			PBRMetallicRoughness: &pbrMetallicRoughness{
				BaseColorTexture: &textureInfo{Index: textureIdx},
			},
		}},
		Textures: []texture{{Source: &imageIdx}},
		Images:   []image{{BufferView: intPtr(3), MimeType: "image/jpeg"}},
	}

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("failed to marshal test document: %v", err)
	}
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ')
	}

	var buf bytes.Buffer
	writeUint32 := func(v uint32) {
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			t.Fatalf("failed to write uint32: %v", err)
		}
	}

	totalLength := uint32(glbHeaderLength + glbChunkHeaderLen + len(jsonBytes) + glbChunkHeaderLen + len(bin))

	writeUint32(glbMagic)
	writeUint32(2)
	writeUint32(totalLength)

	writeUint32(uint32(len(jsonBytes)))
	writeUint32(glbChunkTypeJSON)
	buf.Write(jsonBytes)

	writeUint32(uint32(len(bin)))
	writeUint32(glbChunkTypeBIN)
	buf.Write(bin)

	return buf.Bytes()
}

func intPtr(i int) *int { return &i }

func TestParseWithTexture(t *testing.T) {
	data := buildTestGLBWithTexture(t)

	prim, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	wantTexCoords := []float32{0, 0, 1, 0, 0, 1}
	if len(prim.TexCoords) != len(wantTexCoords) {
		t.Fatalf("TexCoords length = %d, want %d", len(prim.TexCoords), len(wantTexCoords))
	}
	for i, v := range wantTexCoords {
		if prim.TexCoords[i] != v {
			t.Errorf("TexCoords[%d] = %v, want %v", i, prim.TexCoords[i], v)
		}
	}

	wantImage := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x01, 0x02, 0x03}
	if !bytes.Equal(prim.TextureData, wantImage) {
		t.Errorf("TextureData = %v, want %v", prim.TextureData, wantImage)
	}
	if prim.TextureMimeType != "image/jpeg" {
		t.Errorf("TextureMimeType = %q, want %q", prim.TextureMimeType, "image/jpeg")
	}
}

func TestParseMinimalTriangle(t *testing.T) {
	data := buildTestGLB(t, false)

	prim, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	wantPositions := []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}
	if len(prim.Positions) != len(wantPositions) {
		t.Fatalf("Positions length = %d, want %d", len(prim.Positions), len(wantPositions))
	}
	for i, v := range wantPositions {
		if prim.Positions[i] != v {
			t.Errorf("Positions[%d] = %v, want %v", i, prim.Positions[i], v)
		}
	}

	wantIndices := []uint16{0, 1, 2}
	if len(prim.Indices) != len(wantIndices) {
		t.Fatalf("Indices length = %d, want %d", len(prim.Indices), len(wantIndices))
	}
	for i, v := range wantIndices {
		if prim.Indices[i] != v {
			t.Errorf("Indices[%d] = %v, want %v", i, prim.Indices[i], v)
		}
	}

	// マテリアル無しの場合はglTF仕様どおりデフォルトの白になる。
	if prim.BaseColor != [3]float32{1, 1, 1} {
		t.Errorf("BaseColor = %v, want default white", prim.BaseColor)
	}

	// TEXCOORD_0が無い場合は(0, 0)で埋める。
	wantTexCoords := []float32{0, 0, 0, 0, 0, 0}
	if len(prim.TexCoords) != len(wantTexCoords) {
		t.Fatalf("TexCoords length = %d, want %d", len(prim.TexCoords), len(wantTexCoords))
	}
	for i, v := range wantTexCoords {
		if prim.TexCoords[i] != v {
			t.Errorf("TexCoords[%d] = %v, want %v", i, prim.TexCoords[i], v)
		}
	}
	if prim.TextureData != nil {
		t.Errorf("TextureData = %v, want nil (no texture in test fixture)", prim.TextureData)
	}
}

func TestParseWithMaterialBaseColor(t *testing.T) {
	data := buildTestGLB(t, true)

	prim, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := [3]float32{0.1, 0.2, 0.3}
	if prim.BaseColor != want {
		t.Errorf("BaseColor = %v, want %v", prim.BaseColor, want)
	}
}

func TestParseRejectsInvalidMagic(t *testing.T) {
	_, err := Parse([]byte("not a glb file"))
	if err == nil {
		t.Fatal("Parse() error = nil, want error for invalid magic")
	}
}
