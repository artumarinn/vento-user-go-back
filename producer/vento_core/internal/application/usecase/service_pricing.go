package usecase

import (
	"errors"
	"fmt"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/pricing"
)

var (
	// ErrMissingRequiredVariable indicates a required Service variable has no
	// value supplied on the order item.
	ErrMissingRequiredVariable = errors.New("pricing: missing required variable")
	// ErrInvalidVariableOption indicates a select variable's value is not one
	// of the Service's defined options.
	ErrInvalidVariableOption = errors.New("pricing: invalid variable option")
	// ErrUnknownVariableName indicates an order item references a variable
	// name not declared on the Service's schema.
	ErrUnknownVariableName = errors.New("pricing: unknown variable name")
)

// buildVarMap validates an OrderItem's typed variable instances against the
// Service's VariablesSchema and resolves each into a pricing.VariableValue
// the formula evaluator can consume. It also writes the resolved PricedValue
// back onto each variable instance for audit (mutates items in place).
func buildVarMap(service *entity.Service, items []entity.OrderItemVariable) (map[string]pricing.VariableValue, error) {
	defsByName := make(map[string]catalog.VariableDefinition, len(service.VariablesSchema))
	for _, def := range service.VariablesSchema {
		defsByName[def.Name] = def
	}

	providedByName := make(map[string]int, len(items))
	for i, item := range items {
		if _, ok := defsByName[item.Name]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownVariableName, item.Name)
		}
		providedByName[item.Name] = i
	}

	vars := make(map[string]pricing.VariableValue)

	for _, def := range service.VariablesSchema {
		idx, provided := providedByName[def.Name]
		if !provided {
			if def.Required {
				return nil, fmt.Errorf("%w: %s", ErrMissingRequiredVariable, def.Name)
			}
			continue
		}

		item := &items[idx]

		switch def.Type {
		case catalog.VariableTypeSelect:
			if item.OptionValue == nil {
				if def.Required {
					return nil, fmt.Errorf("%w: %s", ErrMissingRequiredVariable, def.Name)
				}
				continue
			}
			var matched *catalog.VariableOption
			for i := range def.Options {
				if def.Options[i].Value == *item.OptionValue {
					matched = &def.Options[i]
					break
				}
			}
			if matched == nil {
				return nil, fmt.Errorf("%w: %s=%s", ErrInvalidVariableOption, def.Name, *item.OptionValue)
			}
			priced := matched.UnitCost
			item.PricedValue = &priced
			vars[def.Name] = pricing.VariableValue(priced)

		case catalog.VariableTypeNumber:
			if item.NumberValue == nil {
				if def.Required {
					return nil, fmt.Errorf("%w: %s", ErrMissingRequiredVariable, def.Name)
				}
				continue
			}
			unitCost := 1.0
			if def.UnitCost != nil {
				unitCost = *def.UnitCost
			}
			priced := *item.NumberValue * unitCost
			item.PricedValue = &priced
			vars[def.Name] = pricing.VariableValue(priced)

		case catalog.VariableTypeText:
			// Text variables carry no price contribution and are not
			// addressable in the formula identifier map.
			continue
		}
	}

	return vars, nil
}
