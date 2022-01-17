
package parser

import (
  "github.com/robertkrimen/otto"
)

// script is a definition of a function called f, that takes datum as an argument
func RunScript(script string, datum interface{}) interface{} {
  vm := otto.New()
  _, err := vm.Run(script)  // define function
  if err != nil {
    log("Got an error defining function '" + script + "'")
    panic(err)
  }

  err = vm.Set("d", datum)  // define input to function
  if err != nil {
    panic(err)
  }

  value, err := vm.Run("f(d)")  // run function
  if err != nil {
    panic(err)
  }

  d, err := value.Export()
  if err != nil {
    panic(err)
  }
  return d
}
