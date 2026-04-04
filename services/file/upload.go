package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	filepb "github.com/dennisdijkstra/memprint/proto/file"
	"github.com/dennisdijkstra/memprint/shared/events"
)

func (s *FileServer) UploadFile(ctx context.Context, req *filepb.UploadFileRequest) (*filepb.UploadFileResponse, error) {
	log.Printf("received upload: user=%s, filename=%s, size=%d bytes", req.UserId, req.Filename, len(req.Content))

	fileID := fmt.Sprintf("file_%d", time.Now().UnixNano())

	f, err := os.CreateTemp("", "upload-*.bin")
	if err != nil {
		return nil, fmt.Errorf("create tmp file: %w", err)
	}
	tmpPath := f.Name()

	meta := captureMetadata(f.Fd(), req.Content)

	if _, err = f.Write(req.Content); err != nil {
		return nil, fmt.Errorf("write tmp file: %w", err)
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("sync tmp file: %w", err)
	}
	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("close tmp file: %w", err)
	}
	if err = os.Remove(tmpPath); err != nil {
		log.Printf("remove tmp file %s: %v", tmpPath, err)
	}

	log.Printf("metadata captured: %s", meta)

	_, err = s.db.Exec(ctx, `
		INSERT INTO files (id, user_id, filename, status)
        VALUES ($1, $2, $3, 'uploaded')
	`, fileID, req.UserId, req.Filename)
	if err != nil {
		return nil, fmt.Errorf("insert file: %w", err)
	}

	log.Printf("saved file: id=%s, user=%s, filename=%s", fileID, req.UserId, req.Filename)

	_, err = s.db.Exec(ctx, `
		INSERT INTO mem_metadata (
			file_id, pid, tid, heap_addr, heap_size,
			stack_offset, fd, nr_mmap, nr_write,
			nr_fsync, nr_openat, captured_at,
			num_goroutines, num_cpu, go_max_procs,
			num_gc, gc_pause_total_ns,
			page_size, file_pages,
			file_entropy, magic_bytes
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17,
			$18, $19,
			$20, $21
		)
	`, fileID,
		meta.PID, meta.TID,
		meta.HeapAddr, meta.HeapSize,
		meta.StackOffset, meta.FD,
		meta.NRMmap, meta.NRWrite,
		meta.NRFsync, meta.NROpenat,
		meta.CapturedAt,
		meta.NumGoroutines, meta.NumCPU, meta.GoMaxProcs,
		meta.NumGC, meta.GCPauseTotalNs,
		meta.PageSize, meta.FilePages,
		meta.FileEntropy, meta.MagicBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("insert metadata: %w", err)
	}

	log.Printf("saved metadata for file: %s", fileID)

	event := events.FileUploadedEvent{
		FileID:    fileID,
		UserID:    req.UserId,
		Filename:  req.Filename,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if err := s.mq.publish(ctx, event); err != nil {
		return nil, fmt.Errorf("publish event: %w", err)
	}

	log.Printf("published file.uploaded event for %s", fileID)

	return &filepb.UploadFileResponse{
		FileId: fileID,
		Status: "uploaded",
	}, nil
}
