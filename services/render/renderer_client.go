package main

import (
	"context"
	"fmt"
	"os"

	rendererpb "github.com/dennisdijkstra/memprint/proto/renderer"
	"github.com/dennisdijkstra/memprint/shared/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RendererClient struct {
	client rendererpb.RendererServiceClient
}

func toInt32(name string, v int) (int32, error) {
	const (
		minInt32 = -1 << 31
		maxInt32 = 1<<31 - 1
	)

	if v < minInt32 || v > maxInt32 {
		return 0, fmt.Errorf("%s out of int32 range: %d", name, v)
	}

	return int32(v), nil
}

func newRendererClient() (*RendererClient, error) {
	addr := os.Getenv("RENDERER_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50053"
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect renderer service: %w", err)
	}

	return &RendererClient{
		client: rendererpb.NewRendererServiceClient(conn),
	}, nil
}

func (r *RendererClient) render(ctx context.Context, meta events.MemMetadata) ([]byte, error) {
	pid, err := toInt32("PID", meta.PID)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	tid, err := toInt32("TID", meta.TID)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	fd, err := toInt32("FD", meta.FD)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	nrOpenat, err := toInt32("NROpenat", meta.NROpenat)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	nrMmap, err := toInt32("NRMmap", meta.NRMmap)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	nrWrite, err := toInt32("NRWrite", meta.NRWrite)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	nrFsync, err := toInt32("NRFsync", meta.NRFsync)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	numGoroutines, err := toInt32("NumGoroutines", meta.NumGoroutines)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	numCPU, err := toInt32("NumCPU", meta.NumCPU)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	goMaxProcs, err := toInt32("GoMaxProcs", meta.GoMaxProcs)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	pageSize, err := toInt32("PageSize", meta.PageSize)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	filePages, err := toInt32("FilePages", meta.FilePages)
	if err != nil {
		return nil, fmt.Errorf("build render request: %w", err)
	}

	req := &rendererpb.RenderRequest{
		Pid:      pid,
		Tid:      tid,
		HeapAddr: meta.HeapAddr,
		HeapSize: meta.HeapSize,
		Fd:       fd,
		NrOpenat: nrOpenat,
		NrMmap:   nrMmap,
		NrWrite:  nrWrite,
		NrFsync:  nrFsync,
		Checksum: meta.Checksum,

		NumGoroutines:  numGoroutines,
		NumCpu:         numCPU,
		GoMaxProcs:     goMaxProcs,
		NumGc:          meta.NumGC,
		GcPauseTotalNs: meta.GCPauseTotalNs,
		PageSize:       pageSize,
		FilePages:      filePages,
		FileEntropy:    meta.FileEntropy,
		MagicBytes:     meta.MagicBytes,
	}

	resp, err := r.client.RenderPoster(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("render poster: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("renderer error: %s", resp.Error)
	}
	if len(resp.PngData) == 0 {
		return nil, fmt.Errorf("renderer returned empty PNG")
	}

	return resp.PngData, nil
}
