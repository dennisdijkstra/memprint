package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"
)

type MemMetadata struct {
	PID         int
	TID         int
	HeapAddr    uint64
	HeapSize    uint64
	StackOffset uint64
	Checksum    uint32
	FD          int
	NRMmap      int
	NRWrite     int
	NRFsync     int
	NROpenat    int
	CapturedAt  string
}

func captureMetadata(fd uintptr) MemMetadata {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	sample := make([]byte, 1)
	heapAddr := uint64(uintptr(unsafe.Pointer(&sample[0])))

	return MemMetadata{
		PID:         os.Getpid(),
		TID:         os.Getpid(),
		HeapAddr:    heapAddr,
		HeapSize:    ms.HeapAlloc,
		StackOffset: ms.StackInuse,
		FD:          int(fd),
		NRMmap:      9,
		NRWrite:     1,
		NRFsync:     74,
		NROpenat:    257,
		CapturedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}

func (m MemMetadata) String() string {
	return fmt.Sprintf(
		"PID:%d TID:%d Heap:0x%08X HeapSize:%dB Stack:%dB FD:%d",
		m.PID, m.TID, m.HeapAddr, m.HeapSize, m.StackOffset, m.FD,
	)
}
