
package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Parse(input string) []interface{} {

	l := lex(input)

	//data := make([]interface{}, 0)
	row := make([]interface{}, 0)  // TODO row logic
	for item := range l.items {
		if item.typ == itemNumber {
			number, err := strconv.ParseFloat(item.val, 64)
			if err != nil {
				panic(err)
			}
			row = append(row, map[string]float64{ "value": number })
		} else if item.typ == itemTab {
			// todo
		} else if item.typ == itemNewline {
			// todo
		} else if item.typ == itemEOF {
			// no value
		} else if item.typ == itemUDT {
			row = append(row, ParseUDT(item.val))
		} else {
			// TODO
		}
	}

  return row
}

func UnitTest() {
	verbose = true

	testInput := "∆ = { unit: pizza, length: 2, +: extra large, =:slices, #: on my tab, >: comment }\n"

	l := lex(testInput + "§µ🚀 = { unit: baseball caps }\n346 §µ🚀 ∆+ 34∆ 3.4∆=12+23#>hello \"hello\"\n/*comment*/")

	for item := range l.items {
		value := ""
		jsonString := ""
		if item.typ == itemSpace {
			value = "[space]"
		} else if item.typ == itemTab {
			value = "[tab]"
		} else if item.typ == itemNewline {
			value = "[newline]"
		} else if item.typ == itemEOF {
			value = "[EOF]"
		} else if item.typ == itemUDT {
			value = item.val
			data := ParseUDT(item.val)
			j, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			jsonString = string(j)
		} else {
			value = item.val
		}
		fmt.Print("\n  ", itemNames[item.typ], " ",  "\033[92m", value, "\033[0m \033[91m", jsonString, "\033[0m")
	}

	fmt.Print("\n\n")
}

func init() {
	defineBuiltInTypes()
}