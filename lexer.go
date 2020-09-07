package icu

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func newLexer(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan token),
	}
	go l.run()
	return l
}

type lexer struct {
	input string
	state stateFn
	depth int
	pos   int
	start int
	width int
	items chan token
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexMessage; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) emit(t tokenCategory) {
	n := token{t, l.input[l.start:l.pos]}
	l.items <- n
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
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
	l.items <- token{tokenError, fmt.Sprintf(format, args...)}
	return nil
}

// nextToken returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextToken() token {
	item := <-l.items
	return item
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

const eof = -1

type stateFn func(*lexer) stateFn

func lexMessage(l *lexer) stateFn {
	r := l.next()
	switch r {
	case hash:
		l.backup()
		if l.pos > l.start {
			l.emit(tokenText)
		}
		l.next()
		l.emit(tokenHash)
		return lexMessage
	case quote:
		n := l.peek()
		switch n {
		case leftDelim, rightDelim, quote:
			l.backup()
			if l.pos > l.start {
				l.emit(tokenText)
			}
			l.next()
			l.ignore()
			return lexQuote
		default:
			return lexMessage
		}
	case leftDelim:
		l.backup()
		if l.pos > l.start {
			l.emit(tokenText)
		}
		return lexLeftDelim
	case rightDelim:
		l.backup()
		if l.pos > l.start {
			l.emit(tokenText)
		}
		return lexRightDelim
	case eof:
		if l.pos > l.start {
			l.emit(tokenText)
		}
		l.emit(tokenEOF)
		return nil
	default:
		return lexMessage
	}
}

func lexQuote(l *lexer) stateFn {
	r := l.next()
	switch r {
	case quote:
		l.backup()
		l.emit(tokenQuotedText)
		l.next()
		l.ignore()
		if p := l.next(); p == quote {
			return lexQuote
		} else {
			l.backup()
		}
		return lexMessage
	default:
		return lexQuote
	}
}

func lexLeftDelim(l *lexer) stateFn {
	l.pos += 1
	l.depth++
	if l.depth%2 == 1 {
		l.emit(tokenStartAction)
		return lexAction
	}
	l.emit(tokenStartMessage)
	return lexMessage
}

func lexRightDelim(l *lexer) stateFn {
	l.pos += 1
	l.depth--
	if l.depth%2 == 1 {
		l.emit(tokenEndMessage)
		return lexAction
	}
	l.emit(tokenEndAction)
	return lexMessage
}

func lexDelim(l *lexer) stateFn {
	l.pos += 1
	l.emit(tokenDelim)
	return lexAction
}

func lexAction(l *lexer) stateFn {
	switch r := l.next(); {
	case r == leftDelim:
		l.backup()
		return lexLeftDelim
	case isSpace(r):
		return lexSpace
	case isAlphaNumeric(r):
		return lexIdentifier
	case r == delim:
		l.backup()
		return lexDelim
	case r == rightDelim:
		l.backup()
		return lexRightDelim
	}
	return lexAction
}

func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(tokenSpace)
	return lexAction
}

func lexIdentifier(l *lexer) stateFn {
	switch r := l.next(); {
	case isAlphaNumeric(r):
		return lexIdentifier
	default:
		l.backup()
		l.emit(tokenIdentifier)
		return lexAction
	}
}

type token struct {
	cat tokenCategory
	val string
}

func (t token) String() string {
	switch {
	case t.cat == tokenEOF:
		return "EOF"
	case t.cat == tokenError:
		return fmt.Sprintf("Error(%s)", t.val)
	case t.cat == tokenText:
		return fmt.Sprintf("Text(%s)", t.val)
	case t.cat == tokenIdentifier:
		return fmt.Sprintf("Identifier(%s)", t.val)
	case t.cat == tokenSpace:
		return fmt.Sprintf("Space(%s)", t.val)
	case t.cat == tokenDelim:
		return fmt.Sprintf("Delimiter(%s)", t.val)
	case t.cat == tokenQuote:
		return fmt.Sprintf("Quote(%s)", t.val)
	case t.cat == tokenQuotedText:
		return fmt.Sprintf("QuotedText(%s)", t.val)
	case t.cat == tokenStartAction:
		return fmt.Sprintf("StartAction(%s)", t.val)
	case t.cat == tokenEndAction:
		return fmt.Sprintf("EndAction(%s)", t.val)
	case t.cat == tokenStartMessage:
		return fmt.Sprintf("StartMessage(%s)", t.val)
	case t.cat == tokenEndMessage:
		return fmt.Sprintf("EndMessage(%s)", t.val)
	case len(t.val) > 10:
		return fmt.Sprintf("%.10q...", t.val)
	}
	return fmt.Sprintf("%q", t.val)
}

type tokenCategory int

const (
	tokenError tokenCategory = iota
	tokenEOF
	tokenText
	tokenIdentifier
	tokenSpace
	tokenDelim
	tokenQuote
	tokenQuotedText
	tokenStartAction
	tokenEndAction
	tokenStartMessage
	tokenEndMessage
	tokenHash
)

const (
	leftDelim      = '{'
	rightDelim     = '}'
	delim          = ','
	quote          = '\''
	colon          = ':'
	hash           = '#'
	space          = ' '
	newLine        = '\n'
	carriageReturn = '\r'
)

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\v' || r == '\f'
}

func isNewLine(r rune) bool {
	return r == '\n' || r == '\r'
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || r == '=' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
