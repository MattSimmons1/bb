package parser

import (
	"github.com/robertkrimen/otto"
	"math"
)

// script is a definition of a function called f, that takes datum as an argument
func RunScript(script string, datum interface{}) interface{} {
	vm := otto.New()
	_, err := vm.Run(script) // define function
	if err != nil {
		log("Got an error defining function '" + script + "'")
		panic(err)
	}

	err = vm.Set("d", datum) // define input to function
	if err != nil {
		// TODO: fail in strict mode
		//panic(err)
		log("Couldn't set d")
		return nil
	}

	value, err := vm.Run("f(d)") // run function
	if err != nil {
		// TODO: fail in strict mode
		//panic(err)
		log("Couldn't run the function 'f(d)'")
		return nil
	}

	d, err := value.Export()

	// NaN doesn't convert to JSON, so convert it to nil
	if floatD, ok := d.(float64); ok && math.IsNaN(floatD) {
		return nil
	}

	if err != nil {
		// TODO: fail in strict mode
		//panic(err)
		log("Couldn't export the result")
		return nil
	}

	return d
}
