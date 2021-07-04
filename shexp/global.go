package shexp

import (
	"io"
)

var envExpander = NewExpander(nil)

func ExpandString(s string, args ...interface{}) (string, error) {
	return envExpander.ExpandString(s, args...)
}

func ExpandBytes(b []byte) (string, error) {
	return envExpander.ExpandBytes(b)
}

func Expand(in io.Reader) (string, error) {
	return envExpander.Expand(in)
}

func NewStringParser(s string, args ...interface{}) *Parser {
	return envExpander.NewStringParser(s, args...)
}

func NewBytesParser(b []byte) *Parser {
	return envExpander.NewBytesParser(b)
}

func NewParser(in io.Reader) *Parser {
	return envExpander.NewParser(in)
}
