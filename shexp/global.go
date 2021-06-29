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
