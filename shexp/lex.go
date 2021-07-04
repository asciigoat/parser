package shexp

import (
	"fmt"

	"asciigoat.org/core/lexer"
	"asciigoat.org/core/runes"
)

const (
	TokenChannelDepth = 2
)

type Token struct {
	lexer.Token
}

// simple pretty printing of a token
func (t Token) String() string {
	if et, ok := t.Token.(lexer.ErrorToken); ok {
		if err := et.Unwrap(); err == lexer.EOF {
			return "EOF"
		} else {
			return fmt.Sprintf("Error: %s", et.Error())
		}
	} else if s := t.Token.String(); len(s) > 10 {
		return fmt.Sprintf("%.10q...", s)
	} else {
		return fmt.Sprintf("%q", s)
	}
}

// lexer
func (p *Parser) newLexer() lexer.Lexer {
	return lexer.NewLexer(lexText, p.in, TokenChannelDepth)
}

const (
	t_error     = lexer.TokenError   // token error as defined by the lexer package
	t_text      = t_error + 1 + iota // just text
	t_start                          // start of ${foo} expansion
	t_end                            // end of ${foo} expansion
	t_ident                          // identifier
	t_mode                           // expansion mode
	t_expansion                      // $foo expansion
)

var (
	// new line
	r_nl      = runes.Rune('\n')
	r_cr      = runes.Rune('\r')
	r_crln    = runes.And(r_cr, r_nl)
	r_newLine = runes.Or(r_nl, r_crln)

	// expansion
	r_dollar   = runes.Rune('$')
	r_lbracket = runes.Rune('{')
	r_rbracket = runes.Rune('}')
	r_start    = runes.And(r_dollar, r_lbracket) // t_start
	r_end      = r_rbracket                      // t_end

	// t_mode
	r_colon = runes.Rune(':')
	r_mode  = runes.And(r_colon, runes.Or(runes.Rune('-'), runes.Rune('+'), runes.Rune('=')))

	// escaped $ in t_text
	r_slash      = runes.Rune('\\')
	r_esc_dollar = runes.And(r_slash, r_dollar)

	// t_ident
	r_id_first = runes.If(func(r rune) bool {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r == '_') {
			return true
		}
		return false
	})
	r_id_more = runes.If(func(r rune) bool {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			(r == '_') {
			return true
		}
		return false
	})

	r_ident      = runes.And(r_id_first, runes.Any(r_id_more))
	r_ident_more = runes.AtLeast(1, r_id_more)
)

// lexText collects literal text
//
// t_text* ---> ${foo}
//         `--> $foo
//
func lexText(lex lexer.Lexer) lexer.StateFn {
	for {
		input, err := lex.AtLeast(2)

		if _, ok := r_start(input); ok {
			// t_text '${' ...
			// ^^^^^^
			lex.Emit(t_text)
			// t_text '${' ...
			//         ^^
			lex.Step(2)
			lex.Emit(t_start)
			// t_text '${' ...
			//             ^^^
			return lexExpVar
		} else if _, ok := r_end(input); ok {
			// '${' t_id t_mode t_text '}' t_text
			//                  ...^^^
			lex.Emit(t_text)
			// '${' t_id t_mode t_text '}' t_text
			//                          ^
			lex.Step(1)
			lex.Emit(t_end)
			// '${' t_id t_mode t_text '}' t_text
			//                             ^^^^^^
			continue
		} else if _, ok := r_dollar(input); ok {
			// t_text '$' t_id t_text
			// ^^^^^^
			lex.Emit(t_text)
			// t_text '$' t_id t_text
			//         ^^^...
			lex.Step(1)
			// t_text '$' t_id t_text
			//         ...^^^^
			return lexVar
		} else if s, ok := r_newLine(input); ok {
			// ... \n ...
			//     ^^
			lex.Step(len(s))
			lex.NewLine()
			continue
		} else if s, ok := r_esc_dollar(input); ok {
			// ... \$ ...
			//     ^^
			lex.Step(len(s))
			continue
		} else if len(input) > 0 {
			// ... ? ...
			//     ^
			lex.Step(1)
			continue
		} else {
			// t_text err !
			// ^^^^^^
			lex.Emit(t_text)
			// t_text err !
			//        ^^^
			lex.EmitError(err)
			// t_text err !
			//            ^
			return nil
		}
	}
}

// lexExpVar collects the variable name within a ${foo} expansion
//
func lexExpVar(lex lexer.Lexer) lexer.StateFn {
	input, err := lex.AtLeast(1)

	if s, ok := r_ident(input); ok {
		// t_start t_ident t_mode t_text t_end
		//         ^^^....
		lex.Step(len(s))
		return lexExpVarMore
	} else if len(input) > 0 {
		// t_start WTH !
		//         ^^^
		lex.EmitSyntaxError("invalid var name")
		// t_start WTH !
		//             ^
		return nil
	} else {
		// t_start err !
		//         ^^^
		lex.EmitError(err)
		// t_start err !
		//             ^
		return nil
	}
}

// lexExpVarMode continues collecting t_id for a ${foo} expansion
func lexExpVarMore(lex lexer.Lexer) lexer.StateFn {
	for {
		input, err := lex.AtLeast(1)

		if s, ok := r_ident_more(input); ok {
			// t_start t_ident t_mode ....
			//         ...^^..
			lex.Step(len(s))
			continue
		} else if _, ok := r_end(input); ok {
			// t_start t_ident t_end t_text
			//         ...^^^^
			lex.Emit(t_ident)
			// t_start t_ident t_end t_text
			//                 ^^^^^
			lex.Step(1)
			lex.Emit(t_end)
			// t_start t_ident t_end t_text
			//                       ^^^^^^
			return lexText
		} else if _, ok := r_colon(input); ok {
			// t_start t_ident t_mode ....
			//         ...^^^^
			lex.Emit(t_ident)
			// t_start t_ident t_mode ....
			//                 ^^...
			return lexExpMode
		} else if len(input) > 0 {
			// t_start WTH !
			//         ^^^
			lex.EmitSyntaxError("invalid var name")
			// t_start WTH !
			//             ^
			return nil
		} else {
			// t_start t_ident err !
			//         ...^^^^
			lex.Emit(t_ident)
			// t_start t_ident err !
			//                 ^^^
			lex.EmitError(err)
			// t_start t_ident err !
			//                     ^
			return nil
		}
	}
}

// lexExpMode identifies the optional expansion mode
func lexExpMode(lex lexer.Lexer) lexer.StateFn {
	input, err := lex.AtLeast(2)

	if _, ok := r_mode(input); ok {
		// t_start t_ident t_mode t_text t_end
		//                 ^^^^^^
		lex.Step(2)
		lex.Emit(t_mode)
		// t_start t_ident t_mode t_text t_end
		//                        ^^....
		return lexText
	} else if len(input) > 1 {
		// t_start t_ident ':' WTF !
		//                 ^^^^^^^
		lex.EmitSyntaxError("invalid expansion mode")
		// t_start t_ident ':' WTF !
		//                         ^
		return nil
	} else {
		// t_start t_ident ':' err !
		//                     ^^^
		lex.EmitError(err)
		// t_start t_ident ':' err !
		//                         ^
		return nil
	}
}

// lexVar represents a $foo expansion
func lexVar(lex lexer.Lexer) lexer.StateFn {
	input, err := lex.AtLeast(1)

	if s, ok := r_ident(input); ok {
		// '$' t_ident
		//     ^^...
		lex.Step(len(s))
		return lexVarMore
	} else if len(input) > 0 {
		// '$' WTH !
		//     ^^^
		lex.EmitSyntaxError("invalid expansion")
		// '$' WTH !
		//         ^
		return nil
	} else {
		// '$' err !
		//     ^^^
		lex.EmitError(err)
		// '$' err !
		//         ^
		return nil
	}
}

// lexVarMore continues a $foo expansion
func lexVarMore(lex lexer.Lexer) lexer.StateFn {
	for {
		input, err := lex.AtLeast(1)

		if s, ok := r_ident_more(input); ok {
			// '$' t_ident t_text
			//     ...^^..
			lex.Step(len(s))
			continue
		} else {
			// '$' t_ident ...
			// ^^^^^^^^^^^
			lex.Emit(t_expansion)

			if len(input) > 0 {
				// '$' t_ident t_text
				//             ^^....
				return lexText
			} else {
				// '$' t_ident err !
				//             ^^^
				lex.EmitError(err)
				// '$' t_ident err !
				//                 ^
				return nil
			}
		}
	}
}
