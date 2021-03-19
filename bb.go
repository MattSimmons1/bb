/*
bb - Data Entry Markup Language

*/
package main

import (
  "encoding/json"
  "fmt"
  "github.com/spf13/cobra"
  "log"
  "strconv"
  "strings"

  //"golang.org/x/tools/cmd/goyacc"
  //"text/template/parse"
)

var verbose = false

func verbose_print(message string) {
  if verbose {
    if message == "lexBb" {
      fmt.Print("\n", message)
    } else {
      fmt.Print("/" + message)
    }
  }
}

// highlight bb syntax to preview how bb will interpret the input
func Preview(input string) {

  input = strings.Replace(input, "\\n", "\n", -1)  // convert raw escaped chars to literals
  input = strings.Replace(input, "\\t", "\t", -1)

  l := lex(input)

  for item := range l.items {

    colour := ""
    if item.typ == itemUDT {
      colour = "94"
    } else if item.typ == itemString {
      colour = "92"
    } else if item.typ == itemNumber {
      colour = "91"
    } else if item.typ == itemDefinition {
      colour = "90"
    }

    fmt.Print("\033[", colour, "m", item.val, "\033[0m")
  }
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

func Convert(input string) {
  //verbose= true
  input = strings.Replace(input, "\\n", "\n", -1)  // convert raw escaped chars to literals
  input = strings.Replace(input, "\\t", "\t", -1)

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

  j, err := json.Marshal(row)
  if err != nil {
    panic(err)
  }

  fmt.Println(string(j))
}

func init() {
  defineBuiltInTypes()
}

// https://github.com/golang/go/tree/master/src/text/template/parse
// https://www.youtube.com/watch?v=HxaD_trXwRE
func main() {
if err := func() (rootCmd *cobra.Command) {
  var IsPreview bool

  rootCmd = &cobra.Command{
    Use: "bb",
    Short: "bb command line tools",
    Args: cobra.ArbitraryArgs,
    Run: func(c *cobra.Command, args []string){
      if len(args) < 1 {
        fmt.Println("bb command line tools.\nUsage:\n  bb <input>\nUse \"bb help\" for more information.")
        return
      }
      if IsPreview {
        Preview(args[0])
        return
      }
      Convert(args[0])
      return
    },
  }
  rootCmd.PersistentFlags().BoolVarP(&IsPreview, "preview", "p", false,
    "View the interpretation of the input without converting")

  rootCmd.AddCommand(func() (createCmd *cobra.Command) {
    createCmd = &cobra.Command{
      Use:   "test",
      Short: "run unit tests",
      Run: func(c *cobra.Command, args []string){
        UnitTest()
      },
    }
    return
  }())
  return
  }().Execute(); err != nil {
    log.Panicln(err)
  }
}

/*

To Do
MVP
- [ ] Think of a better name
- [x] Refactor for multiple types
- [ ] Built in types
- [ ] lex '//' comments
- [ ] lone '+' and '-' are not allowed
- [ ] convert invalid udt to string?
- [ ] lex props
- [ ] parse udt
- [ ] backticks after unit

∆           // utc
34∆         // utc
34∆-a       // utc
34∆--       // string

*a2         // string
2021        // number
2021-02-22  // string

*/
