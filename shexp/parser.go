package shexp

import (
	"fmt"
	"io"

	"asciigoat.org/core/runes"
)

type Parser struct {
	exp *Expander
	in  *runes.Feeder
}

func (p *Parser) Execute() (string, error) {
	return "", nil
}

//
// Constructor
//
func (exp *Expander) NewStringParser(s string, args ...interface{}) *Parser {
	if len(s) > 0 {
		if len(args) > 0 {
			s = fmt.Sprintf(s, args...)
		}
	}

	return &Parser{
		exp: exp,
		in:  runes.NewFeederString(s),
	}
}

func (exp *Expander) NewBytesParser(b []byte) *Parser {
	return &Parser{
		exp: exp,
		in:  runes.NewFeederBytes(b),
	}
}

func (exp *Expander) NewParser(in io.Reader) *Parser {
	return &Parser{
		exp: exp,
		in:  runes.NewFeeder(in),
	}
}
