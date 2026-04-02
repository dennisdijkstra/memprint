import * as grpc from '@grpc/grpc-js'
import * as protoLoader from '@grpc/proto-loader'
import * as path from 'path'
import { renderPoster } from './poster'
import { RenderMeta } from './types'

const PROTO_PATH = path.join(__dirname, '../../proto/renderer.proto')

const packageDef = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
})

const proto = grpc.loadPackageDefinition(packageDef)
const rendererPackage = proto['renderer'] as grpc.GrpcObject
const RendererService = rendererPackage['RendererService'] as grpc.ServiceClientConstructor

interface RenderResponse {
  png_data: Buffer
  error: string
}

async function renderPosterHandler(
  call: grpc.ServerUnaryCall<RenderMeta, RenderResponse>,
  callback: grpc.sendUnaryData<RenderResponse>
): Promise<void> {
  const meta = call.request

  console.log(`render job: pid=${meta.pid} heap=0x${parseInt(meta.heap_addr, 10).toString(16).toUpperCase()}`)

  try {
    const pngBuffer = await renderPoster(meta)

    callback(null, {
      png_data: pngBuffer,
      error: '',
    })

    console.log(`render done: pid=${meta.pid} size=${pngBuffer.length}B`)
  } catch (err) {
    const error = err as Error
    console.error(`render failed: ${error.message}`)
    callback(null, {
      png_data: Buffer.alloc(0),
      error: error.message,
    })
  }
}

function main(): void {
  const server = new grpc.Server()

  server.addService(RendererService.service, {
    RenderPoster: renderPosterHandler,
  })

  const port = process.env['RENDERER_PORT'] ?? '50053'
  server.bindAsync(
    `0.0.0.0:${port}`,
    grpc.ServerCredentials.createInsecure(),
    (err: Error | null, boundPort: number) => {
      if (err) {
        console.error(`failed to start: ${err.message}`)
        process.exit(1)
      }
      console.log(`renderer service listening on :${boundPort}`)
    }
  )
}

main()
