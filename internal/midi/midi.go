// Package midi は、MIDIキーボード(MPK Mini等)からの入力を受け取り、
// ノートイベントの解析・メロディパターンの判定を担当する。
//
// ブラウザ環境では Web MIDI API から渡されるイベントを
// (#3で構築する) JS-Go Bridge 経由で受け取る想定。
package midi
