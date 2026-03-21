package main

import (
	"fmt"
	"testing"

	"github.com/dennisdijkstra/memprint/shared/events"
)

func makeMemMetadata(pid int) events.MemMetadata {
	return events.MemMetadata{
		PID:         pid,
		TID:         pid,
		HeapAddr:    0x00C000124000,
		HeapSize:    1024000,
		StackOffset: 425984,
		FD:          9,
		NRMmap:      9,
		NRWrite:     1,
		NRFsync:     74,
		NROpenat:    257,
	}
}

func TestMakeLayout(t *testing.T) {
	meta := makeMemMetadata(12345)
	layout := makeLayout(meta)

	t.Run("exact values", func(t *testing.T) {
		exact := []struct {
			name   string
			value  any
			expect any
		}{
			{"width is 400", layout.Width, 400},
			{"height is 600", layout.Height, 600},
			{"seed matches PID", layout.Seed, int64(meta.PID)},
			{"first element is heap addr", layout.Elements[0].Content, fmt.Sprintf("0x%08X", meta.HeapAddr)},
		}

		for _, tt := range exact {
			t.Run(tt.name, func(t *testing.T) {
				if tt.value != tt.expect {
					t.Errorf("value=%v expect=%v", tt.value, tt.expect)
				}
			})
		}
	})

	t.Run("layout is valid", func(t *testing.T) {
		present := []struct {
			name string
			ok   bool
		}{
			{"has elements", len(layout.Elements) > 0},
			{"all elements have size", haveSize(layout.Elements)},
			{"all elements have effect", haveEffect(layout.Elements)},
			{"all opacities in range", haveValidOpacity(layout.Elements)},
		}

		for _, tt := range present {
			t.Run(tt.name, func(t *testing.T) {
				if !tt.ok {
					t.Errorf("%s: check failed", tt.name)
				}
			})
		}
	})

	t.Run("determinism", func(t *testing.T) {
		determinism := []struct {
			name string
			ok   bool
		}{
			{
				"same PID produces same layout",
				isMatchingLayout(makeLayout(meta), makeLayout(meta)),
			},
			{
				"different PID produces different layout",
				!isMatchingLayout(makeLayout(meta), makeLayout(makeMemMetadata(99999))),
			},
		}

		for _, tt := range determinism {
			t.Run(tt.name, func(t *testing.T) {
				if !tt.ok {
					t.Errorf("%s: check failed", tt.name)
				}
			})
		}
	})
}

func haveSize(elements []LayoutElement) bool {
	for _, el := range elements {
		if el.Size == 0 {
			return false
		}
	}
	return true
}

func haveEffect(elements []LayoutElement) bool {
	for _, el := range elements {
		if el.Effect == "" {
			return false
		}
	}
	return true
}

func haveValidOpacity(elements []LayoutElement) bool {
	for _, el := range elements {
		if el.Opacity < 0 || el.Opacity > 1 {
			return false
		}
	}
	return true
}

func isMatchingLayout(a, b Layout) bool {
	if len(a.Elements) != len(b.Elements) {
		return false
	}
	for i, el := range a.Elements {
		if el.Content != b.Elements[i].Content {
			return false
		}
		if el.Rotation != b.Elements[i].Rotation {
			return false
		}
		if el.Size != b.Elements[i].Size {
			return false
		}
	}
	return true
}
