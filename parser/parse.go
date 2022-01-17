
package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func removeQuotes(s string) string {
	if len(s) == 0 {
		return s
	}
	if quoteChar := s[:1]; quoteChar == `"` || quoteChar == "`" {
		if s[len(s)-1:] == quoteChar {
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
				row = append(row, item.val)  // if number doesn't parse keep as string
			} else {
			  row = append(row, number)
			}
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
		} else if item.typ == itemNull {
			row = append(row, nil)
		} else if item.typ == itemUDT {
			row = append(row, ParseUDT(item.val))
		} else {
			// definitions, comments, and spaces are ignored
		}
	}

  return row
}

func ParseInjectionMode(input string) []interface{}  {
	injectedInput := lexInjectionMode(input)

	return Parse(injectedInput)
}

func Debug(input string) {

	l := lex(input)

	for item := range l.items {
		typeName := ""
		value := ""
		jsonString := ""
		switch item.typ {
		case itemSpace:
			typeName = " space"
		case itemTab:
			typeName = " tab"
		case itemNewline:
			typeName = " newline"
		case itemEOF:
			typeName = " EOF"
		case itemUDT:
			typeName = "\nUDT"
			value = item.val
			data := ParseUDT(item.val)
			j, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			jsonString = string(j)
		case itemString:
			value = item.val
			typeName = "\nString"
		case itemNumber:
			value = item.val
			typeName = "\nNumber"
		case itemDefinition:
			value = item.val
			typeName = "\nDefinition"
		case itemAssignment:
			value = item.val
			typeName = "\nAssignment"
		case itemPropName:
			value = item.val
			typeName = "\nPropName"
		case itemPropValue:
			value = item.val
			typeName = "\nPropValue"
		case itemBool:
			value = item.val
			typeName = "\nBool"
		case itemNull:
			value = item.val
			typeName = "\nNull"
		default:
			value = item.val
			typeName = "\nvalue"
		}

		if value != "" {
			fmt.Print(typeName, " \033[92m", value, "\033[0m")
		} else {
			fmt.Print("\033[90m"+typeName, "\033[0m")
		}
		if jsonString != "" {
			fmt.Print(" \033[91m", jsonString, "\033[0m")
		}
	}

	fmt.Print("\n\n")
}

func init()  {
	defineBuiltInTypes()
}