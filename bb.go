// bb - Pictographic Markup Language

package main

import (
  "bb/parser"
  "encoding/json"
  "fmt"
  "github.com/spf13/cobra"
  "io/ioutil"
  "log"
  "strings"
)

// highlight bb syntax to preview how bb will interpret the input
func Preview(input string) {
  input = strings.Replace(input, "\\n", "\n", -1)  // convert raw escaped chars to literals
  input = strings.Replace(input, "\\t", "\t", -1)
  parser.Preview(input)
}

func UnitTest() {
  parser.UnitTest()
}

func Convert(input string) {
  input = strings.Replace(input, "\\n", "\n", -1)  // convert raw escaped chars to literals
  input = strings.Replace(input, "\\t", "\t", -1)

  data := parser.Parse(input)

  j, err := json.Marshal(data)
  if err != nil {
    panic(err)
  }

  fmt.Println(string(j))
}


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

      input := ""

      // try to open argument as a file
      data, err := ioutil.ReadFile(args[0])
      if err == nil {
        input = string(data)
      } else {
        input = args[0]
      }

      if IsPreview {
        Preview(input)
        return
      }
      Convert(input)
      return
    },
  }
  rootCmd.PersistentFlags().BoolVarP(&IsPreview, "preview", "p", false,
    "view the interpretation of the input without converting")

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
- [ ] Only look for known modifiers
- [ ] anything can be a modifier
  - [ ] make sure we still detect UDTs when the modifier is a string
- [x] lex value for UDTs with no quantity (before lexNumber)
- [ ] Invalid DTs
- [ ] UDT followed by UDT - two UDTs
- [ ] strings

v0.2.0
- [ ] lex '//' comments
- [ ] structures

*/
