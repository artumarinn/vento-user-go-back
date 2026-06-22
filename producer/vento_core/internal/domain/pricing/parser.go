package pricing

import "fmt"

// astNode is a node in the parsed formula's abstract syntax tree.
type astNode interface{}

type numberNode struct {
	value float64
}

type identNode struct {
	name string
}

type unaryNode struct {
	op   tokenKind // tokenMinus
	expr astNode
}

type binaryNode struct {
	op    tokenKind // tokenPlus, tokenMinus, tokenStar, tokenSlash
	left  astNode
	right astNode
}

// parser implements a recursive-descent parser over the grammar:
//
//	expr   := term (("+"|"-") term)*
//	term   := factor (("*"|"/") factor)*
//	factor := number | identifier | "(" expr ")" | "-" factor
type parser struct {
	lex  *lexer
	cur  token
	curE error
}

func newParser(formula string) (*parser, error) {
	p := &parser{lex: newLexer(formula)}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *parser) advance() error {
	t, err := p.lex.next()
	if err != nil {
		return err
	}
	p.cur = t
	return nil
}

// parse parses the full formula and ensures all input was consumed.
func (p *parser) parse() (astNode, error) {
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.cur.kind != tokenEOF {
		return nil, fmt.Errorf("%w: unexpected token %q", ErrSyntax, p.cur.text)
	}
	return node, nil
}

func (p *parser) parseExpr() (astNode, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.cur.kind == tokenPlus || p.cur.kind == tokenMinus {
		op := p.cur.kind
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = binaryNode{op: op, left: left, right: right}
	}

	return left, nil
}

func (p *parser) parseTerm() (astNode, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.cur.kind == tokenStar || p.cur.kind == tokenSlash {
		op := p.cur.kind
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = binaryNode{op: op, left: left, right: right}
	}

	return left, nil
}

func (p *parser) parseFactor() (astNode, error) {
	switch p.cur.kind {
	case tokenNumber:
		v := p.cur.value
		if err := p.advance(); err != nil {
			return nil, err
		}
		return numberNode{value: v}, nil

	case tokenIdent:
		name := p.cur.text
		if err := p.advance(); err != nil {
			return nil, err
		}
		return identNode{name: name}, nil

	case tokenLParen:
		if err := p.advance(); err != nil {
			return nil, err
		}
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.cur.kind != tokenRParen {
			return nil, fmt.Errorf("%w: expected closing parenthesis", ErrSyntax)
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return inner, nil

	case tokenMinus:
		if err := p.advance(); err != nil {
			return nil, err
		}
		inner, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		return unaryNode{op: tokenMinus, expr: inner}, nil

	default:
		return nil, fmt.Errorf("%w: unexpected token %q", ErrSyntax, p.cur.text)
	}
}
