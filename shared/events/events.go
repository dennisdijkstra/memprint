package events

import "fmt"

const (
	QueueFileUploaded = "file.uploaded"
	QueuePosterReady  = "poster.ready"
	QueueRenderFailed = "render.failed"
)

type FileUploadedEvent struct {
	FileID    string      `json:"file_id"`
	UserID    string      `json:"user_id"`
	Filename  string      `json:"filename"`
	Meta      MemMetadata `json:"meta"`
	Timestamp string      `json:"timestamp"`
}

type PosterReadyEvent struct {
	FileID    string `json:"file_id"`
	UserID    string `json:"user_id"`
	JobID     string `json:"job_id"`
	PosterURL string `json:"poster_url"`
	Timestamp string `json:"timestamp"`
}

type RenderFailedEvent struct {
	FileID    string `json:"file_id"`
	UserID    string `json:"user_id"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp"`
}

type MemMetadata struct {
	PID         int    `json:"PID"`
	TID         int    `json:"TID"`
	HeapAddr    uint64 `json:"HeapAddr"`
	HeapAddrHex string `json:"heap_addr_hex"`
	HeapSize    uint64 `json:"HeapSize"`
	StackOffset uint64 `json:"StackOffset"`
	FD          int    `json:"FD"`
	NRMmap      int    `json:"NRMmap"`
	NRWrite     int    `json:"NRWrite"`
	NRFsync     int    `json:"NRFsync"`
	NROpenat    int    `json:"NROpenat"`
	Checksum    uint32 `json:"checksum"`

	NumGoroutines  int     `json:"num_goroutines"`
	NumCPU         int     `json:"num_cpu"`
	GoMaxProcs     int     `json:"go_max_procs"`
	NumGC          uint32  `json:"num_gc"`
	GCPauseTotalNs uint64  `json:"gc_pause_total_ns"`
	PageSize       int     `json:"page_size"`
	FilePages      int     `json:"file_pages"`
	FileEntropy    float64 `json:"file_entropy"`
	MagicBytes     string  `json:"magic_bytes"`

	CapturedAt  string `json:"CapturedAt"`
}

func (m MemMetadata) String() string {
	return fmt.Sprintf(
		"PID:%d TID:%d Heap:0x%08X HeapSize:%dB Stack:%dB FD:%d",
		m.PID, m.TID, m.HeapAddr, m.HeapSize, m.StackOffset, m.FD,
	)
}
