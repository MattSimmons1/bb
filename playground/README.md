
# bb Playground

    go run server.go
    
### Build

Build the Web Assembly module: 

    GOOS=js GOARCH=wasm go build -o playground/bb.wasm ./wasm.go