package main

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/dennisdijkstra/memprint/shared/events"
)

func shannonEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}
	var freq [256]int
	for _, b := range data {
		freq[b]++
	}
	n := float64(len(data))
	var h float64
	for _, c := range freq {
		if c == 0 {
			continue
		}
		p := float64(c) / n
		h -= p * math.Log2(p)
	}
	return h
}

func magicBytesHex(data []byte) string {
	var padded [8]byte
	copy(padded[:], data)
	s := hex.EncodeToString(padded[:])
	return s[:8] + " " + s[8:]
}

func captureMetadata(fd uintptr, content []byte) events.MemMetadata {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	sample := make([]byte, 1)
	heapAddr := uint64(uintptr(unsafe.Pointer(&sample[0]))) //#nosec G103 -- intentional heap address capture for memory metadata

	pageSize := os.Getpagesize()
	filePages := (len(content) + pageSize - 1) / pageSize

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
		Checksum:    crc32.ChecksumIEEE(content),

		NumGoroutines:  runtime.NumGoroutine(),
		NumCPU:         runtime.NumCPU(),
		GoMaxProcs:     runtime.GOMAXPROCS(0),
		NumGC:          ms.NumGC,
		GCPauseTotalNs: ms.PauseTotalNs,
		PageSize:       pageSize,
		FilePages:      filePages,
		FileEntropy:    shannonEntropy(content),
		MagicBytes:     magicBytesHex(content),

		CapturedAt: time.Now().UTC().Format(time.RFC3339),
	}
}
