
# bb Playground

    go run server.go
    
### Build

Build the Web Assembly module: 

    GOOS=js GOARCH=wasm go build -o playground/bb.wasm ./wasm.go

Remember to update wasm_exec.js

    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./js/wasm_exec.js
