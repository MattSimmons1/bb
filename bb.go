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

func Convert(input string, injectionMode bool) {
  input = strings.Replace(input, "\\n", "\n", -1)  // convert raw escaped chars to literals
  input = strings.Replace(input, "\\t", "\t", -1)

  data := make([]interface{}, 0)
  if injectionMode {
    data = parser.ParseInjectionMode(input)
  } else {
    data = parser.Parse(input)
  }

  j, err := json.Marshal(data)
  if err != nil {
    panic(err)
  }

  fmt.Println(string(j))
}

func main() {
  parser.DefineBuiltInTypes()

if err := func() (rootCmd *cobra.Command) {
  var IsPreview bool
  var IsDebug bool
  var IsVerbose bool
  var isInjectionMode bool
  var definitionsFile string

  rootCmd = &cobra.Command{
    Use: "bb",
    Short: "bb - pictographic programming language\nhttps://mattsimmons1.github.io/bb/",
    Args: cobra.ArbitraryArgs,
    Run: func(c *cobra.Command, args []string){
      if len(args) < 1 {
        fmt.Println("bb - pictographic programming language\nUsage:\n  bb <input>\nUse \"bb help\" for more information.")
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

      // try to open definitions as a file - prepend to the input
      definitionsData, err := ioutil.ReadFile(definitionsFile)
      if err == nil {
        input = string(definitionsData) + "\n" + input
      } else {
        input = definitionsFile + "\n" + input
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
      Convert(input, isInjectionMode)
      return
    },
  }

  rootCmd.AddCommand(func() (createCmd *cobra.Command) {
    createCmd = &cobra.Command{
      Use:   "version",
      Short: "print the version number",
      Run: func(c *cobra.Command, args []string){
        fmt.Println("v0.2.0")
      },
    }
    return
  }())

  rootCmd.PersistentFlags().BoolVarP(&IsVerbose, "verbose", "v", false,
    "show detailed logs from the bb lexer and parser")

  rootCmd.Flags().BoolVarP(&IsPreview, "preview", "p", false,
    "view the interpretation of the input without converting")

  rootCmd.Flags().BoolVarP(&IsDebug, "explain", "e", false,
    "list every value found in the input along with the type and value when converted to JSON")

  rootCmd.Flags().BoolVarP(&isInjectionMode, "injection-mode", "i", false,
    "convert bb within comment strings of another language")

  rootCmd.Flags().StringVarP(&definitionsFile, "definitions", "d", "",
    "string or file path for additional type definitions to be used when parsing")

  return
  }().Execute(); err != nil {
    log.Panicln(err)
  }
}
