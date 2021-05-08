package parser


// lex creates a new scanner for the input string.
func lexInjectionMode(input string) string {

  l := &lexer{
    name:        "bb",
    input:       input + "\n",
    items:       make(chan item),
    line:        1,
    startLine:   1,
  }
  go l.injectionModeRun()

  injectedInput := ""

  for item := range l.items {
    injectedInput += item.val
  }

  log("injectedInput")
  log(injectedInput)
  log("\n---------------------------------------------")

  return injectedInput
}

// run runs the state machine for the lexer in comment mode.
func (l *lexer) injectionModeRun() {
  for state := lexInjection; state != nil; {
    state = state(l)
  }
  close(l.items)
}

/* look through non-bb file for comments that start with bb, e.g.:

//bb ...
--bb ...
#bb ...
/*bb ... *\/
<!--bb ... -->
"""bb ... """
'''bb ... '''
{-bb ... -}


Keep going until we find one of these
*/
func lexInjection(l *lexer) stateFn {
  log("lexInjection")
  log(string(l.peek()))

  switch r := l.next(); {
  case r == eof:
    l.ignore()
    l.emit(itemEOF)
    return nil
  case r == '/':
    if l.accept("/") {
      if len(l.input[l.pos:]) > 1 {

        log("is '" + l.input[l.pos:l.pos+2] + "' injected bb?")
        if l.input[l.pos:l.pos+2] == "bb" {
          log("yes")
          l.pos += 2
          l.ignore()
          return lexInlineInjection
        }
      }

      log("no")
      return lexInjection

    }
  default:
    return lexInjection
  }
  return lexInjection
}

// injected bb that can only be on one line
func lexInlineInjection(l *lexer) stateFn {
  log("lexInlineInjection")
  for {
    if r := l.next(); r == '\n' || r == eof {
      l.emit(itemString)
      return lexInjection
    } else {
      log(string(l.peek()))
    }
  }
}

func lexOtherLanguage(l *lexer) stateFn {
  log("lexOtherLanguage")
  for {
    if r := l.next(); r == '\n' || r == eof {
      l.ignore()
      return lexInjection
    } else {
      log(string(l.peek()))
    }
  }
}
