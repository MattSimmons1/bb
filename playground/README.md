
# bb Playground

    go run server.go
    
### Build

Build the Web Assembly module: 

    GOOS=js GOARCH=wasm go build -o bb.wasm ./wasm.go && cp bb.wasm playground/bb.wasm 