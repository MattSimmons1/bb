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

func Debug(input string) {
  input = strings.Replace(input, "\\n", "\n", -1)  // convert raw escaped chars to literals
  input = strings.Replace(input, "\\t", "\t", -1)
  parser.Debug(input)
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
  var IsDebug bool

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

      if IsDebug {
        Debug(input)
        return
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

  rootCmd.PersistentFlags().BoolVarP(&IsDebug, "debug", "d", false,
    "show each step of the parsing process")

  return
  }().Execute(); err != nil {
    log.Panicln(err)
  }
}

/*

To Do
MVP
- [x] anything can be a modifier
  - [x] Only look for known modifiers
  - [x] string props and modiifiers are the same
- [x] lex value for UDTs with no quantity (before lexNumber)
- [ ] Invalid DTs should become strings? strict mode?
- [ ] rows should become an array of arrays
v1.0.0
- [ ] json PDT
- [ ] yaml PDT
- [ ] structures/arrays
- [ ] allow single - and . as UDTs or Modifiers
*/
