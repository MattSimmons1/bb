// +build js,wasm

package main

import (
  "bb/parser"
  "fmt"
  "syscall/js"
)

func main() {
  fmt.Println("Hello wasm")
  js.Global().Get("wasm").Set("bb", js.FuncOf(WASMConvert))
  js.Global().Get("wasm").Set("bbSyntax", js.FuncOf(WASMSyntax))

  parser.DefineBuiltInTypes()

  select {}  // don't exit
}


func WASMConvert(this js.Value, p []js.Value) interface{} {
  fmt.Println(p[0].String())

  data := parser.Parse(p[0].String())

  return js.ValueOf(data)
}


func WASMSyntax(this js.Value, p []js.Value) interface{} {
  fmt.Println(p[0].String())

  data := parser.Syntax(p[0].String())

  return js.ValueOf(data)
}