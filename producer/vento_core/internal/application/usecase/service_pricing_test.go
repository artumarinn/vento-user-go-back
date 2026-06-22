package usecase

import (
	"errors"
	"testing"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

func ptrFloat(f float64) *float64 { return &f }
func ptrStr(s string) *string     { return &s }

func testServiceWithSchema() *entity.Service {
	s := entity.Service(catalog.Service{
		ID:      "svc-1",
		Name:    "Impresión 3D",
		Formula: "material * peso",
		VariablesSchema: []catalog.VariableDefinition{
			{
				Name:     "material",
				Type:     catalog.VariableTypeSelect,
				Required: true,
				Options: []catalog.VariableOption{
					{Label: "PLA", Value: "PLA", UnitCost: 2},
					{Label: "ABS", Value: "ABS", UnitCost: 3},
				},
			},
			{
				Name:     "peso",
				Type:     catalog.VariableTypeNumber,
				Required: true,
				UnitCost: ptrFloat(1),
			},
			{
				Name: "color",
				Type: catalog.VariableTypeText,
			},
		},
	})
	return &s
}

func TestBuildVarMap_ValidVariables(t *testing.T) {
	svc := testServiceWithSchema()
	items := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
		{Name: "color", Type: "text", TextValue: ptrStr("rojo")},
	}

	vars, err := buildVarMap(svc, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["material"] != 2 {
		t.Errorf("expected material=2, got %v", vars["material"])
	}
	if vars["peso"] != 150 {
		t.Errorf("expected peso=150, got %v", vars["peso"])
	}
	if _, ok := vars["color"]; ok {
		t.Errorf("text variable must not be addressable in formula map")
	}
}

func TestBuildVarMap_MissingRequiredVariable(t *testing.T) {
	svc := testServiceWithSchema()
	items := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
	}

	_, err := buildVarMap(svc, items)
	if err == nil {
		t.Fatal("expected error for missing required variable 'peso'")
	}
	if !errors.Is(err, ErrMissingRequiredVariable) {
		t.Errorf("expected ErrMissingRequiredVariable, got %v", err)
	}
}

func TestBuildVarMap_InvalidSelectOption(t *testing.T) {
	svc := testServiceWithSchema()
	items := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PVC")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
	}

	_, err := buildVarMap(svc, items)
	if err == nil {
		t.Fatal("expected error for invalid select option")
	}
	if !errors.Is(err, ErrInvalidVariableOption) {
		t.Errorf("expected ErrInvalidVariableOption, got %v", err)
	}
}

func TestBuildVarMap_UnknownVariableName(t *testing.T) {
	svc := testServiceWithSchema()
	items := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
		{Name: "tiempo", Type: "number", NumberValue: ptrFloat(5)},
	}

	_, err := buildVarMap(svc, items)
	if err == nil {
		t.Fatal("expected error for unknown variable name")
	}
	if !errors.Is(err, ErrUnknownVariableName) {
		t.Errorf("expected ErrUnknownVariableName, got %v", err)
	}
}

func TestBuildVarMap_PricedValuePersisted(t *testing.T) {
	svc := testServiceWithSchema()
	items := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("ABS")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(10)},
	}

	_, err := buildVarMap(svc, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items[0].PricedValue == nil || *items[0].PricedValue != 3 {
		t.Errorf("expected PricedValue=3 for ABS option, got %v", items[0].PricedValue)
	}
	if items[1].PricedValue == nil || *items[1].PricedValue != 10 {
		t.Errorf("expected PricedValue=10 for peso, got %v", items[1].PricedValue)
	}
}
