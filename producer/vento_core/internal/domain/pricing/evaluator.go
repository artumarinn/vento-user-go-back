// Package pricing implements a closed arithmetic grammar evaluator for
// Service pricing formulas. It is domain-pure (no external dependencies),
// deterministic, and never delegates calculation to the LLM — per the
// project's golden rule that prices are computed by code, never invented.
package pricing

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// VariableValue is the priced scalar an identifier resolves to during
// evaluation (e.g. a select option's unit cost, or number*unit_cost).
type VariableValue float64

var (
	// ErrSyntax indicates the formula could not be parsed.
	ErrSyntax = errors.New("pricing: invalid formula syntax")
	// ErrUnknownIdent indicates the formula references an identifier with no
	// resolvable value (missing variable or not in the allowed set).
	ErrUnknownIdent = errors.New("pricing: unknown identifier")
	// ErrDivByZero indicates a division operation's denominator evaluated to zero.
	ErrDivByZero = errors.New("pricing: division by zero")
	// ErrEmptyFormula indicates the formula string was empty or whitespace-only.
	ErrEmptyFormula = errors.New("pricing: empty formula")
)

// Evaluate parses and evaluates formula against the given variable values,
// returning the computed price. It never silently defaults a missing or
// invalid identifier to zero, and never returns Inf/NaN.
func Evaluate(formula string, vars map[string]VariableValue) (float64, error) {
	if strings.TrimSpace(formula) == "" {
		return 0, ErrEmptyFormula
	}

	p, err := newParser(formula)
	if err != nil {
		return 0, err
	}

	node, err := p.parse()
	if err != nil {
		return 0, err
	}

	result, err := evalNode(node, vars)
	if err != nil {
		return 0, err
	}

	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, fmt.Errorf("%w: result is not a finite number", ErrSyntax)
	}

	return result, nil
}

// Validate parses formula AND checks that every identifier it references
// exists in allowedIdents. Used at Service create/update time, before any
// variable values are known.
func Validate(formula string, allowedIdents []string) error {
	if strings.TrimSpace(formula) == "" {
		return ErrEmptyFormula
	}

	p, err := newParser(formula)
	if err != nil {
		return err
	}

	node, err := p.parse()
	if err != nil {
		return err
	}

	allowed := make(map[string]struct{}, len(allowedIdents))
	for _, name := range allowedIdents {
		allowed[name] = struct{}{}
	}

	return validateNode(node, allowed)
}

func validateNode(node astNode, allowed map[string]struct{}) error {
	switch n := node.(type) {
	case numberNode:
		return nil
	case identNode:
		if _, ok := allowed[n.name]; !ok {
			return fmt.Errorf("%w: %s", ErrUnknownIdent, n.name)
		}
		return nil
	case unaryNode:
		return validateNode(n.expr, allowed)
	case binaryNode:
		if err := validateNode(n.left, allowed); err != nil {
			return err
		}
		return validateNode(n.right, allowed)
	default:
		return fmt.Errorf("%w: unsupported node type", ErrSyntax)
	}
}

func evalNode(node astNode, vars map[string]VariableValue) (float64, error) {
	switch n := node.(type) {
	case numberNode:
		return n.value, nil

	case identNode:
		v, ok := vars[n.name]
		if !ok {
			return 0, fmt.Errorf("%w: %s", ErrUnknownIdent, n.name)
		}
		return float64(v), nil

	case unaryNode:
		inner, err := evalNode(n.expr, vars)
		if err != nil {
			return 0, err
		}
		return -inner, nil

	case binaryNode:
		left, err := evalNode(n.left, vars)
		if err != nil {
			return 0, err
		}
		right, err := evalNode(n.right, vars)
		if err != nil {
			return 0, err
		}

		switch n.op {
		case tokenPlus:
			return left + right, nil
		case tokenMinus:
			return left - right, nil
		case tokenStar:
			return left * right, nil
		case tokenSlash:
			if right == 0 {
				return 0, ErrDivByZero
			}
			return left / right, nil
		default:
			return 0, fmt.Errorf("%w: unsupported operator", ErrSyntax)
		}

	default:
		return 0, fmt.Errorf("%w: unsupported node type", ErrSyntax)
	}
}
