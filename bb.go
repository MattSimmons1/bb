// bb - Pictographic Markup Language

package main

import (
  "bb/parser"
  "encoding/json"
  "fmt"
  "github.com/spf13/cobra"
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
  //verbose= true
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
      if IsPreview {
        Preview(args[0])
        return
      }
      Convert(args[0])
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
- [ ] lex value for UDTs with no quantity (before lexNumber)
- [ ] parse numerical value
- [ ] UDT followed by UDT - two UDTs
- [ ] strings
  - [ ] convert invalid udt to string?
v0.2.0
- [ ] lex '//' comments
- [ ] right argument
  - [ ] backticks after unit

*/
