package main

import (
	"testing"
)

func TestCaptureMetadata(t *testing.T) {
	meta := captureMetadata(9)

	// check exact values — syscall numbers are constants, always known
	t.Run("exact values", func(t *testing.T) {
		exact := []struct {
			name   string
			value  int
			expect int
		}{
			{"FD matches passed value", meta.FD, 9},
			{"NRMmap is 9", meta.NRMmap, 9},
			{"NRWrite is 1", meta.NRWrite, 1},
			{"NRFsync is 74", meta.NRFsync, 74},
			{"NROpenat is 257", meta.NROpenat, 257},
		}

		for _, tt := range exact {
			t.Run(tt.name, func(t *testing.T) {
				if tt.value != tt.expect {
					t.Errorf("value=%d expect=%d", tt.value, tt.expect)
				}
			})
		}
	})

	// check presence — runtime values we can't predict, just verify they were captured
	t.Run("runtime values are captured", func(t *testing.T) {
		present := []struct {
			name string
			ok   bool
		}{
			{"PID is set", meta.PID > 0},
			{"HeapAddr is set", meta.HeapAddr > 0},
			{"HeapSize is set", meta.HeapSize > 0},
			{"CapturedAt is set", meta.CapturedAt != ""},
		}

		for _, tt := range present {
			t.Run(tt.name, func(t *testing.T) {
				if !tt.ok {
					t.Errorf("%s: value was not captured", tt.name)
				}
			})
		}
	})
}
