//go:build js

package renderer

import (
	"encoding/binary"
	"math"
	"syscall/js"
)

// syscall/js には Go のスライスから直接JSの型付き配列を作るAPIが無いため、
// バイト列(Uint8Array)を経由して目的の型付き配列へ変換する。

// float32ArrayOf はGoの[]float32をJS側のFloat32Arrayに変換する。
// 頂点座標などをGPUへアップロードする際に使う。
func float32ArrayOf(data []float32) js.Value {
	buf := make([]byte, len(data)*4)
	for i, f := range data {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return typedArrayFromBytes(buf, "Float32Array")
}

// uint16ArrayOf はGoの[]uint16をJS側のUint16Arrayに変換する。
// メッシュのインデックスバッファをGPUへアップロードする際に使う。
func uint16ArrayOf(data []uint16) js.Value {
	buf := make([]byte, len(data)*2)
	for i, v := range data {
		binary.LittleEndian.PutUint16(buf[i*2:], v)
	}
	return typedArrayFromBytes(buf, "Uint16Array")
}

func typedArrayFromBytes(buf []byte, jsTypedArrayName string) js.Value {
	uint8Array := js.Global().Get("Uint8Array").New(len(buf))
	js.CopyBytesToJS(uint8Array, buf)
	return js.Global().Get(jsTypedArrayName).New(uint8Array.Get("buffer"))
}
