package main

import (
	"fmt"
	"os"

	"github.com/Kan-O435/okarina/internal/gltf"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	prim, err := gltf.Parse(data)
	if err != nil {
		fmt.Println("parse error:", err)
		os.Exit(1)
	}

	fmt.Printf("OK: vertices=%d indices=%d baseColor=%v texture=%s(%dbytes)\n",
		len(prim.Positions)/3, len(prim.Indices), prim.BaseColor,
		textureSummary(prim.TextureMimeType), len(prim.TextureData))
}

func textureSummary(mimeType string) string {
	if mimeType == "" {
		return "none"
	}
	return mimeType
}
