// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
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
	itemChar                         // printable ASCII character; grab bag for comma etc.
	itemCharConstant                 // character constant
	itemComment                      // comment text
	itemComplex                      // complex constant (1+2i); imaginary is just a number
	itemAssign                       // equals ('=') introducing an assignment
	itemDeclare                      // colon-equals (':=') introducing a declaration
	itemEOF
	itemField      // alphanumeric identifier starting with '.'
	itemIdentifier // alphanumeric identifier not starting with '.'
	itemLeftDelim  // left action delimiter
	itemLeftParen  // '(' inside action
	itemNumber     // simple number, including imaginary
	itemPipe       // pipe symbol
	itemRawString  // raw quoted string (includes quotes)
	itemRightDelim // right action delimiter
	itemRightParen // ')' inside action
	itemSpace      // run of spaces separating arguments
	itemString     // quoted string (includes quotes)
	itemText       // plain text
	itemVariable   // variable starting with '$', such as '$' or  '$1' or '$hello'
	// Keywords appear after all the rest.
	itemKeyword  // used only to delimit the keywords
	itemBlock    // block keyword
	itemDot      // the cursor, spelled '.'
	itemDefine   // define keyword
	itemElse     // else keyword
	itemEnd      // end keyword
	itemIf       // if keyword
	itemNil      // the untyped nil constant, easiest to treat as a keyword
	itemRange    // range keyword
	itemTemplate // template keyword
	itemWith     // with keyword
	itemUDT      // user defined type
	itemTab      // two or more spaces or a tab
	itemDefinition  // type definition, e.g. ∆ = { unit: pizza }
	itemNewline  // \n
)

var itemNames = []string{
"itemError",
"Bool",         // boolean constant
"Char",         // printable ASCII character; grab bag for comma etc.
"CharConstant", // character constant
"Comment",                     // comment text
"Complex",                     // complex constant (1+2i); imaginary is just a number
"Assign",                      // equals ('=') introducing an assignment
"Declare",                     // colon-equals (':=') introducing a declaration
"EOF",
"Field",      // alphanumeric identifier starting with '.'
"Identifier", // alphanumeric identifier not starting with '.'
"LeftDelim",  // left action delimiter
"LeftParen",  // '(' inside action
"Number",     // simple number, including imaginary
"Pipe",       // pipe symbol
"RawString",  // raw quoted string (includes quotes)
"RightDelim", // right action delimiter
"RightParen", // ')' inside action
"space",      // run of spaces separating arguments
"String",    // quoted string (includes quotes)
"Text",      // plain text
"Variable",  // variable starting with '$', such as '$' or  '$1' or '$hello'
// Keywords appear after all the rest.
"Keyword", // used only to delimit the keywords
"Block",   // block keyword
"Dot",     // the cursor, spelled '.'
"Define",  // define keyword
"Else",    // else keyword
"End",     // end keyword
"If",      // if keyword
"Nil",      // the untyped nil constant, easiest to treat as a keyword
"Range",    // range keyword
"Template", // template keyword
"With",     // with keyword
"UDT",      // user defined type
"tab",      // two or more spaces or a tab
"definition",
"Newline",  // \n
}

var key = map[string]itemType{
	".":        itemDot,
	"block":    itemBlock,
	"define":   itemDefine,
	"else":     itemElse,
	"end":      itemEnd,
	"if":       itemIf,
	"range":    itemRange,
	"nil":      itemNil,
	"template": itemTemplate,
	"with":     itemWith,
}

const eof = -1

// Trimming spaces.
// If the action begins "{{- " rather than "{{", then all space/tab/newlines
// preceding the action are trimmed; conversely if it ends " -}}" the
// leading spaces are trimmed. This is done entirely in the lexer; the
// parser never sees it happen. We require an ASCII space (' ', \t, \r, \n)
// to be present to avoid ambiguity with things like "{{-3}}". It reads
// better with the space present anyway. For simplicity, only ASCII
// does the job.
const (
	spaceChars    = " \t\r\n"  // These are the space characters defined by Go itself.
	trimMarker    = '-'        // Attached to left/right delimiter, trims trailing spaces from preceding/following text.
	trimMarkerLen = Pos(1 + 1) // marker plus space before or after
	modifiers     = "+-~=<>:;/|#&≠≥≤^"  // all potential modifiers
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
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.startLine}
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
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.startLine}
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

// lexText scans until an opening action delimiter, "{{".
func lexText(l *lexer) stateFn {
	verbose_print("lexText")

	l.width = 0
	//if x := strings.Index(l.input[l.pos:], l.leftDelim); x >= 0 {
	//	ldn := Pos(len(l.leftDelim))
	//	l.pos += Pos(x)
	//	trimLength := Pos(0)
	//	if hasLeftTrimMarker(l.input[l.pos+ldn:]) {
	//		trimLength = rightTrimLength(l.input[l.start:l.pos])
	//	}
	//	l.pos -= trimLength
	//	if l.pos > l.start {
	//		l.line += strings.Count(l.input[l.start:l.pos], "\n")
	//		l.emit(itemText)
	//	}
	//	l.pos += trimLength
	//	l.ignore()
	//	return lexLeftDelim
	//}
	l.pos = Pos(len(l.input))
	verbose_print("EOF")
	// Correctly reached EOF.
	if l.pos > l.start {
		l.line += strings.Count(l.input[l.start:l.pos], "\n")
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

// rightTrimLength returns the length of the spaces at the end of the string.
func rightTrimLength(s string) Pos {
	return Pos(len(s) - len(strings.TrimRight(s, spaceChars)))
}

//// atRightDelim reports whether the lexer is at a right delimiter, possibly preceded by a trim marker.
//func (l *lexer) atRightDelim() (delim, trimSpaces bool) {
//	if hasRightTrimMarker(l.input[l.pos:]) && strings.HasPrefix(l.input[l.pos+trimMarkerLen:], l.rightDelim) { // With trim marker.
//		return true, true
//	}
//	if strings.HasPrefix(l.input[l.pos:], l.rightDelim) { // Without trim marker.
//		return true, false
//	}
//	return false, false
//}

// leftTrimLength returns the length of the spaces at the beginning of the string.
func leftTrimLength(s string) Pos {
	return Pos(len(s) - len(strings.TrimLeft(s, spaceChars)))
}

//// lexLeftDelim scans the left delimiter, which is known to be present, possibly with a trim marker.
//func lexLeftDelim(l *lexer) stateFn {
//	verbose_print("lexLeftDelim")
//	l.pos += Pos(len(l.leftDelim))
//	trimSpace := hasLeftTrimMarker(l.input[l.pos:])
//	afterMarker := Pos(0)
//	if trimSpace {
//		afterMarker = trimMarkerLen
//	}
//	if strings.HasPrefix(l.input[l.pos+afterMarker:], leftComment) {
//		l.pos += afterMarker
//		l.ignore()
//		return lexComment
//	}
//	l.emit(itemLeftDelim)
//	l.pos += afterMarker
//	l.ignore()
//	l.parenDepth = 0
//	return lexBb
//}

// lexComment scans a comment. The left comment marker is known to be present.
func lexComment(l *lexer) stateFn {
	verbose_print("lexComment")
	l.pos += Pos(len(leftComment))
	i := strings.Index(l.input[l.pos:], rightComment)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i + len(rightComment))
	//delim, trimSpace := l.atRightDelim()
	//if !delim {
	//	return l.errorf("comment ends before closing delimiter")
	//}
	//if trimSpace {
	//	l.pos += trimMarkerLen
	//}
	//l.pos += Pos(len(l.rightDelim))
	//if trimSpace {
	//	l.pos += leftTrimLength(l.input[l.pos:])
	//}
	l.ignore()
	return lexBb
}

//// lexRightDelim scans the right delimiter, which is known to be present, possibly with a trim marker.
//func lexRightDelim(l *lexer) stateFn {
//	trimSpace := hasRightTrimMarker(l.input[l.pos:])
//	if trimSpace {
//		l.pos += trimMarkerLen
//		l.ignore()
//	}
//	l.pos += Pos(len(l.rightDelim))
//	l.emit(itemRightDelim)
//	if trimSpace {
//		l.pos += leftTrimLength(l.input[l.pos:])
//		l.ignore()
//	}
//	return lexText
//}

// lexBb scans bb
func lexBb(l *lexer) stateFn {
	verbose_print("lexBb")

	//afterMarker := Pos(0)
	//if strings.HasPrefix(l.input[l.pos+afterMarker:], leftComment) {
	//	l.pos += afterMarker
	//	l.ignore()
	//	return lexComment
	//}

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
	case r == '$':
		return lexVariable
	case r == '\'':
		return lexChar
	//case r == '.':
		// special look-ahead for ".field" so we don't break l.backup().
		//if l.pos < Pos(len(l.input)) {
		//	r := l.input[l.pos]
		//	if r < '0' || '9' < r {
		//		return lexField
		//	}
		//}
	//	fallthrough // '.' can start a number.
	// all user defined typed values and numbers
	case couldBeUDT(r) ||  r == '+' || r == '-' || ('0' <= r && r <= '9') || r == '.':
    // TODO: if it's an invalid udt it could be a string
		l.backup()  // do not consume the pizza
		return lexUDT
	case r == '+' || r == '-' || ('0' <= r && r <= '9') || r == '.':
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case r == '(':
		l.emit(itemLeftParen)
		l.parenDepth++
	case r == ')':
		l.emit(itemRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren %#U", r)
		}
	case r == '/':
		if l.accept("*") {
		  return lexComment
		} else {
			return lexIdentifier  // TODO: this might be invalid
		}
	case r <= unicode.MaxASCII && unicode.IsPrint(r):
		l.emit(itemChar)
	default:
		return lexIdentifier  // all unicode is allowed, so assume everything else is the start of a definition
		//return l.errorf("unrecognized character in action: %#U", r)
	}
	return lexBb
}

// lexSpace scans a run of space characters.
// We have not consumed the first space, which is known to be present.
func lexSpace(l *lexer) stateFn {
	verbose_print("lexSpace")
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
			verbose_print("found newline")
			l.emit(itemNewline)
			l.acceptRun(" ")  // ignore whitespace at start of next line
			l.ignore()
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

// lexIdentifier scans an alphanumeric that isn't a udt or a number (could be a definition or bool)
func lexIdentifier(l *lexer) stateFn {
	verbose_print("lexIdentifier")
Loop:
	for {
		switch r := l.next(); {
		case isUnitChar(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch {
			case key[word] > itemKeyword:
				l.emit(key[word])
			case word[0] == '.':
				l.emit(itemField)
			case word == "true", word == "false":
				l.emit(itemBool)
			default:
				// look-ahead for assignment
				l.acceptRun(" ")  // todo: don't consume tabs here if there isn't an assignment
				if l.accept("=") {  // todo: make '=' optional
					return lexDefinition
				}
				l.emit(itemIdentifier)
			}
			break Loop
		}
	}

	return lexBb
}

// '∆ =' has already been consumed
func lexDefinition(l *lexer) stateFn {
	verbose_print("lexDefinition")
	l.acceptRun(" ")
	if !l.accept("{") {
		l.errorf("Invalid assignment, expected '{")
	}
	// look-ahead for } before the next newline
Loop:
	for {
		switch l.next() {
		case '\n', eof:
			return l.errorf("Expected '}' at the end of type definition")  // TODO: more helpful error message
		case '\\':
			l.accept("}")
		case '}':
			break Loop
		}
	}

	// special logic: add the definition to the global map of UDTs now - lex it properly later
	definitionValue := l.input[l.start:l.pos]
	NewUDTFromDefinition(definitionValue)

	l.emit(itemDefinition)

  return lexBb
}

// lexField scans a field: .Alphanumeric.
// The . has been scanned.
func lexField(l *lexer) stateFn {
	return lexFieldOrVariable(l, itemField)
}

// lexVariable scans a Variable: $Alphanumeric.
// The $ has been scanned.
func lexVariable(l *lexer) stateFn {
	verbose_print("lexVariable")

	if l.atTerminator() { // Nothing interesting follows -> "$".
		l.emit(itemVariable)
		return lexBb
	}
	return lexFieldOrVariable(l, itemVariable)
}

// lexVariable scans a field or variable: [.$]Alphanumeric.
// The . or $ has been scanned.
func lexFieldOrVariable(l *lexer, typ itemType) stateFn {
	verbose_print("lexFieldOrVariable")

	if l.atTerminator() { // Nothing interesting follows -> "." or "$".
		if typ == itemVariable {
			l.emit(itemVariable)
		} else {
			l.emit(itemDot)
		}
		return lexBb
	}
	var r rune
	for {
		r = l.next()
		if !isAlphaNumeric(r) {
			l.backup()
			break
		}
	}
	if !l.atTerminator() {
		return l.errorf("bad character %#U", r)
	}
	l.emit(typ)
	return lexBb
}

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier. Breaks .X.Y into two pieces. Also catches cases
// like "$x+2" not being acceptable without a space, in case we decide one
// day to implement arithmetic.
func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) {
		return true
	}
	switch r {
	case eof, '.', ',', '|', ':', ')', '(':
		return true
	}
	// Does r start the delimiter? This can be ambiguous (with delim=="//", $x/2 will
	// succeed but should fail) but only in extremely rare cases caused by willfully
	// bad choice of delimiter.
	//if rd, _ := utf8.DecodeRuneInString(l.rightDelim); rd == r {
	//	return true
	//}
	return false
}

// lexChar scans a character constant. The initial quote is already
// scanned. Syntax checking is done by the parser.
func lexChar(l *lexer) stateFn {
	verbose_print("lexChar")

Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated character constant")
		case '\'':
			break Loop
		}
	}
	l.emit(itemCharConstant)
	return lexBb
}

// scans something that could be a UDT, an invalid UDT, or a string
// the unit could be multiple different units that start with the same letter
// We have not consumed any characters
func lexUDT(l *lexer) stateFn {
	verbose_print("lexUDT")

	if !l.scanUDT() {
		return lexNumber  // if it can't be a UDT then it's probably a number
	}
	l.emit(itemUDT)

	return lexBb
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	verbose_print("lexNumber")
	if !l.scanNumber(false) {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(itemNumber)
	return lexBb
}

func (l *lexer) scanUDT() bool {

	if l.scanUnit() {  // if starts with a unit then only modifiers can come next
		verbose_print("udt with no value")
		if !isSpace(l.peek()) {
			// next thing could be a modifier, else it's invalid  TODO: right side value
			return l.scanModifier()
		}
		return true
	}
  return l.scanNumber(true)
}

// check if the next few chars could be a UDT unit
func (l *lexer) scanUnit() bool {
  verbose_print("scanUnit")

  // there could be left arguments already scanned - make sure they're ignored
  start := l.pos

Loop:  // keep going through until there's a
	for {
		switch r := l.next(); {
		case !(isSpace(r) || unicode.IsDigit(r) || isModifierChar(r)):  // if non unit character // TODO modifier chars
			verbose_print(string(r))
			// absorb
		default:
		  l.backup()
			word := l.input[start:l.pos]
			if len(word) == 0 {
				return false
			}

			verbose_print("unit is " + word)

			for unit := range UDTs {
				verbose_print("does " + word + " == " + unit + "?")

				if word == unit {
					verbose_print("it's a known unit")

					// now we know what the unit is, store so we know which units we have later - speeds up parsing
					INSTANCES = append(INSTANCES, unit)
					return true
				}
			}
			break Loop
		}
	}
	verbose_print("it's not a known unit")

	return false  // unit found did not match known unit - could be a string
}

// determine if we have a valid number or udt
func (l *lexer) scanNumber(udt bool) bool {
	if udt {
	  verbose_print("scanUDT")
	} else {
		verbose_print("scanNumber")
	}
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789_"
	if l.accept("0") {
		// Note: Leading 0 does not mean octal in floats.
		if l.accept("xX") {
			digits = "0123456789abcdefABCDEF_"
		} else if l.accept("oO") {
			digits = "01234567_"
		} else if l.accept("bB") {
			digits = "01_"
		}
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if len(digits) == 10+1 && l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	if len(digits) == 16+6+1 && l.accept("pP") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	// accept UDT character(s) at the end
	if udt {

		if !l.scanUnit() {
			return false  // if there's no UDT unit now then it's not a UDT
		}

		if !isSpace(l.peek()) {
			// next thing could be a modifier, else it's invalid
			return l.scanModifier()
		}

		return true

	} else {  // for numbers only:
		// Next thing mustn't be alphanumeric.
		if isAlphaNumeric(l.peek()) {
			l.next()
			return false
		}
		return true
	}
}

func (l *lexer) scanModifier() bool {
	verbose_print("scanModifier")

Loop:  // loop for multiple modifier/value pairs
	for {
		if isSpace(l.peek()) { // end of UDT
			break Loop
		}

		if l.accept(modifiers) {
			// it's a value
			l.scanValue()
		} else {  // invalid
			l.errorf("Expected a modifier, found '" + string(l.peek()) + "', invalid!")
			return false
		}

		//// if a value comes next then it's valid, else it's value is 1
		//if isAlphaNumeric(l.peek()) {
		//
		//  return true
		//}
		//return true  // value is

	}
	return true
}

// values of modifiers can be numbers, strings, quoted strings, UDTs?
func (l *lexer) scanValue() bool {
	verbose_print("scanValue")

Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):  // TODO: quoted strings
			// absorb.
		//case l.scanUDT("∆"): TODO
			// absorb.
		default:
			l.backup()
			break Loop
		}
	}

	return true
}

// lexQuote scans a quoted string.
func lexQuote(l *lexer) stateFn {
	verbose_print("lexQuote")

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
	verbose_print("lexRawQuote")

Loop:
	for {
		switch l.next() {
		case eof:
			return l.errorf("unterminated raw quoted string")
		case '`':
			break Loop
		}
	}
	l.emit(itemRawString)
	return lexBb
}

func isModifierChar(r rune) bool {
	return strings.ContainsRune(modifiers, r)
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// is character valid in a unit? i.e. not a space, modifier char, or number
func isUnitChar(r rune) bool {
	return !isSpace(r) && !unicode.IsDigit(r) && !isModifierChar(r)
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func hasLeftTrimMarker(s string) bool {
	return len(s) >= 2 && s[0] == trimMarker && isSpace(rune(s[1]))
}

func hasRightTrimMarker(s string) bool {
	return len(s) >= 2 && isSpace(rune(s[0])) && s[1] == trimMarker
}
