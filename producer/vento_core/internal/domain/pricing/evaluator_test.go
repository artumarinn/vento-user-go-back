package pricing

import (
	"errors"
	"math"
	"testing"
)

func TestEvaluate_BasicArithmetic(t *testing.T) {
	cases := []struct {
		name    string
		formula string
		vars    map[string]VariableValue
		want    float64
	}{
		{"addition", "2 + 3", nil, 5},
		{"subtraction", "10 - 4", nil, 6},
		{"multiplication", "material * peso", map[string]VariableValue{"material": 2, "peso": 150}, 300},
		{"division", "10 / 2", nil, 5},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Evaluate(c.formula, c.vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestEvaluate_OperatorPrecedence(t *testing.T) {
	got, err := Evaluate("2 + 3 * 4", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 14 {
		t.Errorf("got %v, want 14", got)
	}
}

func TestEvaluate_Parentheses(t *testing.T) {
	got, err := Evaluate("(2 + 3) * 4", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 20 {
		t.Errorf("got %v, want 20", got)
	}
}

func TestEvaluate_NestedParentheses(t *testing.T) {
	got, err := Evaluate("((1 + 2) * (3 + 4))", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 21 {
		t.Errorf("got %v, want 21", got)
	}
}

func TestEvaluate_UnaryMinus(t *testing.T) {
	got, err := Evaluate("-5 + 10", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 5 {
		t.Errorf("got %v, want 5", got)
	}
}

func TestEvaluate_MissingVariable(t *testing.T) {
	_, err := Evaluate("peso * 2", map[string]VariableValue{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnknownIdent) {
		t.Errorf("expected ErrUnknownIdent, got %v", err)
	}
}

func TestEvaluate_UnknownIdentifier(t *testing.T) {
	_, err := Evaluate("tiempo * 2", map[string]VariableValue{"peso": 10})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnknownIdent) {
		t.Errorf("expected ErrUnknownIdent, got %v", err)
	}
}

func TestEvaluate_DivisionByZero(t *testing.T) {
	_, err := Evaluate("costoTotal / cantidad", map[string]VariableValue{"costoTotal": 100, "cantidad": 0})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDivByZero) {
		t.Errorf("expected ErrDivByZero, got %v", err)
	}
}

func TestEvaluate_EmptyFormula(t *testing.T) {
	_, err := Evaluate("", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrEmptyFormula) {
		t.Errorf("expected ErrEmptyFormula, got %v", err)
	}

	_, err = Evaluate("   ", nil)
	if !errors.Is(err, ErrEmptyFormula) {
		t.Errorf("expected ErrEmptyFormula for whitespace-only formula, got %v", err)
	}
}

func TestEvaluate_MalformedSyntax(t *testing.T) {
	cases := []string{
		"2 +",
		"* 3",
		"(2 + 3",
		"2 + 3)",
		"2 ** 3",
		"2 3",
		"2 + + 3", // unary plus not supported, only unary minus via factor
	}

	for _, formula := range cases {
		t.Run(formula, func(t *testing.T) {
			_, err := Evaluate(formula, nil)
			if err == nil {
				t.Fatalf("expected syntax error for %q, got nil", formula)
			}
			if !errors.Is(err, ErrSyntax) {
				t.Errorf("expected ErrSyntax for %q, got %v", formula, err)
			}
		})
	}
}

func TestEvaluate_NaNInfGuard(t *testing.T) {
	// 0/0 style scenarios are already caught by ErrDivByZero before computing,
	// but guard against any other path producing Inf/NaN.
	got, err := Evaluate("1 / (2 - 2)", nil)
	if err == nil {
		t.Fatalf("expected error, got result %v", got)
	}
	if !errors.Is(err, ErrDivByZero) {
		t.Errorf("expected ErrDivByZero, got %v", err)
	}
}

func TestEvaluate_ResultIsFinite(t *testing.T) {
	got, err := Evaluate("2 + 2", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Errorf("result must be finite, got %v", got)
	}
}

func TestValidate_AllIdentifiersAllowed(t *testing.T) {
	err := Validate("material * peso", []string{"material", "peso"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_RejectsUnknownIdentifier(t *testing.T) {
	err := Validate("material * peso + tiempo", []string{"material", "peso"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnknownIdent) {
		t.Errorf("expected ErrUnknownIdent, got %v", err)
	}
}

func TestValidate_RejectsSyntaxError(t *testing.T) {
	err := Validate("material *", []string{"material"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrSyntax) {
		t.Errorf("expected ErrSyntax, got %v", err)
	}
}

func TestValidate_RejectsEmptyFormula(t *testing.T) {
	err := Validate("", []string{"material"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrEmptyFormula) {
		t.Errorf("expected ErrEmptyFormula, got %v", err)
	}
}
