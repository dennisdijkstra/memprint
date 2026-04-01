const { createCanvas } = require('canvas')

const W = 400
const H = 600

// p5.js Perlin noise implementation (standalone, no browser needed)
// ported from p5.js source
class PerlinNoise {
  constructor(seed = 0) {
    this.perm = new Uint8Array(512)
    const p = new Uint8Array(256)
    for (let i = 0; i < 256; i++) p[i] = i
    // shuffle with seed
    let s = seed
    for (let i = 255; i > 0; i--) {
      s = (s * 1664525 + 1013904223) & 0xffffffff
      const j = (s >>> 0) % (i + 1);
      [p[i], p[j]] = [p[j], p[i]]
    }
    for (let i = 0; i < 512; i++) this.perm[i] = p[i & 255]
  }

  fade(t) { return t * t * t * (t * (t * 6 - 15) + 10) }
  lerp(a, b, t) { return a + t * (b - a) }
  grad(hash, x, y, z) {
    const h = hash & 15
    const u = h < 8 ? x : y
    const v = h < 4 ? y : h === 12 || h === 14 ? x : z
    return ((h & 1) ? -u : u) + ((h & 2) ? -v : v)
  }

  noise(x, y = 0, z = 0) {
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

const PALETTES = {
  ink:      { border:[17,17,17],    paper:[237,234,226], ink:[17,17,17],    diag:[17,17,17] },
  inverted: { border:[237,234,226], paper:[17,17,17],    ink:[237,234,226], diag:[237,234,226] },
  red:      { border:[192,38,38],   paper:[255,248,235], ink:[192,38,38],   diag:[192,38,38] },
  navy:     { border:[18,42,90],    paper:[255,248,235], ink:[18,42,90],    diag:[192,38,38], accent:[192,38,38] },
  forest:   { border:[30,60,40],    paper:[245,238,218], ink:[30,60,40],    diag:[180,100,20], accent:[180,100,20] },
}

function rgb(arr, alpha = 1) {
  return `rgba(${arr[0]},${arr[1]},${arr[2]},${alpha})`
}

async function renderPoster(meta) {
  const canvas = createCanvas(W, H)
  const ctx = canvas.getContext('2d')

  const waveAmt    = meta.wave       || 6
  const distAmt    = meta.distortion || 7
  const grainAmt   = meta.grain      || 3
  const b          = meta.border     || 18
  const paletteName = meta.palette   || 'ink'
  const palette    = PALETTES[paletteName] || PALETTES.ink

  const noise = new PerlinNoise(meta.pid)

  // seeded random
  let seed = meta.pid
  function random(min = 0, max = 1) {
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

  // ── diagram ──
  drawDiagram(ctx, meta, b, waveAmt, palette, noise, random)

  // ── typography ──
  await drawTypography(ctx, meta, b, distAmt, palette, noise, random, canvas)

  // ── border overlay ──
  ctx.fillStyle = rgb(palette.border)
  ctx.fillRect(0, 0, W, b)
  ctx.fillRect(0, H-b, W, b)
  ctx.fillRect(0, 0, b, H)
  ctx.fillRect(W-b, 0, b, H)

  // ── wavy inner border edge ──
  ctx.strokeStyle = rgb(palette.border)
  ctx.lineWidth = 0.8
  wavyLine(ctx, b, b, W-b, b, waveAmt*0.25, 0.1, noise)
  wavyLine(ctx, b, H-b, W-b, H-b, waveAmt*0.25, 0.2, noise)
  wavyLine(ctx, b, b, b, H-b, waveAmt*0.25, 0.3, noise)
  wavyLine(ctx, W-b, b, W-b, H-b, waveAmt*0.25, 0.4, noise)

  return canvas.toBuffer('image/png')
}

function applyGrain(ctx, b, amt, random) {
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

function wavyLine(ctx, x1, y1, x2, y2, strength, seed, noise) {
  const steps = 24
  ctx.beginPath()
  for (let t = 0; t <= steps; t++) {
    const pct = t / steps
    const x = x1 + (x2-x1)*pct
    const y = y1 + (y2-y1)*pct
    const dx = x2-x1, dy = y2-y1
    const len = Math.sqrt(dx*dx+dy*dy) || 1
    const px = -dy/len, py = dx/len
    const n = noise.noise(x*0.013+seed, y*0.013+seed) * 2 - 1
    const wx = x + px*n*strength
    const wy = y + py*n*strength
    t === 0 ? ctx.moveTo(wx, wy) : ctx.lineTo(wx, wy)
  }
  ctx.stroke()
}

function wavyRect(ctx, x, y, w, h, strength, seed, noise) {
  wavyLine(ctx, x,   y,   x+w, y,   strength, seed,   noise)
  wavyLine(ctx, x+w, y,   x+w, y+h, strength, seed+1, noise)
  wavyLine(ctx, x+w, y+h, x,   y+h, strength, seed+2, noise)
  wavyLine(ctx, x,   y+h, x,   y,   strength, seed+3, noise)
}

function wavyArrow(ctx, x1, y1, x2, y2, strength, seed, noise) {
  wavyLine(ctx, x1, y1, x2, y2, strength, seed, noise)
  const angle = Math.atan2(y2-y1, x2-x1)
  const s = 5
  ctx.beginPath()
  ctx.moveTo(x2, y2)
  ctx.lineTo(x2 - s*Math.cos(angle-0.4), y2 - s*Math.sin(angle-0.4))
  ctx.moveTo(x2, y2)
  ctx.lineTo(x2 - s*Math.cos(angle+0.4), y2 - s*Math.sin(angle+0.4))
  ctx.stroke()
}

function drawDiagram(ctx, meta, b, wa, palette, noise, random) {
  const ws = wa * 0.7
  const dc = palette.diag
  const yo = b - 8

  ctx.font = '6px monospace'
  ctx.fillStyle = rgb(dc, 0.55)
  ctx.textAlign = 'center'
  ctx.fillText('FILE UPLOAD · JOURNEY DIAGRAM', W/2, b+14)

  ctx.strokeStyle = rgb(dc, 0.35)
  ctx.lineWidth = 0.6
  wavyLine(ctx, b+10, b+20, W-b-10, b+20, ws*0.4, 0.1, noise)

  const nodes = [
    { y:32+yo,  w:120, title:'curl · browser',             sub:'multipart/form-data',              seed:10 },
    { y:90+yo,  w:140, title:'NIC · TCP/IP',               sub:'packets reassembled',               seed:20 },
    { y:148+yo, w:160, title:'socket buffer',              sub:'sk_buff · kernel space',            seed:30 },
    { y:272+yo, w:240, title:'HEAP · RAM',                 sub:`0x${(meta.heap_addr>>>0).toString(16).toUpperCase().padStart(8,'0')} · ${Math.round(meta.heap_size/1024)}KB`, seed:50, tall:true },
    { y:342+yo, w:180, title:`write · NR:${meta.nr_write}`, sub:'copy to kernel buffer',             seed:60 },
    { y:400+yo, w:160, title:`fsync · NR:${meta.nr_fsync}`, sub:'flush to disk',                     seed:70 },
    { y:458+yo, w:140, title:'filesystem · inode',         sub:`/tmp/upload_${meta.pid}.bin`,       seed:80 },
    { y:516+yo, w:140, title:'AWS S3',                     sub:'object stored · permanent',         seed:90, tall:true },
  ]

  nodes.forEach((n, i) => {
    const h = n.tall ? 48 : 36
    const x = W/2 - n.w/2
    const opacity = Math.max(0.3, 0.65 - i*0.04)

    ctx.save()
    const tilt   = (noise.noise(n.seed, 5)*2-1) * wa * 0.28 * Math.PI/180
    const shiftX = (noise.noise(n.seed, 6)*2-1) * ws * 0.5
    const shiftY = (noise.noise(n.seed, 7)*2-1) * ws * 0.25
    ctx.translate(W/2+shiftX, n.y+h/2+shiftY)
    ctx.rotate(tilt)
    ctx.translate(-W/2, -(n.y+h/2))

    ctx.strokeStyle = rgb(dc, opacity)
    ctx.lineWidth = 1.2
    wavyRect(ctx, x, n.y, n.w, h, ws*0.6, n.seed, noise)

    ctx.fillStyle = rgb(dc, opacity)
    ctx.textAlign = 'center'
    ctx.font = 'bold 7px monospace'
    ctx.fillText(n.title, W/2, n.y+h/2-4)
    ctx.font = '5.5px monospace'
    ctx.fillStyle = rgb(dc, opacity*0.65)
    ctx.fillText(n.sub, W/2, n.y+h/2+7)
    ctx.restore()
  })

  // openat + mmap side nodes
  const sideY = 210+yo
  ;[
    { x:12,  cx:77,  title:`openat · NR:${meta.nr_openat}`, sub:`fd:${meta.fd} assigned`, seed:100 },
    { x:258, cx:323, title:`mmap · NR:${meta.nr_mmap}`,     sub:'virtual memory map',      seed:110 },
  ].forEach(n => {
    ctx.save()
    const tilt   = (noise.noise(n.seed,5)*2-1)*wa*0.35*Math.PI/180
    const shiftX = (noise.noise(n.seed,6)*2-1)*ws*0.5
    const shiftY = (noise.noise(n.seed,7)*2-1)*ws*0.25
    ctx.translate(n.cx+shiftX, sideY+18+shiftY)
    ctx.rotate(tilt)
    ctx.translate(-n.cx, -sideY-18)
    ctx.strokeStyle = rgb(dc, 0.58)
    ctx.lineWidth = 1.2
    wavyRect(ctx, n.x, sideY, 130, 36, ws*0.6, n.seed, noise)
    ctx.fillStyle = rgb(dc, 0.58)
    ctx.textAlign = 'center'
    ctx.font = 'bold 6.5px monospace'
    ctx.fillText(n.title, n.cx, sideY+14)
    ctx.font = '5.5px monospace'
    ctx.fillStyle = rgb(dc, 0.38)
    ctx.fillText(n.sub, n.cx, sideY+24)
    ctx.restore()
  })

  // arrows
  ctx.strokeStyle = rgb(dc, 0.48)
  ctx.lineWidth = 0.9
  const yo2 = yo
  ;[
    [200,68+yo2,200,90+yo2],[200,126+yo2,200,148+yo2],
    [165,184+yo2,100,210+yo2],[235,184+yo2,300,210+yo2],
    [77,246+yo2,160,272+yo2],[323,246+yo2,240,272+yo2],
    [200,320+yo2,200,342+yo2],[200,378+yo2,200,400+yo2],
    [200,436+yo2,200,458+yo2],[200,494+yo2,200,516+yo2],
  ].forEach(([x1,y1,x2,y2], i) => wavyArrow(ctx,x1,y1,x2,y2,ws*0.5,i*0.4,noise))

  // footer
  ctx.strokeStyle = rgb(dc, 0.2)
  ctx.lineWidth = 0.5
  wavyLine(ctx, b+10, H-b-12, W-b-10, H-b-12, ws*0.3, 0.5, noise)
  ctx.fillStyle = rgb(dc, 0.25)
  ctx.textAlign = 'center'
  ctx.font = '5px monospace'
  ctx.fillText(
    `PID:${meta.pid} · FD:${meta.fd} · NR:${meta.nr_openat} · NR:${meta.nr_mmap} · NR:${meta.nr_write} · NR:${meta.nr_fsync}`,
    W/2, H-b-4
  )
}

async function drawTypography(ctx, meta, b, da, palette, noise, random, canvas) {
  const d = da
  const ns = 0.013
  const yo = b - 8
  const ink = palette.ink
  const acc = palette.accent || palette.ink

  const elements = [
    { text:`0x${(meta.heap_addr>>>0).toString(16).toUpperCase().padStart(8,'0')}`,
      y:88+yo,  size:78, str:d*3.2, ns:ns*0.8, seed:0.1, col:ink },
    { text:`_${(((meta.heap_addr>>>16)&0xFFFF)>>>0).toString(16).toUpperCase().padStart(4,'0')}_${((meta.heap_addr&0xFFFF)>>>0).toString(16).toUpperCase().padStart(4,'0')}`,
      y:148+yo, size:54, str:d*2.6, ns:ns*0.9, seed:0.2, col:ink },
    { text:`PID:${meta.pid}`,
      y:210+yo, size:48, str:d*3.8, ns:ns,     seed:0.3, col:ink },
    { text:`TID:${meta.tid}`,
      y:262+yo, size:40, str:d*2.4, ns:ns*1.1, seed:0.4, col:acc },
    { text:'HEAP',
      y:348+yo, size:86, str:d*5.5, ns:ns*0.7, seed:0.5, col:ink },
    { text:`NR:${meta.nr_mmap}·MMAP`,
      y:412+yo, size:38, str:d*2.8, ns:ns,     seed:0.6, col:acc },
    { text:`NR:${meta.nr_write}·WRITE`,
      y:458+yo, size:36, str:d*3.8, ns:ns*0.9, seed:0.7, col:ink },
    { text:`NR:${meta.nr_fsync}·FSYNC`,
      y:504+yo, size:34, str:d*5.0, ns:ns*0.8, seed:0.8, col:acc },
    { text:`G#${meta.pid+1}·G#${meta.pid+2}·G#${meta.pid+3}`,
      y:544+yo, size:26, str:d*2.4, ns:ns,     seed:0.9, col:ink },
    { text:`${Math.round(meta.heap_size/1024)}KB·CHECKSUM·32B`,
      y:578+yo, size:19, str:d*4.5, ns:ns*0.8, seed:1.0, col:ink },
  ]

  for (const el of elements) {
    await displacedText(ctx, el.text, b+4, el.y, el.size, el.str, el.ns, el.seed, el.col, noise)
  }

  // press line
  for (let x = b; x < W-b; x++) {
    const ny = noise.noise(x*0.02, 888) * d - d*0.5
    ctx.fillStyle = rgb(palette.ink, 0.12 + random(0, 0.05))
    ctx.fillRect(x, 365+yo+ny, 1, 2)
  }
}

async function displacedText(ctx, txt, x, y, size, strength, noiseScale, seedOffset, col, noise) {
  // render text to offscreen canvas
  const pad = Math.ceil(strength) + 10
  const offW = W + pad*2
  const offH = size*2 + pad*2
  const off = createCanvas(offW, offH)
  const offCtx = off.getContext('2d')

  offCtx.font = `bold ${size}px monospace`
  offCtx.fillStyle = rgb(col)
  offCtx.textBaseline = 'alphabetic'
  offCtx.fillText(txt, pad, size + pad)

  // pixel displacement
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

module.exports = { renderPoster }