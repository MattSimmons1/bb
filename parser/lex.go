// bb lexer
// Heavily based on https://github.com/golang/go/tree/master/src/text/template/parse
// See this talk for a great explanation of how it works: https://www.youtube.com/watch?v=HxaD_trXwRE

package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pos represents a byte position in the original input text  from which
// this template was parsed.
type Pos int

// item represents a token or text string returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
	message string  // additional info about the item, e.g. an error message
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError        itemType = iota // error occurred; value is text of error
	itemBool                         // boolean constant
	itemNull                         // JSON null
	itemEOF
	itemNumber      // simple number, including imaginary
	itemPipe        // pipe symbol
	itemSpace       // run of spaces separating arguments
	itemString      // quoted string (includes quotes)
	itemKeyword     // used only to delimit the keywords
	itemUDT         // user defined type
	itemTab         // two or more spaces or a tab
	itemDefinition  // type definition, e.g. ∆ = { unit: pizza }
	itemNewline     // \n
	itemAssignment
	itemPropName
	itemPropValue
	itemComment     // inline or multiline comment
)

const eof = -1

const (
	modifiers     = "+~<>:;/|#&≠≥≤^*$£,?!•°·"  // all standard modifiers
	quotes        = "`\""
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name        string    // the name of the input; used only for error reports
	input       string    // the string being scanned
	emitComment bool      // emit itemComment tokens.
	pos         Pos       // current position in the input
	start       Pos       // start position of this item
	width       Pos       // width of last rune read from input
	items       chan item // channel of scanned items
	parenDepth  int       // nesting depth of ( ) exprs
	line        int       // 1+number of newlines seen
	startLine   int       // start line of this item
}

var verbose = false

func SetVerbose() {
	verbose = true
}

func log(message string) {
	if verbose {
		if message == "lexBb" {
			fmt.Print("\n", "\033[92m", message, "\033[0m")
		} else {
			if strings.HasPrefix(message, "lex") {
			  fmt.Print("/", "\033[92m", message, "\033[0m")
			} else {
			  fmt.Print("/", message)
			}
		}
	}
}


// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.startLine, ""}
	l.start = l.pos
	l.startLine = l.line
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, l.input[l.start:l.pos], l.startLine, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {

	l := &lexer{
		name:        "bb",
		input:       input + "\n",
		items:       make(chan item),
		line:        1,
		startLine:   1,
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexBb; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// state functions

const (
	leftComment  = "/*"
	rightComment = "*/"
)

// lexComment scans a comment. The left comment marker is known to be present.
func lexComment(l *lexer) stateFn {
	log("lexComment")
	l.pos += Pos(len(leftComment))
	i := strings.Index(l.input[l.pos:], rightComment)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i + len(rightComment))
	l.emit(itemComment)
	return lexBb
}

func lexInlineComment(l *lexer) stateFn {
	log("lexInlineComment")
	i := strings.Index(l.input[l.pos:], "\n")  // there will always be one because we add one

	log("comment is: " + l.input[l.pos:l.pos+Pos(i)])
	cleanedComment := strings.TrimSpace(strings.Replace(l.input[l.pos:l.pos+Pos(i)], "//", "", 1))
	splitComment := strings.SplitN(cleanedComment, " ", 2)
	if splitComment[0] == "import" && len(splitComment) > 1 {
		defineImportedTypes(strings.ToLower(splitComment[1]))
	}
	l.pos += Pos(i)
	l.emit(itemComment)
	return lexBb
}

// lexBb scans bb
func lexBb(l *lexer) stateFn {
	log("lexBb")

	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
		return nil
	case isSpace(r):
		l.backup() // Put space back in case we have " -}}".
		return lexSpace
	case r == '|':
		l.emit(itemPipe)
	case r == '"':
		return lexQuote
	case r == '`':
		return lexRawQuote
	case couldBeUDT(r) || isNumeric(r):
		l.backup()
		return lexUDT
	case isNumeric(r):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case r == '/':
		if l.accept("*") {
		  return lexComment
		} else if l.accept("/") {
			return lexInlineComment
		} else {
			return lexIdentifier
		}
	default:
		return lexIdentifier  // all unicode is allowed, so assume everything else is the start of a definition
	}
	return lexBb
}

// lexSpace scans a run of space characters.
// We have not consumed the first space, which is known to be present.
func lexSpace(l *lexer) stateFn {
	log("lexSpace")
	var r rune
	var numSpaces int
	for {
		r = l.peek()
		if !isSpace(r) {
			break
		}
		l.next()
		if r == ' ' {
			numSpaces++
		} else if r == '\n' {
			log("found newline")
			l.acceptRun(" ")  // ignore whitespace at start of next line
			l.emit(itemNewline)
			return lexBb
		} else {
		  numSpaces+=2  // tabs count as 2 spaces
		}
	}

	if numSpaces > 1 {
		l.emit(itemTab)
	}

	l.emit(itemSpace)
	return lexBb
}

// scans an alphanumeric that isn't a udt or a number (could be a definition or bool or string)
func lexIdentifier(l *lexer) stateFn {
	log("lexIdentifier")
Loop:
	for {
		switch r := l.next(); {
		case r != '=' && isUnitChar(r):  // catch assignment with no space before '='
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]

			switch {
			case word == "true", word == "false":
				l.emit(itemBool)
			case word == "null":
				l.emit(itemNull)
			default:
				log("word is " + word)

				// look-ahead for assignment
				l.acceptRun(" ")  // todo: don't consume tabs here if there isn't an assignment
				if l.accept("=") {
					return lexDefinition
				}
				l.emit(itemString)
			}
			break Loop
		}
	}

	return lexBb
}

// lex and parse at the same time - '∆ =' has already been consumed
func lexDefinition(l *lexer) stateFn {
	log("lexDefinition")
	unit := strings.TrimSpace(l.input[l.start:l.pos-1])

	l.acceptRun(" ")

	if !l.accept("{") {
		return l.errorf("Invalid assignment, expected '{")
	}

	l.emit(itemAssignment)  // ignored by the parser, for syntax highlighting only

	props := make(map[string]string)

Loop:
	for {
		switch l.next() {
		case eof:
			return l.errorf("Expected '}' at the end of type definition") // TODO: more helpful error message
		case '}':
			break Loop
		case ',':
			l.emit(itemAssignment)  // just ',', ignored by the parser, for syntax highlighting only
		case ' ', '\n':
			// absorb
		case '/':  // deal with comments
			if l.peek() == '/' {
				l.backup()
				lexInlineComment(l)
			} else if l.peek() == '*' {
				l.backup()
				lexComment(l)
			}
		default:
			l.backup()
			err, propName, propValue := l.scanProp()
			if err != nil {
				return err
			}
			props[propName] = propValue

		}
	}

	l.emit(itemAssignment)  // just '}', ignored by the parser, for syntax highlighting only


	// special logic: add the definition to the global map of UDTs now - lex it properly later
	//definitionValue := l.input[l.start:l.pos]
	NewUDTFromDefinition(unit, props)

	//l.emit(itemDefinition)

  return lexBb
}

// Extract the prop name and value. Ignores quotes and spaces.
func (l *lexer) scanProp() (stateFn, string, string) {
	log("scanProp")

	start := l.pos - 1
	propName := ""
	propValue := ""
  // TODO: allow commas in quoted value
Loop:
	// look for an unescaped ':' to end of the prop name
	for {
		switch r := l.next(); r {
		case eof, '}':
			return l.errorf("Expected ':' at the end of prop name"), "", ""
		case '"', '`':
			err := l.scanQuotedString(r)
			log("finished scanning quoted string, " + string(l.peek()) + " is next.")
			if err != nil {
				return err, "", ""
			}
		case '\\':
			l.accept("}:")
		case ':':
			l.backup()
			if start == l.pos {
				return l.errorf("Prop name cannot be empty"), "", ""
			}
			propName = l.input[start:l.pos]
			break Loop
		}
	}

	l.emit(itemPropName)  // ignored by the parser, for syntax highlighting only

	l.accept(":")
	l.emit(itemAssignment)  // just ':', ignored by the parser, for syntax highlighting only

	l.acceptRun(" \n")

	// allow comments between name and value
	if l.peek() == '/' {
		if l.peek() == '/' {
			lexInlineComment(l)
		} else if l.peek() == '*' {
			lexComment(l)
		}
	}

	start = l.pos

Loop2:
	// look for an unescaped ',' or '}' to end of the prop value
	for {
		switch r := l.next(); r {
		case eof:
			return l.errorf("Expected '}' at the end of definition"), "", ""
		case '"', '`':
			err := l.scanQuotedString(r)
			if err != nil {
				return err, "", ""
			}
		case '\\':
			l.accept("},")
		case '{':  // possible js code block
			err := l.scanJavaScript()
			if err != nil {
				return err, "", ""
			}
		case ',', '}':
			l.backup()
			if start == l.pos {
				return l.errorf("Prop value cannot be empty"), "", ""
			}
			propValue = l.input[start:l.pos]
			break Loop2
		}
	}
	l.emit(itemPropValue)  // ignored by the parser, for syntax highlighting only

	return nil, propName, propValue
}

// scans a js code block following '{'
func (l *lexer) scanJavaScript() stateFn {
	log("scanJavaScript")

Loop:
	// look for an unescaped '}' to end of the prop name
	for {
		switch r := l.next(); r {
		case eof:
			return l.errorf("Expected '}' at the end of JavaScript")
		case '{':
			err := l.scanJavaScript()  // recurse
			if err != nil {
				return err
			}
		case '"', '\'', '`':  // don't end for } in string
			err := l.scanQuotedString(r)
			if err != nil {
				return err
			}
		case '}':
			break Loop
		}
	}
	return nil
}

func (l *lexer) scanQuotedString(quoteChar rune) stateFn {
	log("scanQuotedString")
	log("started scanning quoted string, " + string(l.peek()) + " is next.")

Loop:
	for {
		switch l.next() {
		case eof:
			return l.errorf("Expected '" + string(quoteChar) + "', found EOF")
			// absorb
		case '\\':
			if l.next() == quoteChar {
				// absorb escaped quote
				log("found escaped quote")
			} else {
				log("found stray backslash")
				l.backup()  // backslash is absorbed
			}
		case quoteChar:
			break Loop
		}

	}
  return nil
}


// scans something that could be a UDT, an invalid UDT, or a string
// the unit could be multiple different units that start with the same letter
// it could also be a definition
// We have not consumed any characters
func lexUDT(l *lexer) stateFn {
	log("lexUDT")
	start := l.pos

	if isNumeric(l.peek()) {  // if starts with quantity - scan the number then the unit
		if l.scanNumber(true) {
			l.emit(itemUDT)
			return lexBb
		} else {  // not a DT so must be a number - can't be anything else because it starts with a number
		  l.pos = start
			return lexNumber
		}
	} else if l.scanUnit() {  // must start with a unit, or could be string or identifier

		log("DT with no quantity")

		if !l.scanValue() {  // next thing could be a value or nothing
		  log("removing unit from instances")
		  INSTANCES = INSTANCES[:len(INSTANCES)-1]
			l.emit(itemError)  // modifier has an invalid value
			return lexBb
		}
		if !l.scanModifier() {  // next thing could be a modifier or nothing
			log("removing unit from instances")
			INSTANCES = INSTANCES[:len(INSTANCES)-1]
			l.emit(itemError)  // modifier has an invalid value
			return lexBb
		}

		l.emit(itemUDT)
		return lexBb
	} else {  // not a udt - could be string or identifier
		log("started like a UDT but wasn't. " + string(l.peek()) + " is next")
		if isNumeric(l.peek()) {
			if verbose {
			 return l.errorf("This shouldn't happen: DT was found to be a number after scanning for numbers")
			}
			return lexNumber
		} else {
			log("must be a definition or key word")
			return lexIdentifier
		}
	}
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	log("lexNumber")
	if !l.scanNumber(false) {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(itemNumber)
	return lexBb
}

// check if the next few chars could be a UDT unit - backtrack if not
func (l *lexer) scanUnit() bool {
  log("scanUnit")

  start := l.pos
  word := ""

  // find the longest potential unit we can
  // TODO: get the length of the longest unit (or PDT or key word) and only look up to this length to save time
Loop:
	for {
		switch r := l.next(); {
		//case !(isSpace(r) || isNumeric(r) || isQuoteChar(r) || r == '='):  // if non unit character
		case !isSpace(r):
			log(string(r))
			// absorb
		default:
		  l.backup()
			word = l.input[start:l.pos]
			if len(word) == 0 {
				return false
			}
			break Loop
		}
	}

  wordEnd := l.pos
  // now we have the full word we need to make sure it's not a definition or key word
	// look-ahead for assignment
	l.acceptRun(" ")
	if l.accept("=") {
		l.acceptRun(" ")
		if l.accept("{") {
	    l.pos = start  // backtrack
		  return false  // must be a definition
		}
	}
  if word == "true" || word == "false" || word == "null" {
		l.pos = start  // backtrack
		return false  // it's a key word
	}

  l.pos = wordEnd  // backtrack in case we looked ahead too much

  // find the longest unit that matches this word - UDTs take priority over PDTs even if they're shorter
	// e.g. UDTs are `W`. Input is `Wb`. Assumed to be [`W`, `b`].
	log("looking for a unit to match '" + word + "'")

  bestUnit := ""

	for unit := range UDTs {
		if strings.HasPrefix(word, unit) && len(unit) > len(bestUnit) {
			log("it could be UDT '" + unit + "'")
			bestUnit = unit
		}
	}

	if bestUnit == "" {
		for unit := range PDTs {
			if strings.HasPrefix(word, unit) && len(unit) > len(bestUnit) {
				log("it could be PDT '" + unit + "'")
				bestUnit = unit
			}
		}
	}

  if bestUnit != "" {
		log("unit is " + bestUnit)
		// now we know what the unit is, store so we know which units we have later - speeds up parsing
		INSTANCES = append(INSTANCES, bestUnit)
		l.pos = start + Pos(len(bestUnit))  // backtrack to the end of the unit
		return true
	} else {
	  log("it's not a known unit")
		l.pos = start  // backtrack
		return false  // unit found did not match any known unit - could be a string
	}
}

// determine if we have a valid number or udt
func (l *lexer) scanNumber(udt bool) bool {
	startPos := l.pos

	if udt {
	  log("scanUDT (scanNumber)")
	} else {
		log("scanNumber")
	}
	l.accept("-")  // Optional leading sign. bb does not allow leading +
	l.acceptRun("0123456789")
	if l.accept(".") {
		l.acceptRun("0123456789")
	}

	// accept UDT character(s) at the end
	if udt {

		if !l.scanUnit() {
			l.pos = startPos  // reset
			return false  // if there's no unit now then it's not a DT (it's a number)
		}

		if !l.scanValue() {
			log("removing unit from instances")
			INSTANCES = INSTANCES[:len(INSTANCES)-1]
			return false
		} // if there's no value that's fine
		if !l.scanModifier() {
			log("removing unit from instances")
			INSTANCES = INSTANCES[:len(INSTANCES)-1]
			return false
		}  // scans until we get to an unknown modifier (start of something else)
		return true

	} else {  // for numbers only:
		// Next thing mustn't be alphanumeric.
		if isAlphaNumeric(l.peek()) {
			l.pos = startPos  // reset
			return false
		}
		return true
	}
}

// stops when the next character isn't part of the same DT
// returns false if invalid
func (l *lexer) scanModifier() bool {
	log("scanModifier")

	udt := INSTANCES[len(INSTANCES)-1]
	var rawModifiers [][2]string
	MODIFIER_INSTANCES = append(MODIFIER_INSTANCES, &rawModifiers)  // initialise array to store modifiers

	modifierStart := l.pos

  Loop:  // loop for multiple modifier/value pairs
	for {
		r := l.peek()
		log(string(r))

		// read until we hit a non modifier (number, dot followed by number, dash followed by number)
		// then check it's a known modifier - if not then stop
		// then scan value
		//if isSpace(r) || isNumeric(r) || r == '"' || r == '`' || r == '.' || r == '-' {
		if isSpace(r) {
			// TODO: check for dot not followed by number or dash not followed by number or dot then number
			// check it's a known modifier

			m := l.input[modifierStart:l.pos]
			backtrackCharacters := 0

		  LoopBacktrack: for {  // loop for multiple lengths of modifier, i.e. #>? then #> then #

				if len(m) == backtrackCharacters {
					log("nothing matches " + m)
					break LoopBacktrack  // if we've already looked for modifiers of length 1 then give up
				} else {
					m2 := m[:len(m)-backtrackCharacters]

					for modifier := range UDTs[udt].StringProps { // get all the modifiers for the current type
						if modifier == m2 {
							l.pos = l.pos - Pos(backtrackCharacters)
							log("modifier is: \033[92m" + m2 + "\033[0m")
							if !l.scanValue() {
								log("value is invalid")
								return false
							} else {
								log("value is \033[92m" + l.input[modifierStart+Pos(len(m2)):l.pos] + "\033[0m")
								rawModifiers = append(rawModifiers, [2]string{ m2, l.input[modifierStart+Pos(len(m2)):l.pos] })  // store the modifier and value we've found
								// keep going - onto the next modifier
								modifierStart = l.pos
								continue Loop
							}
						}
					}
					log("Couldn't find a match for " + m2)
					// if there are no matches - look for a shorter modifier
					backtrackCharacters += 1
				}
			}

			// if there are still no matches - reset then stop (assume the next character is part of something else)
			l.pos = modifierStart
			return true
		} else {
			l.next()
		}
	}

	return true
}

// values of modifiers can be numbers, quoted strings, or structures TODO JSON (structure)
// returns false if invalid
func (l *lexer) scanValue() bool {
	log("scanValue")

	start := l.pos

	quoted := false
	quoteChar := l.peek()
	if quoteChar == '"' || quoteChar == '`' {
		quoted = true
		l.next()
	}

	isDecimal := false

Loop:
	for {
		switch r := l.next(); {
		case quoted && r == '\\':
			if l.next() == quoteChar {
				// absorb escaped quote
				log("found escaped quote")
			} else {
				log("found stray backslash")
				l.backup()  // backslash is absorbed
			}
		case quoted && r != quoteChar && r != eof:
			// absorb
		case r == '-':
			if l.pos != start + 1 {  // absorb but only at the start
			  l.backup()
				break Loop  // not a valid number but could be a modifier
			}
		case r == '.':
			if isDecimal {
        l.backup()
        break Loop  // not a valid number but could be a modifier
			}
			isDecimal = true  // absorb but only once
		case unicode.IsNumber(r):
			// absorb
		default:
			if quoted {
        if r != quoteChar {
					log("Unescaped value! Invalid!")
					return false
				}
        break Loop
			}
			l.backup()
			break Loop
		}
	}

	if l.input[start:l.pos] == "-" || l.input[start:l.pos] == "." {
		l.backup()  // check value is not invalid number - if so, remove it
	}

	if l.pos > start && l.input[l.pos-1:l.pos] == "." {  // don't allow numbers like '2.' as these could break modifiers starting with '.'
		l.backup()  // remove decimal point
	}

	return true
}

// lexQuote scans a quoted string.
func lexQuote(l *lexer) stateFn {
	log("lexQuote")

Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}
	l.emit(itemString)
	return lexBb
}

// lexRawQuote scans a raw quoted string.
func lexRawQuote(l *lexer) stateFn {
	log("lexRawQuote")

Loop:
	for {
		switch l.next() {
		case eof:
			return l.errorf("unterminated raw quoted string")
		case '`':
			break Loop
		}
	}
	l.emit(itemString)
	return lexBb
}

// only used for seeing if a prop should be included in the output - standard modifier chars are not
func isModifierChar(r rune) bool {
	return strings.ContainsRune(modifiers, r)
}

func isQuoteChar(r rune) bool {
	return strings.ContainsRune(quotes, r)
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// is character valid in a unit? i.e. not a space, or number
func isUnitChar(r rune) bool {
	return !isSpace(r) && !unicode.IsDigit(r)
}

func isNumeric(r rune) bool {
	return unicode.IsNumber(r) || r == '.' || r == '-'

}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func Preview(input string) {
	l := lex(input)
	idx := 0  // keep track of which udt we're looking at
	for item := range l.items {

		colour := ""  // https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
		if item.typ == itemUDT {
			colour = "37"

		} else if item.typ == itemString {
			colour = "92"
		} else if item.typ == itemNumber {
			colour = "96"
		} else if item.typ == itemAssignment {
			colour = "30"
		} else if item.typ == itemPropName {
			colour = "32"
		} else if item.typ == itemPropValue {
			colour = "34"
		} else if item.typ == itemBool {
			colour = "95"
		} else if item.typ == itemNull {
			colour = "95"
		} else if item.typ == itemError {
			colour = "91"
		} else if item.typ == itemComment {
			colour = "90"
		}

		if item.typ == itemUDT {

			unit := INSTANCES[idx]
			idx += 1
			halves := strings.SplitN(item.val, unit, 2)  // split into quantity and everything else

			fmt.Print("\033[", colour, "m", halves[0], "\033[1m", "\033[94m", unit, "\033[0m", "\033[", colour, "m", halves[1], "\033[0m")

		} else {
		  fmt.Print("\033[", colour, "m", item.val, "\033[0m")
		}
	}
}

// return all items from the input and what colour they should be as a JSON object
func Syntax(input string) map[string]interface{} {
	l := lex(input)

	classes := make([]interface{}, 0)
	output := make([]interface{}, 0)

	for item := range l.items {

		switch item.typ {
		  case itemUDT:

				unit := INSTANCES[instanceIdx]
				//udt := UDTs[unit]
				data := ParseUDT(item.val)
				udt := make([]interface{}, 0)
				log(item.val)

				halves := strings.SplitN(item.val, unit, 2)  // split into quantity and everything else
				quantity := halves[0]
				udt = append(udt, map[string]interface{}{ "class": "quantity", "value": quantity })
				udt = append(udt, map[string]interface{}{ "class": "unit", "value": unit })
				// TODO: split everything else into modifiers and values
				if len(halves) > 1 {
				  udt = append(udt, map[string]interface{}{ "class": "value", "value": halves[1] })
				}
				// TODO: modifiers
				//for modifierUnit, modifierValue := range(modifiers) {
			  //  output = append(output, map[string]interface{}{ "class": "modifier modifier-" + modifierUnit + " modifierUnit", "value": modifierValue })
			  //  output = append(output, map[string]interface{}{ "class": "modifier modifier-" + modifierUnit + " modifierValue", "value": modifierValue })
		  	//}

		  	output = append(output, map[string]interface{}{ "class": "UDT UDT-" + unit, "value": udt, "data": data })

		  case itemString:
				output = append(output, map[string]interface{}{ "class": "string", "value": item.val })
			case itemNumber:
			  output = append(output, map[string]interface{}{ "class": "number", "value": item.val })
		  case itemAssignment:
			  output = append(output, map[string]interface{}{ "class": "assignment", "value": item.val })
		  case itemPropName:
			  output = append(output, map[string]interface{}{ "class": "propName", "value": item.val })
		  case itemPropValue:
			  output = append(output, map[string]interface{}{ "class": "propValue", "value": item.val })
		  case itemBool:
			  output = append(output, map[string]interface{}{ "class": "bool", "value": item.val })
		  case itemNull:
			  output = append(output, map[string]interface{}{ "class": "null", "value": item.val })
		  case itemError:
			  output = append(output, map[string]interface{}{ "class": "error", "value": item.val, "error": item.message })
		  case itemComment:
			  output = append(output, map[string]interface{}{ "class": "comment", "value": item.val })
		  case itemEOF:
			  // do nothing
			default:
        output = append(output, item.val)
		}

	}

	return map[string]interface{}{ "classes": classes, "items": output }
}
