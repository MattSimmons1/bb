
package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func removeQuotes(s string) string {
	if s[:1] == `"` {
		if s[len(s)-1:] == `"` {
			return s[1:len(s)-1]
		}
	}
	return s
}

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
			row = append(row, number)
		} else if item.typ == itemTab {
			// todo
		} else if item.typ == itemNewline {
			// todo
		} else if item.typ == itemString {
			row = append(row, strings.TrimSpace(removeQuotes(item.val)))
		} else if item.typ == itemBool {
			if item.val == "true" {
			  row = append(row, true)
			} else {
			  row = append(row, false)
			}
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

	testInput := "∆ = { unit: pizza, length: 2, +: extra large, =:slices, #: on my tab, >: comment }\n" +
		           "§µ🚀 = { unit: baseball caps }\n" +
		           "𓂀 = { type: eye of horus }\n" +
		           "£2 gg 9g\"# w\\\"ow\" ∆ 346 1元100 100°C hello §µ🚀 ∆+ 34∆-2.0 3.4∆=12+23#>`hello` \"hello\"\n/*comment*/"
	l := lex(testInput)

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