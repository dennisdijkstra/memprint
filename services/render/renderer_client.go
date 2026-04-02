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
	req := &rendererpb.RenderRequest{
		Pid:        int32(meta.PID),
		Tid:        int32(meta.TID),
		HeapAddr:   meta.HeapAddr,
		HeapSize:   meta.HeapSize,
		Fd:         int32(meta.FD),
		NrOpenat:   int32(meta.NROpenat),
		NrMmap:     int32(meta.NRMmap),
		NrWrite:    int32(meta.NRWrite),
		NrFsync:    int32(meta.NRFsync),
		Checksum:   meta.Checksum,
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
