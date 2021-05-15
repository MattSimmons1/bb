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
```bb ... ```
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
  case r == '/' || r == '\'' || r == '"' || r == '`' || r == '{' || r == '<' || r == '#' || r == '-':


      prefix := string(r) + (l.input + "     ")[l.pos:l.pos + 5]  // take the next 6 characters - pad the input so I don't have to check it's long enough to slice
      log("comment prefix is: " + prefix)

      if prefix == "<!--bb" { // look for multiline with length 6 prefix
        l.pos += 5
        l.ignore()
        return l.scanMultilineInjection("-->")
      } else if prefix = prefix[:5]; prefix == "\"\"\"bb" || prefix == "'''bb" || prefix == "```bb" { // look for multiline with length 5 prefix
        l.pos += 4
        l.ignore()
        return l.scanMultilineInjection(prefix[:3])
      } else if prefix = prefix[:4]; prefix == "/*bb" || prefix == "{-bb" { // look for multiline with length 4 prefix
        l.pos += 3
        l.ignore()
        if prefix == "/*bb" {
          return l.scanMultilineInjection("*/")
        } else {
          return l.scanMultilineInjection("-}")
        }
      } else if prefix == "//bb" || prefix == "--bb" {  // look for inline with length 4 prefix
        l.pos += 3
        l.ignore()
        return lexInlineInjection
      } else if prefix = prefix[:3]; prefix == "#bb" { // look for inline with length 3 prefix
        l.pos += 2
        l.ignore()
        return lexInlineInjection
      } else {
        return lexInjection  // it's not an injection
      }

  default:
    return lexInjection
  }
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

// injected bb that can be on multiple lines and must end with a specific suffix
func (l *lexer) scanMultilineInjection(suffix string) stateFn {
  log("scanMultilineInjection")
  log("suffix is: " + suffix)
  log("looking for: " + string(rune(suffix[0])))
  for {
    if r := l.next(); r == rune(suffix[0]) {

      for _, s := range suffix[1:] {
        log(string(rune(s)))
        if l.next() == rune(s) {
          log("is in the suffix")
          // absorb
        } else {
          log("is not part of the suffix")
          break
        }
      }

      // if we complete the loop for the whole suffix:
     l.pos -= Pos(len(suffix))  // backtrack so that the suffix isn't ignored
     l.emit(itemString)
     l.pos += Pos(len(suffix))
     return lexInjection
    } else if r == eof {
      log("Found EOF before comment suffix - fail silently")
      return lexInjection  // unclosed comment - fail silently
    } else {
      log(string(l.peek()))
    }
  }
}
