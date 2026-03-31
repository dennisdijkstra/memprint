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
	rng := rand.New(rand.NewSource(int64(meta.PID))) //#nosec G404 -- math/rand intentional, used for non-security shuffle

	angle := func(max float64) float64 {
		return (rng.Float64()*2 - 1) * max
	}
	opacity := func(min, max float64) float64 {
		return min + rng.Float64()*(max-min)
	}

	elements := []LayoutElement{
		// dominant anchor — hex address, massive, top
		{
			Content:  fmt.Sprintf("0x%08X", meta.HeapAddr),
			X:        -float64(posterWidth)/2 + 5,
			Y:        88,
			Rotation: angle(2),
			SkewX:    angle(4),
			Size:     78,
			Opacity:  0.95,
			Effect:   "pool",
		},
		// _0012_4000 second line
		{
			Content:  fmt.Sprintf("_%04X_%04X", (meta.HeapAddr>>16)&0xFFFF, meta.HeapAddr&0xFFFF),
			X:        -float64(posterWidth)/2 + 5,
			Y:        148,
			Rotation: angle(2),
			SkewX:    angle(3),
			Size:     56,
			Opacity:  0.92,
			Effect:   "warp",
		},
		// PID
		{
			Content:  fmt.Sprintf("PID:%d", meta.PID),
			X:        -float64(posterWidth)/2 + 5,
			Y:        210,
			Rotation: angle(2),
			SkewX:    angle(4),
			Size:     48,
			Opacity:  opacity(0.88, 0.95),
			Effect:   "drag",
		},
		// TID
		{
			Content:  fmt.Sprintf("TID:%d", meta.TID),
			X:        -float64(posterWidth)/2 + 5,
			Y:        262,
			Rotation: angle(2),
			SkewX:    angle(3),
			Size:     40,
			Opacity:  opacity(0.85, 0.92),
			Effect:   "warp",
		},
		// HEAP — massive diagonal
		{
			Content:  "HEAP",
			X:        -float64(posterWidth)/2 - 12,
			Y:        348,
			Rotation: angle(2),
			SkewX:    angle(3),
			Size:     86,
			Opacity:  opacity(0.88, 0.95),
			Effect:   "melt",
		},
		// syscalls
		{
			Content:  fmt.Sprintf("NR:%d·MMAP", meta.NRMmap),
			X:        -float64(posterWidth)/2 + 5,
			Y:        412,
			Rotation: angle(2),
			SkewX:    angle(4),
			Size:     38,
			Opacity:  opacity(0.85, 0.92),
			Effect:   "warp",
		},
		{
			Content:  fmt.Sprintf("NR:%d·WRITE", meta.NRWrite),
			X:        -float64(posterWidth)/2 + 5,
			Y:        458,
			Rotation: angle(2),
			SkewX:    angle(3),
			Size:     36,
			Opacity:  opacity(0.82, 0.9),
			Effect:   "drag",
		},
		{
			Content:  fmt.Sprintf("NR:%d·FSYNC", meta.NRFsync),
			X:        -float64(posterWidth)/2 + 5,
			Y:        504,
			Rotation: angle(2),
			SkewX:    angle(3),
			Size:     34,
			Opacity:  opacity(0.8, 0.88),
			Effect:   "melt",
		},
		{
			Content:  fmt.Sprintf("G#%d·G#%d·G#%d", meta.PID+1, meta.PID+2, meta.PID+3),
			X:        -float64(posterWidth)/2 + 5,
			Y:        544,
			Rotation: angle(1),
			Size:     26,
			Opacity:  opacity(0.78, 0.86),
			Effect:   "pool",
		},
		// ghost bottom
		{
			Content:  fmt.Sprintf("%dB·CHECKSUM·32B", meta.HeapSize),
			X:        -float64(posterWidth)/2 + 5,
			Y:        578,
			Rotation: angle(1),
			Size:     20,
			Opacity:  opacity(0.6, 0.7),
			Effect:   "melt",
		},
	}

	return Layout{
		Seed:     int64(meta.PID),
		Width:    posterWidth,
		Height:   posterHeight,
		Elements: elements,
	}
}
