package main

import (
	"fmt"
	"math/rand"

	"github.com/dennisdijkstra/memprint/shared/events"
)

type LayoutElement struct {
	Content  string
	X, Y     float64
	Rotation float64
	SkewX    float64
	Size     float64
	Opacity  float64
	Effect   string
}

type Layout struct {
	Seed     int64
	Width    int
	Height   int
	Elements []LayoutElement
}

func makeLayout(meta events.MemMetadata) Layout {
	rng := rand.New(rand.NewSource(int64(meta.PID)))

	angle := func(max float64) float64 {
		return (rng.Float64()*2 - 1) * max
	}
	opacity := func(min, max float64) float64 {
		return min + rng.Float64()*(max-min)
	}

	elements := []LayoutElement{
		{
			Content:  fmt.Sprintf("0x%08X", meta.HeapAddr),
			X:        rng.Float64() * 15,
			Y:        80 + rng.Float64()*20,
			Rotation: angle(3),
			SkewX:    angle(8),
			Size:     80 + rng.Float64()*12,
			Opacity:  0.95,
			Effect:   "pool",
		},
		{
			Content:  fmt.Sprintf("PID:%d", meta.PID),
			X:        rng.Float64() * 10,
			Y:        155 + rng.Float64()*15,
			Rotation: angle(5),
			SkewX:    angle(14),
			Size:     48 + rng.Float64()*8,
			Opacity:  opacity(0.85, 0.97),
			Effect:   "warp",
		},
		{
			Content:  fmt.Sprintf("TID:%d", meta.TID),
			X:        rng.Float64() * 10,
			Y:        210 + rng.Float64()*15,
			Rotation: angle(6),
			SkewX:    angle(18),
			Size:     44 + rng.Float64()*8,
			Opacity:  opacity(0.82, 0.95),
			Effect:   "drag",
		},
		{
			Content:  fmt.Sprintf("HEAP·%dB", meta.HeapSize),
			X:        rng.Float64()*20 - 10,
			Y:        310 + rng.Float64()*20,
			Rotation: angle(42),
			SkewX:    angle(10),
			Size:     72 + rng.Float64()*16,
			Opacity:  opacity(0.65, 0.85),
			Effect:   "melt",
		},
		{Content: fmt.Sprintf("NR:%d·MMAP", meta.NRMmap), Y: 370, Size: 28, Rotation: angle(3), Opacity: opacity(0.7, 0.9), Effect: "aberration"},
		{Content: fmt.Sprintf("NR:%d·WRITE", meta.NRWrite), Y: 408, Size: 26, Rotation: angle(4), Opacity: opacity(0.65, 0.85), Effect: "drag"},
		{Content: fmt.Sprintf("NR:%d·FSYNC", meta.NRFsync), Y: 444, Size: 24, Rotation: angle(5), Opacity: opacity(0.6, 0.8), Effect: "melt"},
		{Content: fmt.Sprintf("FD:%d", meta.FD), Y: 478, Size: 18, Rotation: angle(4), Opacity: opacity(0.55, 0.75), Effect: "aberration"},
		{
			Content:  fmt.Sprintf("STACK·%dB", meta.StackOffset),
			X:        rng.Float64() * 20,
			Y:        560 + rng.Float64()*20,
			Rotation: angle(8),
			Size:     58 + rng.Float64()*14,
			Opacity:  opacity(0.08, 0.18),
			Effect:   "ghost",
		},
	}

	return Layout{
		Seed:     int64(meta.PID),
		Width:    400,
		Height:   600,
		Elements: elements,
	}
}
