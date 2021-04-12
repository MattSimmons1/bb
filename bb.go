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
<<<<<<< HEAD
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
=======
  parser.DefineBuiltInTypes()

if err := func() (rootCmd *cobra.Command) {
  var IsPreview bool
  var IsDebug bool
  var IsVerbose bool

  rootCmd = &cobra.Command{
    Use: "bb",
    Short: "bb command line tools",
    Args: cobra.ArbitraryArgs,
    Run: func(c *cobra.Command, args []string){
      if len(args) < 1 {
        fmt.Println("bb command line tools.\nUsage:\n  bb <input>\nUse \"bb help\" for more information.")
>>>>>>> main
        return
      },
    }
    rootCmd.PersistentFlags().BoolVarP(&IsPreview, "preview", "p", false,
      "view the interpretation of the input without converting")

    rootCmd.PersistentFlags().BoolVarP(&IsDebug, "debug", "d", false,
      "show each step of the parsing process")

<<<<<<< HEAD
    return
=======
      // try to open argument as a file
      data, err := ioutil.ReadFile(args[0])
      if err == nil {
        input = string(data)
      } else {
        input = args[0]
      }

      if IsVerbose {
        parser.SetVerbose()
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
  rootCmd.PersistentFlags().BoolVarP(&IsVerbose, "verbose", "v", false,
    "show detailed logs from the bb lexer and parser")

  rootCmd.PersistentFlags().BoolVarP(&IsPreview, "preview", "p", false,
    "view the interpretation of the input without converting")

  rootCmd.PersistentFlags().BoolVarP(&IsDebug, "debug", "d", false,
    "show each step of the parsing process")

  return
>>>>>>> main
  }().Execute(); err != nil {
    log.Panicln(err)
  }
}

/*

To Do
MVP
<<<<<<< HEAD
- [x] nulls
- [ ] import currency and SI
- [ ] Only look for known modifiers
- [ ] anything can be a modifier
  - [ ] make sure we still detect UDTs when the modifier is a string
- [x] lex value for UDTs with no quantity (before lexNumber)
- [ ] Invalid DTs
- [ ] rows should become an array

=======
- [x] anything can be a modifier
- [ ] Invalid DTs should become strings? strict mode?
- [ ] rows should become an array of arrays
>>>>>>> main
v1.0.0
- [ ] json PDT
- [ ] yaml PDT
- [ ] structures/arrays
- [ ] allow single - and . as UDTs or Modifiers
- [ ] safe mode / strict mode
*/
