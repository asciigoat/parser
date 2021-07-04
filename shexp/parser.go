package shexp

import (
	"fmt"
	"io"
	"log"
	"sync"

	"go.sancus.dev/core/errors"

	"asciigoat.org/core/lexer"
	"asciigoat.org/core/runes"
)

type Parser struct {
	sync.Mutex // lock

	exp  *Expander     // resolver
	in   *runes.Feeder // input
	done chan bool     // signals parser has finished
}

func (p *Parser) Execute() (string, error) {
	p.Start()

	p.Lock()
	defer p.Unlock()

	<-p.done
	return "", nil
}

func (p *Parser) Start() {
	p.Lock()
	defer p.Unlock()

	if p.done == nil {
		// flag
		p.done = make(chan bool)

		// lexer
		lex := p.newLexer()
		go lex.Run()

		// parser
		go p.run(lex)
	}
}

func (p *Parser) run(lex lexer.Lexer) {
	defer close(p.done)

	for {
		select {
		case t, ok := <-lex.Tokens():
			if t != nil {
				p.process(t)
			} else if !ok {
				break
			}
		}
	}

}

func (p *Parser) process(t lexer.Token) {
	log.Printf("%#v: %#v", errors.Here(), t)
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
