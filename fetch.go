// Package Fetch allows the querying of nested data through javascript-style accessors

package fetch

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	itemError = iota
	itemBeginArray
	itemEndArray
	itemString
	itemNumber
	itemDot
	itemField
	itemSpace
	fieldDot
	fieldMap
	fieldArray
)

var ident = map[int]string{
	itemError:      "itemError",
	itemBeginArray: "itemBeginArray",
	itemEndArray:   "itemEndArray",
	itemString:     "itemString",
	itemNumber:     "itemNumber",
	itemDot:        "itemDot",
	itemField:      "itemField",
	itemSpace:      "itemSpace",
	fieldDot:       "fieldDot",
	fieldMap:       "fieldMap",
	fieldArray:     "fieldArray",
}

const eof = -1

type itemType int
type stateFn func(*Query) stateFn

type item struct {
	typ itemType
	pos int
	val string
}

type fieldType int

type field struct {
	typ   fieldType
	index int
	key   string
}

type Query struct {
	state   stateFn
	pos     int
	width   int
	input   string
	start   int
	lastPos int
	items   chan item
	fields  []field
}

func (l *Query) run() {
	for l.state = startLex; l.state != nil; {
		l.state = l.state(l)
	}
}

func (l *Query) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *Query) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Query) backup() {
	l.pos -= l.width
}

func (l *Query) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *Query) ignore() {
	l.start = l.pos
}

func (l *Query) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

func (l *Query) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Query) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *Query) String() string {
	return l.input
}

func startLex(l *Query) stateFn {
	c := l.next()
	switch {
	case c == '[':
		l.emit(itemBeginArray)
		return startLex
	case c == ']':
		l.emit(itemEndArray)
		return startLex
	case c == '"':
		return lexQuote
	case c == '\'':
		return lexSQuote
	case c == '.':
		if !isAlphaNumeric(l.peek()) {
			l.emit(itemDot)
		} else {
			return lexField
		}
		return startLex
	case '0' <= c && c <= '9':
		l.backup()
		return lexNumber
	case c == eof:
		l.emit(eof)
		return nil
	case isAlphaNumeric(c):
		l.emit(itemError)
	case !isAlphaNumeric(c):
		l.emit(itemError)
		return startLex
	}

	return startLex
}

func lexField(l *Query) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch {
			case word[0] == '.':
				l.emit(itemField)
			default:
				l.emit(itemError)
			}
			break Loop
		}
	}
	return startLex
}

func lexQuote(l *Query) stateFn {
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
	return startLex
}

func lexSQuote(l *Query) stateFn {
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
		case '\'':
			break Loop
		}
	}
	l.emit(itemString)
	return startLex
}

func lexNumber(l *Query) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(itemNumber)
	return startLex
}

func lexSpace(l *Query) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(itemSpace)
	return startLex
}

func (l *Query) atTerminator() bool {
	r := l.peek()
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	switch r {
	case eof, '.', ',', '|', ':', ')', '(', '[', ']', '{', '}', '+', '-', '/', '*':
		return true
	}
	return false
}

func (l *Query) runField() error {
	accessor := false
	pos := 0
	var i *field
	for pos < len(l.input) {
		c := l.nextItem()

		switch c.typ {
		case itemField:
			l.fields = append(l.fields, field{typ: fieldMap, key: c.val[1:]})
		case itemBeginArray:
			if accessor {
				return errors.New(fmt.Sprintf("Unexpected token %s at position %d", c.val, c.pos))
			}
			accessor = true
		case itemString:
			if i != nil || !accessor {
				return errors.New(fmt.Sprintf("Unexpected token %s at position %d", c.val, c.pos))
			}

			k := c.val[1:]
			k = k[:len(k)-1]
			i = &field{
				typ: fieldMap,
				key: k,
			}
		case itemNumber:
			if i != nil || !accessor {
				return errors.New(fmt.Sprintf("Unexpected token %s at position %d", c.val, c.pos))
			}
			index, err := strconv.Atoi(c.val)
			if err != nil {
				return err
			}

			i = &field{
				typ:   fieldArray,
				index: index,
			}
		case itemEndArray:
			if i == nil || !accessor {
				return errors.New(fmt.Sprintf("Unexpected token %s at position %d", c.val, c.pos))
			}

			l.fields = append(l.fields, *i)
			i = nil
			accessor = false
		case eof:
			return nil
		case itemDot:
			if pos == 0 {
				break
			}
			fallthrough
		default:
			return errors.New(fmt.Sprintf("Unexpected token %s at position %d", c.val, c.pos))
		}
		pos += len(c.val)
	}

	return nil
}

func (l *Query) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

func (l *Query) scanNumber() bool {
	digits := "0123456789"
	l.acceptRun(digits)
	return true
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func mapValue(o interface{}, key string) (interface{}, error) {
	n, ok := o.(map[string]interface{})
	if !ok {
		return nil, errors.New("Not of type object")
	}
	p, ok := n[key]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Key (%s) does not exist", key))
	}
	return p, nil
}

func indexValue(o interface{}, index int) (interface{}, error) {
	n, ok := o.([]interface{})
	if !ok {
		return nil, errors.New("Not of type array")
	}
	if index > len(n) {
		return nil, errors.New(fmt.Sprintf("Index (%d) out of range", index))
	}
	return n[index], nil
}

// Converts a query string into a *Fetch.Query.
// Fetch.Parse is similar to jq, in that in order to reference the base value,
// you must begin a query with '.'
// For example, a query string of '.' will return an entire value, a query string
// of '.foo' will return the value of key foo on the root of the value. Every
// subsequent field can be accessed through javascript-style dot/bracket notation.
// for example, .foo[0] would return the first element of array foo, and
// .["foo"][0] would do the same as well.
func Parse(input string) (*Query, error) {
	l := &Query{
		input:  input,
		items:  make(chan item),
		fields: []field{},
	}

	go l.run()
	err := l.runField()
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Executes a *Fetch.Query on some data. Returns the result of the query.
func Run(l *Query, o interface{}) (interface{}, error) {
	var err error
	for _, v := range l.fields {
		switch v.typ {
		case fieldMap:
			o, err = mapValue(o, v.key)
			if err != nil {
				return nil, err
			}
		case fieldArray:
			o, err = indexValue(o, v.index)
			if err != nil {
				return nil, err
			}
		}
	}
	return o, nil
}

// A convenience function that runs both Parse() and Run() automatically.
// It is highly recommended that you parse your query ahead of time
// with Fetch.Parse() and follow up with Fetch.Run() instead.
func Fetch(input string, obj interface{}) (interface{}, error) {
	l, err := Parse(input)
	if err != nil {
		return nil, err
	}
	return Run(l, obj)
}
