package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/dennisdijkstra/memprint/shared/events"
)

func captureMetadata(fd uintptr) events.MemMetadata {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	sample := make([]byte, 1)
	heapAddr := uint64(uintptr(unsafe.Pointer(&sample[0]))) //#nosec G103 -- intentional heap address capture for memory metadata

	return events.MemMetadata{
		PID:         os.Getpid(),
		TID:         os.Getpid(),
		HeapAddr:    heapAddr,
		HeapAddrHex: fmt.Sprintf("0x%08X", heapAddr),
		HeapSize:    ms.HeapAlloc,
		StackOffset: ms.StackInuse,
		FD:          int(fd), //#nosec G115 -- fd is a file descriptor, always within int range
		NRMmap:      9,
		NRWrite:     1,
		NRFsync:     74,
		NROpenat:    257,
		CapturedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}
