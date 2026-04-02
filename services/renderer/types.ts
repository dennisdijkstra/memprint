// uint64 fields arrive as strings from proto-loader (longs: String)
export interface RenderMeta {
  pid: number
  tid: number
  heap_addr: string
  heap_size: string
  fd: number
  nr_openat: number
  nr_mmap: number
  nr_write: number
  nr_fsync: number
  checksum: number
}

export interface PaletteConfig {
  border: [number, number, number]
  paper: [number, number, number]
  ink: [number, number, number]
  diag: [number, number, number]
  accent?: [number, number, number]
}

export interface DiagramNode {
  y: number
  w: number
  title: string
  sub: string
  seed: number
  tall?: boolean
}

export interface SideNode {
  x: number
  cx: number
  title: string
  sub: string
  seed: number
}

export interface TypographyElement {
  text: string
  y: number
  size: number
  str: number
  ns: number
  seed: number
  col: [number, number, number]
}
