package pricing

import (
	"fmt"
	"strings"
	"unicode"
)

// tokenKind identifies the lexical category of a token.
type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenNumber
	tokenIdent
	tokenPlus
	tokenMinus
	tokenStar
	tokenSlash
	tokenLParen
	tokenRParen
)

type token struct {
	kind  tokenKind
	text  string
	value float64 // populated for tokenNumber
}

// lexer tokenizes a closed arithmetic formula: numeric literals, identifiers
// `[a-zA-Z_][a-zA-Z0-9_]*`, and the operators `+ - * / ( )`.
type lexer struct {
	input []rune
	pos   int
}

func newLexer(formula string) *lexer {
	return &lexer{input: []rune(formula)}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) advance() rune {
	r := l.peek()
	l.pos++
	return r
}

func (l *lexer) skipWhitespace() {
	for unicode.IsSpace(l.peek()) {
		l.pos++
	}
}

// next returns the next token in the stream, or a syntax error if the
// character sequence cannot be tokenized.
func (l *lexer) next() (token, error) {
	l.skipWhitespace()

	r := l.peek()
	switch {
	case r == 0:
		return token{kind: tokenEOF}, nil
	case r == '+':
		l.advance()
		return token{kind: tokenPlus, text: "+"}, nil
	case r == '-':
		l.advance()
		return token{kind: tokenMinus, text: "-"}, nil
	case r == '*':
		l.advance()
		return token{kind: tokenStar, text: "*"}, nil
	case r == '/':
		l.advance()
		return token{kind: tokenSlash, text: "/"}, nil
	case r == '(':
		l.advance()
		return token{kind: tokenLParen, text: "("}, nil
	case r == ')':
		l.advance()
		return token{kind: tokenRParen, text: ")"}, nil
	case unicode.IsDigit(r) || r == '.':
		return l.lexNumber()
	case unicode.IsLetter(r) || r == '_':
		return l.lexIdent()
	default:
		return token{}, fmt.Errorf("%w: unexpected character %q", ErrSyntax, r)
	}
}

func (l *lexer) lexNumber() (token, error) {
	var sb strings.Builder
	dotSeen := false
	for {
		r := l.peek()
		if unicode.IsDigit(r) {
			sb.WriteRune(r)
			l.advance()
			continue
		}
		if r == '.' {
			if dotSeen {
				return token{}, fmt.Errorf("%w: malformed number %q", ErrSyntax, sb.String()+".")
			}
			dotSeen = true
			sb.WriteRune(r)
			l.advance()
			continue
		}
		break
	}

	text := sb.String()
	if text == "" || text == "." {
		return token{}, fmt.Errorf("%w: malformed number", ErrSyntax)
	}

	var value float64
	if _, err := fmt.Sscanf(text, "%g", &value); err != nil {
		return token{}, fmt.Errorf("%w: malformed number %q", ErrSyntax, text)
	}
	return token{kind: tokenNumber, text: text, value: value}, nil
}

func (l *lexer) lexIdent() (token, error) {
	var sb strings.Builder
	for {
		r := l.peek()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			sb.WriteRune(r)
			l.advance()
			continue
		}
		break
	}
	return token{kind: tokenIdent, text: sb.String()}, nil
}
