package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/dennisdijkstra/memprint/shared/events"
	"github.com/fogleman/gg"
)

// node spacing — evenly distributed across 600px
const (
	nodeX      = 90.0  // left edge of most nodes
	nodeW      = 220.0 // width of most nodes
	nodeH      = 36.0  // height of each node
	nodeStartY = 32.0  // first node Y
	nodeGap    = 64.0  // vertical spacing between nodes
)

func drawJourneyDiagram(dc *gg.Context, meta events.MemMetadata) {
	dc.SetColor(color.RGBA{17, 17, 17, 255})
	dc.SetLineWidth(1.2)

	// load a small font for diagram labels
	loadFont(dc, 7)

	// header
	drawDiagramHeader(dc)

	// nodes — Y positions spread evenly across full page height
	y := nodeStartY

	// NODE 1: curl / browser
	drawNode(dc, nodeX, y, nodeW, nodeH,
		"curl · browser",
		"multipart/form-data",
	)
	drawArrow(dc, 200, y+nodeH, 200, y+nodeH+nodeGap-nodeH, "HTTP POST")
	y += nodeGap

	// NODE 2: NIC
	drawNode(dc, nodeX, y, nodeW, nodeH,
		"NIC · TCP/IP",
		"packets reassembled",
	)
	drawArrow(dc, 200, y+nodeH, 200, y+nodeH+nodeGap-nodeH, "kernel recv")
	y += nodeGap

	// NODE 3: socket buffer
	drawNode(dc, nodeX, y, nodeW, nodeH,
		"socket buffer",
		"sk_buff · kernel space",
	)

	// split arrows to openat + mmap
	midY := y + nodeH
	splitY := midY + (nodeGap-nodeH)/2
	dc.SetLineWidth(0.9)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.45)
	dc.DrawLine(170, midY, 90, splitY)
	dc.Stroke()
	dc.DrawLine(230, midY, 310, splitY)
	dc.Stroke()
	drawArrowHead(dc, 90, splitY, math.Pi*0.65)
	drawArrowHead(dc, 310, splitY, math.Pi*0.35)
	drawSmallLabel(dc, 52, midY+12, "openat")
	drawSmallLabel(dc, 316, midY+12, "mmap")
	y += nodeGap

	// NODE 4a: openat — left side
	dc.SetLineWidth(1.2)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.58)
	drawNodeAt(dc, 12, y, 130, nodeH,
		fmt.Sprintf("openat · NR:%d", meta.NROpenat),
		fmt.Sprintf("fd:%d assigned", meta.FD),
	)

	// NODE 4b: mmap — right side
	drawNodeAt(dc, 258, y, 130, nodeH,
		fmt.Sprintf("mmap · NR:%d", meta.NRMmap),
		"virtual memory map",
	)

	// converge arrows
	dc.SetLineWidth(0.9)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.44)
	convergeY := y + nodeH + (nodeGap-nodeH)/2
	dc.DrawLine(77, y+nodeH, 160, convergeY)
	dc.Stroke()
	dc.DrawLine(323, y+nodeH, 240, convergeY)
	dc.Stroke()
	drawArrowHead(dc, 165, convergeY, math.Pi*0.4)
	drawArrowHead(dc, 235, convergeY, math.Pi*0.6)
	y += nodeGap

	// NODE 5: HEAP / RAM — wider, taller, most important
	dc.SetLineWidth(1.4)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.62)
	drawNodeAt(dc, 80, y, 240, nodeH+12,
		fmt.Sprintf("HEAP · RAM"),
		fmt.Sprintf("0x%08X · %dB", meta.HeapAddr, meta.HeapSize),
	)
	// bit squares inside
	drawBitRow(dc, 110, y+nodeH+2, meta.Checksum)
	drawArrow(dc, 200, y+nodeH+12, 200, y+nodeH+12+nodeGap-nodeH-12,
		fmt.Sprintf("write NR:%d", meta.NRWrite))
	y += nodeGap + 4

	// NODE 6: write
	dc.SetLineWidth(1.2)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.56)
	drawNode(dc, nodeX, y, nodeW, nodeH,
		fmt.Sprintf("write · NR:%d", meta.NRWrite),
		"copy to kernel buffer",
	)
	drawArrow(dc, 200, y+nodeH, 200, y+nodeH+nodeGap-nodeH,
		fmt.Sprintf("fsync NR:%d", meta.NRFsync))
	y += nodeGap

	// NODE 7: fsync
	dc.SetRGBA(0.07, 0.07, 0.07, 0.54)
	drawNode(dc, nodeX, y, nodeW, nodeH,
		fmt.Sprintf("fsync · NR:%d", meta.NRFsync),
		"flush to disk · durability",
	)
	drawArrow(dc, 200, y+nodeH, 200, y+nodeH+nodeGap-nodeH, "")
	y += nodeGap

	// NODE 8: filesystem
	dc.SetRGBA(0.07, 0.07, 0.07, 0.52)
	drawNode(dc, nodeX, y, nodeW, nodeH,
		"filesystem · inode",
		fmt.Sprintf("/tmp/upload_%d.bin", meta.PID),
	)
	drawArrow(dc, 200, y+nodeH, 200, y+nodeH+nodeGap-nodeH, "")
	y += nodeGap

	// NODE 9: S3 — tallest, final destination
	dc.SetLineWidth(1.3)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.58)
	drawNodeAt(dc, nodeX, y, nodeW, nodeH+12,
		"AWS S3",
		"object stored · permanent",
	)
	loadFont(dc, 5)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.28)
	dc.DrawStringAnchored(
		fmt.Sprintf("posters/file_%d.png", meta.PID),
		200, y+nodeH+4, 0.5, 0.5,
	)

	// footer rule
	dc.SetLineWidth(0.5)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.2)
	dc.DrawLine(20, 578, 380, 578)
	dc.Stroke()

	loadFont(dc, 5)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.25)
	dc.DrawStringAnchored(
		fmt.Sprintf("PID:%d · FD:%d · NR:%d · NR:%d · NR:%d · NR:%d · HEAP:0x%08X",
			meta.PID, meta.FD, meta.NROpenat, meta.NRMmap,
			meta.NRWrite, meta.NRFsync, meta.HeapAddr),
		200, 590, 0.5, 0.5,
	)
}

func drawDiagramHeader(dc *gg.Context) {
	loadFont(dc, 6)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.55)
	dc.DrawStringAnchored("FILE UPLOAD · JOURNEY DIAGRAM", 200, 18, 0.5, 0.5)
	dc.Fill()

	dc.SetLineWidth(0.6)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.35)
	dc.DrawLine(20, 24, 380, 24)
	dc.Stroke()
}

// drawNode draws a centred node at the standard x position
func drawNode(dc *gg.Context, x, y, w, h float64, title, subtitle string) {
	drawNodeAt(dc, x, y, w, h, title, subtitle)
}

// drawNodeAt draws a node at an arbitrary position
func drawNodeAt(dc *gg.Context, x, y, w, h float64, title, subtitle string) {
	dc.DrawRoundedRectangle(x, y, w, h, 2)
	dc.Stroke()

	cx := x + w/2

	loadFont(dc, 7)
	dc.DrawStringAnchored(title, cx, y+h/2-5, 0.5, 0.5)
	dc.Fill()

	loadFont(dc, 5.5)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.4)
	dc.DrawStringAnchored(subtitle, cx, y+h/2+6, 0.5, 0.5)
	dc.Fill()
}

func drawArrow(dc *gg.Context, x1, y1, x2, y2 float64, label string) {
	dc.SetLineWidth(0.9)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.45)
	dc.DrawLine(x1, y1, x2, y2)
	dc.Stroke()

	drawArrowHead(dc, x2, y2, math.Pi/2)

	if label != "" {
		loadFont(dc, 5)
		dc.SetRGBA(0.07, 0.07, 0.07, 0.35)
		dc.DrawString(label, x2+4, (y1+y2)/2)
		dc.Fill()
	}
}

func drawArrowHead(dc *gg.Context, x, y, angle float64) {
	size := 5.0
	dc.SetLineWidth(0.9)
	dc.NewSubPath()
	dc.MoveTo(x, y)
	dc.LineTo(
		x-size*math.Cos(angle-0.4),
		y-size*math.Sin(angle-0.4),
	)
	dc.MoveTo(x, y)
	dc.LineTo(
		x-size*math.Cos(angle+0.4),
		y-size*math.Sin(angle+0.4),
	)
	dc.Stroke()
}

func drawSmallLabel(dc *gg.Context, x, y float64, text string) {
	loadFont(dc, 5)
	dc.SetRGBA(0.07, 0.07, 0.07, 0.4)
	dc.DrawString(text, x, y)
	dc.Fill()
}

// drawBitRow renders a row of 1/0 bit squares seeded by checksum
func drawBitRow(dc *gg.Context, x, y float64, checksum uint32) {
	for i := 0; i < 8; i++ {
		bit := (checksum >> uint(i)) & 1
		if bit == 1 {
			dc.SetRGBA(0.07, 0.07, 0.07, 0.38)
			dc.DrawRoundedRectangle(x+float64(i)*8, y, 6, 6, 0.5)
			dc.Fill()
		} else {
			dc.SetLineWidth(0.6)
			dc.SetRGBA(0.07, 0.07, 0.07, 0.3)
			dc.DrawRoundedRectangle(x+float64(i)*8, y, 6, 6, 0.5)
			dc.Stroke()
		}
	}
}
