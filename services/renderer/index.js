const grpc = require('@grpc/grpc-js')
const protoLoader = require('@grpc/proto-loader')
const path = require('path')
const { renderPoster } = require('./poster')

// load proto from project root
const PROTO_PATH = path.join(__dirname, '../../proto/renderer.proto')

const packageDef = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
})

const proto = grpc.loadPackageDefinition(packageDef).renderer

async function renderPosterHandler(call, callback) {
  const meta = call.request

  console.log(`render job: pid=${meta.pid} heap=0x${meta.heap_addr.toString(16).toUpperCase()}`)

  try {
    const pngBuffer = await renderPoster(meta)

    callback(null, {
      png_data: pngBuffer,
      error: '',
    })

    console.log(`render done: pid=${meta.pid} size=${pngBuffer.length}B`)
  } catch (err) {
    console.error(`render failed: ${err.message}`)
    callback(null, {
      png_data: Buffer.alloc(0),
      error: err.message,
    })
  }
}

function main() {
  const server = new grpc.Server()

  server.addService(proto.RendererService.service, {
    RenderPoster: renderPosterHandler,
  })

  const port = process.env.RENDERER_PORT || '50053'
  server.bindAsync(
    `0.0.0.0:${port}`,
    grpc.ServerCredentials.createInsecure(),
    (err, port) => {
      if (err) {
        console.error(`failed to start: ${err.message}`)
        process.exit(1)
      }
      console.log(`renderer service listening on :${port}`)
    }
  )
}

main()