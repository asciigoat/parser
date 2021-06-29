// shexp provides shell-like varianble expansions
package shexp

import (
	"io"
)

// A Expander is an object that can be used to Expand strings
type Expander struct {
	resolver Resolver
}

// NewExpander allocates a new Expander using a
// given Resolver, or os.Getenv if none is given
func NewExpander(r Resolver) *Expander {
	if r == nil {
		r = &EnvResolver{}
	}

	return &Expander{
		resolver: r,
	}
}

// Get resolves a variable names as the given Expander
// would do
func (exp *Expander) Get(key string) string {
	return exp.resolver.Get(key)
}

// Set sets a variable to a given value for future use
// by the expander
func (exp *Expander) Set(key, value string) string {
	return exp.resolver.Set(key, value)
}

// Reset attempts to reset the Resolver if supported
func (exp *Expander) Reset() {
	if r, ok := exp.resolver.(interface {
		Reset()
	}); ok {
		r.Reset()
	}
}

// ExpandString expands variables on a given string
func (exp *Expander) ExpandString(s string, args ...interface{}) (string, error) {
	if len(s) > 0 {
		return exp.NewStringParser(s, args...).Execute()
	}

	return "", nil
}

// ExpandBytes expands variables on a given text as []byte
func (exp *Expander) ExpandBytes(b []byte) (string, error) {
	if len(b) > 0 {
		return exp.NewBytesParser(b).Execute()
	}

	return "", nil
}

// Expand expands variables on the text provided by a Reader
func (exp *Expander) Expand(in io.Reader) (string, error) {
	if in != nil {
		return exp.NewParser(in).Execute()
	}

	return "", nil
}
