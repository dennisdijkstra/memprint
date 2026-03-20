package main

import (
	"testing"
)

func TestCaptureMetadata(t *testing.T) {
	meta := captureMetadata(9)

	tests := []struct {
		name string
		got  any
		want any
		zero bool // true = just check it's not zero
	}{
		{name: "PID is set", got: meta.PID, zero: true},
		{name: "HeapAddr is set", got: meta.HeapAddr, zero: true},
		{name: "HeapSize is set", got: meta.HeapSize, zero: true},
		{name: "CapturedAt is set", got: meta.CapturedAt, zero: true},
		{name: "FD matches", got: meta.FD, want: 9},
		{name: "NRMmap is 9", got: meta.NRMmap, want: 9},
		{name: "NRWrite is 1", got: meta.NRWrite, want: 1},
		{name: "NRFsync is 74", got: meta.NRFsync, want: 74},
		{name: "NROpenat is 257", got: meta.NROpenat, want: 257},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.zero {
				switch v := tt.got.(type) {
				case int:
					if v == 0 {
						t.Errorf("%s: expected non-zero value", tt.name)
					}
				case uint64:
					if v == 0 {
						t.Errorf("%s: expected non-zero value", tt.name)
					}
				case string:
					if v == "" {
						t.Errorf("%s: expected non-empty string", tt.name)
					}
				}
				return
			}

			if tt.got != tt.want {
				t.Errorf("%s: got=%v want=%v", tt.name, tt.got, tt.want)
			}
		})
	}
}
