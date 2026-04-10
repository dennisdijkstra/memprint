import { createCanvas, CanvasRenderingContext2D } from 'canvas'
import { RenderMeta, PaletteConfig, DiagramNode, SideNode, TypographyElement } from './types'

export { RenderMeta }

const S = 2                                    // render scale factor — change to increase/decrease output resolution
const H = 700 * S                              // 1400px
const W = Math.round(H / Math.SQRT2 / S) * S  // A-series ratio (1:√2) → 990px

// p5.js Perlin noise implementation (standalone, no browser needed)
// ported from p5.js source
class PerlinNoise {
  private perm: Uint8Array

  constructor(seed = 0) {
    this.perm = new Uint8Array(512)
    const p = new Uint8Array(256)
    for (let i = 0; i < 256; i++) p[i] = i
    let s = seed
    for (let i = 255; i > 0; i--) {
      s = (s * 1664525 + 1013904223) & 0xffffffff
      const j = (s >>> 0) % (i + 1)
      ;[p[i], p[j]] = [p[j], p[i]]
    }
    for (let i = 0; i < 512; i++) this.perm[i] = p[i & 255]
  }

  private fade(t: number): number { return t * t * t * (t * (t * 6 - 15) + 10) }
  private lerp(a: number, b: number, t: number): number { return a + t * (b - a) }
  private grad(hash: number, x: number, y: number, z: number): number {
    const h = hash & 15
    const u = h < 8 ? x : y
    const v = h < 4 ? y : h === 12 || h === 14 ? x : z
    return ((h & 1) ? -u : u) + ((h & 2) ? -v : v)
  }

  noise(x: number, y = 0, z = 0): number {
    const X = Math.floor(x) & 255
    const Y = Math.floor(y) & 255
    const Z = Math.floor(z) & 255
    x -= Math.floor(x); y -= Math.floor(y); z -= Math.floor(z)
    const u = this.fade(x), v = this.fade(y), w = this.fade(z)
    const p = this.perm
    const A = p[X]+Y, AA = p[A]+Z, AB = p[A+1]+Z
    const B = p[X+1]+Y, BA = p[B]+Z, BB = p[B+1]+Z
    return (this.lerp(
      this.lerp(
        this.lerp(this.grad(p[AA],x,y,z),   this.grad(p[BA],x-1,y,z),   u),
        this.lerp(this.grad(p[AB],x,y-1,z), this.grad(p[BB],x-1,y-1,z), u), v),
      this.lerp(
        this.lerp(this.grad(p[AA+1],x,y,z-1),   this.grad(p[BA+1],x-1,y,z-1),   u),
        this.lerp(this.grad(p[AB+1],x,y-1,z-1), this.grad(p[BB+1],x-1,y-1,z-1), u), v),
      w) + 1) / 2
  }
}

const PAPER: [number, number, number] = [250, 246, 235]

function hslToRgb(h: number, s: number, l: number): [number, number, number] {
  s /= 100; l /= 100
  const a = s * Math.min(l, 1 - l)
  const f = (n: number) => { const k = (n + h / 30) % 12; return l - a * Math.max(-1, Math.min(k - 3, 9 - k, 1)) }
  return [Math.round(f(0) * 255), Math.round(f(8) * 255), Math.round(f(4) * 255)]
}

// Mix all metadata fields into one deterministic seed — unique per uploaded file
function hashMix(...vals: number[]): number {
  let h = 0x811c9dc5
  for (const v of vals) {
    for (let shift = 0; shift < 32; shift += 8) {
      h ^= (v >>> shift) & 0xff
      h = Math.imul(h, 0x01000193) & 0xffffffff
    }
  }
  return h >>> 0
}

function rgb(arr: [number, number, number], alpha = 1): string {
  return `rgba(${arr[0]},${arr[1]},${arr[2]},${alpha})`
}

export async function renderPoster(meta: RenderMeta): Promise<Buffer> {
  const canvas = createCanvas(W, H)
  const ctx = canvas.getContext('2d')

  const masterSeed = hashMix(
    meta.pid, meta.tid, Number(meta.heap_addr),
    meta.checksum, meta.fd,
    meta.nr_openat, meta.nr_mmap, meta.nr_write, meta.nr_fsync,
    meta.num_goroutines, meta.num_cpu, meta.num_gc,
    Math.round(meta.file_entropy * 1000), meta.page_size, meta.file_pages
  )
  function derive(salt: number): number {
    let n = (masterSeed ^ Math.imul(salt, 2654435761)) & 0xffffffff
    n = Math.imul(n ^ (n >>> 16), 0x45d9f3b) & 0xffffffff
    n = Math.imul(n ^ (n >>> 16), 0x45d9f3b) & 0xffffffff
    return (n >>> 0) / 0xffffffff
  }

  const inkColor = hslToRgb((meta.checksum * 137.508) % 360, 70, 30)
  const palette: PaletteConfig = { border: inkColor, paper: PAPER, ink: inkColor, diag: inkColor }

  const waveAmt  = (3 + derive(2) * 9)  * S
  const distAmt  = (4 + derive(3) * 10) * S
  const grainAmt = 2 + derive(4) * 5
  const b        = 42 * S

  const noise = new PerlinNoise(meta.pid)

  // seeded random
  let seed = meta.pid
  function random(min = 0, max = 1): number {
    seed = (seed * 1664525 + 1013904223) & 0xffffffff
    const r = ((seed >>> 0) / 0xffffffff)
    return min + r * (max - min)
  }

  // ── background ──
  ctx.fillStyle = rgb(palette.border)
  ctx.fillRect(0, 0, W, H)
  ctx.fillStyle = rgb(palette.paper)
  ctx.fillRect(b, b, W-b*2, H-b*2)

  // ── grain ──
  if (grainAmt > 0) applyGrain(ctx, b, grainAmt, random)

  // ── typography ──
  await drawTypography(ctx, meta, b, distAmt, palette, noise, random, derive)

  // ── low-level metadata strips ──
  drawMetaStrips(ctx, meta, b, distAmt, palette, noise, derive)

  // ── border overlay ──
  ctx.fillStyle = rgb(palette.border)
  ctx.fillRect(0, 0, W, b)
  ctx.fillRect(0, H-b, W, b)
  ctx.fillRect(0, 0, b, H)
  ctx.fillRect(W-b, 0, b, H)

  // ── border grain ──
  applyBorderGrain(ctx, b, masterSeed)

  // ── wavy inner border edge ──
  ctx.strokeStyle = rgb(palette.border)
  ctx.lineWidth = 0.8 * S
  wavyLine(ctx, b, b, W-b, b, waveAmt*0.25, 0.1, noise)
  wavyLine(ctx, b, H-b, W-b, H-b, waveAmt*0.25, 0.2, noise)
  wavyLine(ctx, b, b, b, H-b, waveAmt*0.25, 0.3, noise)
  wavyLine(ctx, W-b, b, W-b, H-b, waveAmt*0.25, 0.4, noise)

  return canvas.toBuffer('image/png')
}

function applyGrain(ctx: CanvasRenderingContext2D, b: number, amt: number, random: (min?: number, max?: number) => number): void {
  const imageData = ctx.getImageData(b, b, W-b*2, H-b*2)
  const d = imageData.data
  for (let i = 0; i < d.length; i += 4) {
    const g = random(-amt, amt)
    d[i]   = Math.min(255, Math.max(0, d[i]   + g))
    d[i+1] = Math.min(255, Math.max(0, d[i+1] + g))
    d[i+2] = Math.min(255, Math.max(0, d[i+2] + g))
  }
  ctx.putImageData(imageData, b, b)
}

function applyBorderGrain(ctx: CanvasRenderingContext2D, b: number, seed: number): void {
  // per-pixel hash — returns 0..1, no spatial correlation
  function hash(x: number, y: number): number {
    let n = (x * 1664525 ^ y * 1013904223 ^ seed) & 0xffffffff
    n = Math.imul(n ^ (n >>> 16), 0x45d9f3b) & 0xffffffff
    n = Math.imul(n ^ (n >>> 16), 0x45d9f3b) & 0xffffffff
    return (n >>> 0) / 0xffffffff
  }

  // value noise — smoothstep-interpolated hash, non-repeating and organic
  function valueNoise(x: number, y: number, scale: number): number {
    const sx = x / scale, sy = y / scale
    const ix = Math.floor(sx), iy = Math.floor(sy)
    const fx = sx - ix, fy = sy - iy
    const ux = fx * fx * (3 - 2 * fx)  // smoothstep
    const uy = fy * fy * (3 - 2 * fy)
    return hash(ix,   iy  ) * (1-ux) * (1-uy)
         + hash(ix+1, iy  ) * ux     * (1-uy)
         + hash(ix,   iy+1) * (1-ux) * uy
         + hash(ix+1, iy+1) * ux     * uy
  }

  const strips = [
    { x: 0,   y: 0,   w: W,     h: b       },  // top
    { x: 0,   y: H-b, w: W,     h: b       },  // bottom
    { x: 0,   y: b,   w: b,     h: H-b*2   },  // left
    { x: W-b, y: b,   w: b,     h: H-b*2   },  // right
  ]

  for (const strip of strips) {
    const imageData = ctx.getImageData(strip.x, strip.y, strip.w, strip.h)
    const d = imageData.data
    for (let i = 0; i < d.length; i += 4) {
      const pixIdx = i / 4
      const px = strip.x + (pixIdx % strip.w)
      const py = strip.y + Math.floor(pixIdx / strip.w)
      const lx = (px / S) | 0
      const ly = (py / S) | 0

      // Two octaves of value noise for organic, non-repeating density variation
      const coarse = valueNoise(lx, ly, 10)   // large clusters
      const medium = valueNoise(lx, ly, 4)    // smaller variation
      const organic = coarse * 0.6 + medium * 0.4

      // Per-pixel hash for fine grain detail
      const fine = hash(lx, ly)

      // Mostly hash-driven so it reads as random, organic shapes the clustering
      const combined = organic * 0.25 + fine * 0.75

      const threshold = 0.80
      const g = combined > threshold
        ? Math.pow((combined - threshold) / (1 - threshold), 1.4) * 120
        : 0

      d[i]   = Math.min(255, d[i]   + g)
      d[i+1] = Math.min(255, d[i+1] + g)
      d[i+2] = Math.min(255, d[i+2] + g)
    }
    ctx.putImageData(imageData, strip.x, strip.y)
  }
}

function wavyLine(ctx: CanvasRenderingContext2D, x1: number, y1: number, x2: number, y2: number, strength: number, seed: number, noise: PerlinNoise): void {
  const steps = 24 * S  // more steps for smoother curves at higher resolution
  const dx = x2-x1, dy = y2-y1
  const len = Math.sqrt(dx*dx+dy*dy) || 1
  const px = -dy/len, py = dx/len

  ctx.beginPath()
  for (let t = 0; t <= steps; t++) {
    const pct = t / steps
    const x = x1 + dx*pct
    const y = y1 + dy*pct
    const n = noise.noise(x*(0.013/S)+seed, y*(0.013/S)+seed) * 2 - 1
    const wx = x + px*n*strength
    const wy = y + py*n*strength
    t === 0 ? ctx.moveTo(wx, wy) : ctx.lineTo(wx, wy)
  }
  ctx.stroke()
}

async function drawTypography(ctx: CanvasRenderingContext2D, meta: RenderMeta, b: number, distAmt: number, palette: PaletteConfig, noise: PerlinNoise, random: (min?: number, max?: number) => number, derive: (salt: number) => number): Promise<void> {
  const ns = 0.013/S
  const yo = b - 8*S
  const ink = palette.ink
  const acc = palette.accent ?? palette.ink

  const heapAddr = Number(meta.heap_addr)

  // seeded shuffle — Fisher-Yates using derive(), salt starts at 600
  let salt = 600
  function seededShuffle<T>(arr: T[]): T[] {
    const a = [...arr]
    for (let i = a.length - 1; i > 0; i--) {
      const j = Math.floor(derive(salt++) * (i + 1))
      ;[a[i], a[j]] = [a[j], a[i]]
    }
    return a
  }

  // text content pool — shuffled independently of sizes
  const texts = seededShuffle([
    `0x${(heapAddr >>> 0).toString(16).toUpperCase().padStart(8,'0')}`,
    `_${(((heapAddr>>>16)&0xFFFF)>>>0).toString(16).toUpperCase().padStart(4,'0')}_${((heapAddr&0xFFFF)>>>0).toString(16).toUpperCase().padStart(4,'0')}`,
    `PID:${meta.pid}`,
    `TID:${meta.tid}`,
    'HEAP',
    `NR:${meta.nr_mmap}·MMAP`,
    `NR:${meta.nr_write}·WRITE`,
    `NR:${meta.nr_fsync}·FSYNC`,
    `G#${Math.round(derive(400)*9999)}·G#${Math.round(derive(401)*9999)}·G#${Math.round(derive(402)*9999)}`,
    `${meta.checksum.toString(16).toUpperCase().padStart(8,'0')}·${Math.round(Number(meta.heap_size)/1024)}KB`,
  ])

  // size pool — shuffled independently so any text can be any size
  const sizes = seededShuffle([86, 78, 54, 48, 40, 38, 36, 34, 26, 19])

  // y-slots stay fixed — guarantees even page coverage regardless of shuffle
  const ySlots = [88, 148, 210, 262, 348, 412, 458, 504, 544, 578]

  const elements: TypographyElement[] = texts.map((text, i) => {
    const baseSize = sizes[i]
    return {
      text,
      y:    ySlots[i] * S + yo,
      size: baseSize * S,
      str:  distAmt * (2.0 + (baseSize / 86) * 3.5),
      ns:   ns * (0.7 + derive(salt++) * 0.5),
      seed: derive(salt++),
      col:  derive(salt++) > 0.35 ? ink : acc,
    }
  })

  elements.forEach((el, i) => {
    const yJitter = (derive(200 + i) - 0.5) * 20 * S
    displacedText(ctx, el.text, b+4*S, el.y + yJitter, el.size, el.str, el.ns, el.seed, el.col, noise)
  })

  // press line
  // for (let x = b; x < W-b; x++) {
  //   const ny = noise.noise(x*(0.02/S), 888) * distAmt - distAmt*0.5
  //   ctx.fillStyle = rgb(palette.ink, 0.12 + random(0, 0.05))
  //   ctx.fillRect(x, 365*S+yo+ny, 1, 2)
  // }
}

function drawMetaStrips(
  ctx: CanvasRenderingContext2D,
  meta: RenderMeta,
  b: number,
  distAmt: number,
  palette: PaletteConfig,
  noise: PerlinNoise,
  derive: (salt: number) => number
): void {
  const yo     = b - 8 * S
  const startY = 590 * S + yo
  const rowH   = 28 * S
  const ink    = palette.ink

  const rows = [
    { label: 'CPU TOPOLOGY', data: `CPU·${meta.num_cpu} · MAXPROCS·${meta.go_max_procs} · G·${meta.num_goroutines}`, seed: 1.10 },
    { label: 'GC STATE',     data: `GC#${meta.num_gc} · PAUSE·${Number(meta.gc_pause_total_ns)}ns`,                  seed: 1.20 },
    { label: 'MEMORY PAGING',data: `PAGE·${meta.page_size}B · PAGES·${meta.file_pages}`,                             seed: 1.30 },
    { label: 'FILE ENTROPY', data: `ENTROPY·${meta.file_entropy.toFixed(2)}/8.00`,                                   seed: 1.40 },
  ]

  rows.forEach((row, i) => {
    const y = startY + i * rowH
    ctx.font      = `${5 * S}px monospace`
    ctx.fillStyle = rgb(ink, 0.28)
    ctx.textAlign = 'center'
    ctx.fillText(row.label, W / 2, y - 4 * S)
    displacedText(ctx, row.data, b + 4 * S, y + 14 * S, 16 * S, distAmt * 1.8, 0.013 / S, row.seed, ink, noise)
  })

  // separator before magic bytes
  const sepY = startY + rows.length * rowH + 4 * S
  ctx.strokeStyle = rgb(ink, 0.2)
  ctx.lineWidth   = 0.5 * S
  wavyLine(ctx, b + 10 * S, sepY, W - b - 10 * S, sepY, distAmt * 0.3, derive(500), noise)

  // magic bytes — crisp monospace, no displacement
  ctx.font      = `${9 * S}px monospace`
  ctx.fillStyle = rgb(ink, 0.55)
  ctx.textAlign = 'center'
  ctx.fillText(meta.magic_bytes, W / 2, sepY + 16 * S)
}

function displacedText(ctx: CanvasRenderingContext2D, txt: string, x: number, y: number, size: number, strength: number, noiseScale: number, seedOffset: number, col: [number, number, number], noise: PerlinNoise): void {
  const pad = Math.ceil(strength) + 10*S
  const offW = W + pad*2
  const offH = size*2 + pad*2
  const off = createCanvas(offW, offH)
  const offCtx = off.getContext('2d')

  offCtx.font = `bold ${size}px monospace`
  offCtx.fillStyle = rgb(col)
  offCtx.textBaseline = 'alphabetic'
  offCtx.fillText(txt, pad, size + pad)

  const src = offCtx.getImageData(0, 0, offW, offH)
  const dst = offCtx.createImageData(offW, offH)
  const sd = src.data, dd = dst.data

  for (let py = 0; py < offH; py++) {
    for (let px = 0; px < offW; px++) {
      const nx = noise.noise(px*noiseScale+seedOffset,    py*noiseScale+seedOffset,    0) * 2 - 1
      const ny = noise.noise(px*noiseScale+seedOffset+50, py*noiseScale+seedOffset+50, 1) * 2 - 1
      const srcX = Math.round(Math.min(Math.max(px - nx*strength,     0), offW-1))
      const srcY = Math.round(Math.min(Math.max(py - ny*strength*0.4, 0), offH-1))
      const si = (srcY*offW+srcX)*4
      const di = (py*offW+px)*4
      dd[di]=sd[si]; dd[di+1]=sd[si+1]; dd[di+2]=sd[si+2]; dd[di+3]=sd[si+3]
    }
  }

  offCtx.putImageData(dst, 0, 0)
  ctx.drawImage(off, x-pad, y-size-pad)
}
