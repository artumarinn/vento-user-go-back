package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/pricing"
)

type ServiceUsecases struct {
	repo       port.ServiceRepository
	tagRepo    port.TagRepository
	insumoRepo port.InsumoRepository
}

func NewServiceUsecases(repo port.ServiceRepository, tagRepo port.TagRepository, insumoRepo port.InsumoRepository) *ServiceUsecases {
	return &ServiceUsecases{repo: repo, tagRepo: tagRepo, insumoRepo: insumoRepo}
}

func (uc *ServiceUsecases) ListServices(ctx context.Context, userID string) ([]dto.ServiceResponse, error) {
	services, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ServiceResponse, len(services))
	for i, service := range services {
		tags, err := uc.tagRepo.ListByServiceID(ctx, service.ID)
		if err != nil {
			return nil, err
		}
		insumos, err := uc.repo.ListServiceInsumos(ctx, service.ID)
		if err != nil {
			return nil, err
		}
		res[i] = mapServiceEntityToDTO(service, tags, insumos)
	}
	return res, nil
}

// allowedVariableIdents returns the formula-addressable identifier names from
// a variable schema (text variables are excluded — they carry no price
// contribution and cannot be referenced in a formula).
func allowedVariableIdents(schema []catalog.VariableDefinition) ([]string, error) {
	seen := make(map[string]bool, len(schema))
	idents := make([]string, 0, len(schema))
	for _, def := range schema {
		if seen[def.Name] {
			return nil, fmt.Errorf("duplicate variable name: %s", def.Name)
		}
		seen[def.Name] = true
		if def.Type != catalog.VariableTypeText {
			idents = append(idents, def.Name)
		}
	}
	return idents, nil
}

func (uc *ServiceUsecases) CreateService(ctx context.Context, userID string, req dto.CreateServiceRequest) (dto.ServiceResponse, error) {
	idents, err := allowedVariableIdents(req.VariablesSchema)
	if err != nil {
		return dto.ServiceResponse{}, err
	}
	if err := pricing.Validate(req.Formula, idents); err != nil {
		return dto.ServiceResponse{}, err
	}

	s := entity.NewService(userID, req.Name, req.Formula, req.MinimumLeadTime)
	s.VariablesSchema = req.VariablesSchema

	if err := uc.repo.Save(ctx, s); err != nil {
		return dto.ServiceResponse{}, err
	}

	if err := uc.tagRepo.SetServiceTags(ctx, s.ID, req.TagIDs); err != nil {
		return dto.ServiceResponse{}, err
	}

	if err := uc.repo.SetServiceInsumos(ctx, s.ID, mapServiceInsumoInputsToEntities(s.ID, req.Insumos)); err != nil {
		return dto.ServiceResponse{}, err
	}

	tags, err := uc.tagRepo.ListByServiceID(ctx, s.ID)
	if err != nil {
		return dto.ServiceResponse{}, err
	}

	insumos, err := uc.repo.ListServiceInsumos(ctx, s.ID)
	if err != nil {
		return dto.ServiceResponse{}, err
	}

	return mapServiceEntityToDTO(s, tags, insumos), nil
}

func (uc *ServiceUsecases) BatchCreateServices(ctx context.Context, userID string, req dto.BatchCreateServiceRequest) ([]dto.ServiceResponse, error) {
	services := make([]*entity.Service, len(req.Services))
	res := make([]dto.ServiceResponse, len(req.Services))

	for i, sReq := range req.Services {
		service := entity.NewService(userID, sReq.Name, sReq.Formula, sReq.MinimumLeadTime)
		services[i] = service
		res[i] = mapServiceEntityToDTO(service, nil, nil)
	}

	if err := uc.repo.SaveBatch(ctx, services); err != nil {
		return nil, err
	}

	return res, nil
}

func (uc *ServiceUsecases) UpdateService(ctx context.Context, userID string, serviceID string, req dto.UpdateServiceRequest) (dto.ServiceResponse, error) {
	s, err := uc.repo.GetByID(ctx, serviceID, userID)
	if err != nil {
		return dto.ServiceResponse{}, err
	}
	if s == nil {
		return dto.ServiceResponse{}, nil
	}

	if req.Name != "" {
		s.Name = req.Name
	}
	if req.Formula != "" {
		s.Formula = req.Formula
	}
	if req.MinimumLeadTime != 0 {
		s.MinimumLeadTime = req.MinimumLeadTime
	}
	if req.VariablesSchema != nil {
		s.VariablesSchema = req.VariablesSchema
	}

	idents, err := allowedVariableIdents(s.VariablesSchema)
	if err != nil {
		return dto.ServiceResponse{}, err
	}
	if err := pricing.Validate(s.Formula, idents); err != nil {
		return dto.ServiceResponse{}, err
	}

	if err := uc.repo.Update(ctx, s); err != nil {
		return dto.ServiceResponse{}, err
	}

	if req.TagIDs != nil {
		if err := uc.tagRepo.SetServiceTags(ctx, s.ID, req.TagIDs); err != nil {
			return dto.ServiceResponse{}, err
		}
	}

	if req.Insumos != nil {
		if err := uc.repo.SetServiceInsumos(ctx, s.ID, mapServiceInsumoInputsToEntities(s.ID, req.Insumos)); err != nil {
			return dto.ServiceResponse{}, err
		}
	}

	tags, err := uc.tagRepo.ListByServiceID(ctx, s.ID)
	if err != nil {
		return dto.ServiceResponse{}, err
	}

	insumos, err := uc.repo.ListServiceInsumos(ctx, s.ID)
	if err != nil {
		return dto.ServiceResponse{}, err
	}

	return mapServiceEntityToDTO(s, tags, insumos), nil
}

func (uc *ServiceUsecases) DeleteService(ctx context.Context, userID string, serviceID string) error {
	return uc.repo.Delete(ctx, serviceID, userID)
}

func (uc *ServiceUsecases) GetService(ctx context.Context, userID string, serviceID string) (*dto.ServiceResponse, error) {
	s, err := uc.repo.GetByID(ctx, serviceID, userID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	tags, err := uc.tagRepo.ListByServiceID(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	insumos, err := uc.repo.ListServiceInsumos(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	res := mapServiceEntityToDTO(s, tags, insumos)
	return &res, nil
}

// mapServiceInsumoInputsToEntities converts the request-level recipe input
// into persistence-ready entities, attaching the parent ServiceID — mirrors
// how tag IDs are passed through tagRepo.SetServiceTags but insumos need the
// quantity_per_unit carried along, so they cannot reuse the plain []string shape.
func mapServiceInsumoInputsToEntities(serviceID string, inputs []dto.ServiceInsumoInput) []entity.ServiceInsumo {
	insumos := make([]entity.ServiceInsumo, len(inputs))
	for i, in := range inputs {
		insumos[i] = entity.ServiceInsumo{
			ServiceID:       serviceID,
			InsumoID:        in.InsumoID,
			QuantityPerUnit: in.QuantityPerUnit,
		}
	}
	return insumos
}

func mapServiceEntityToDTO(s *entity.Service, tags []*entity.Tag, insumos []entity.ServiceInsumoDetail) dto.ServiceResponse {
	tagResponses := make([]dto.TagResponse, len(tags))
	for idx, t := range tags {
		tagResponses[idx] = mapTagEntityToDTO(t)
	}

	insumoResponses := make([]dto.ServiceInsumoResponse, len(insumos))
	for idx, in := range insumos {
		insumoResponses[idx] = dto.ServiceInsumoResponse{
			InsumoID:        in.InsumoID,
			InsumoName:      in.InsumoName,
			QuantityPerUnit: in.QuantityPerUnit,
			Unit:            in.Unit,
		}
	}

	return dto.ServiceResponse{
		ID:              s.ID,
		UserID:          s.UserID,
		Name:            s.Name,
		Formula:         s.Formula,
		MinimumLeadTime: s.MinimumLeadTime,
		VariablesSchema: s.VariablesSchema,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		Tags:            tagResponses,
		Insumos:         insumoResponses,
	}
}

// PreviewPrice computes the price a Service formula would yield for the
// given variable values, without persisting an order. It reuses the exact
// same buildVarMap+pricing.Evaluate path used at order creation time, so the
// previewed price is guaranteed to match the price that would be persisted.
func (uc *ServiceUsecases) PreviewPrice(ctx context.Context, userID, serviceID string, vars []entity.OrderItemVariable) (float64, error) {
	s, err := uc.repo.GetByID(ctx, serviceID, userID)
	if err != nil {
		return 0, err
	}
	if s == nil {
		return 0, fmt.Errorf("service not found: %s", serviceID)
	}

	varMap, err := buildVarMap(s, vars)
	if err != nil {
		return 0, err
	}

	return pricing.Evaluate(s.Formula, varMap)
}

// PreviewPriceDraft computes the price an in-progress (not yet persisted)
// formula+schema would yield, reusing the same buildVarMap+pricing.Evaluate
// path as PreviewPrice, so a service author sees the exact price a real
// order would compute before saving the service.
func (uc *ServiceUsecases) PreviewPriceDraft(formula string, schema []catalog.VariableDefinition, vars []entity.OrderItemVariable) (float64, error) {
	idents, err := allowedVariableIdents(schema)
	if err != nil {
		return 0, err
	}
	if err := pricing.Validate(formula, idents); err != nil {
		return 0, err
	}

	draft := &entity.Service{Formula: formula, VariablesSchema: schema}
	varMap, err := buildVarMap(draft, vars)
	if err != nil {
		return 0, err
	}
	return pricing.Evaluate(draft.Formula, varMap)
}

// normalizeForServiceSearch lowercases and strips accents so a customer's
// natural-language Spanish query ("impresión") matches a service name stored
// without accents ("Impresion") — mirrors business_tool_handler.go's
// normalizeForSearch, duplicated here rather than imported since usecase
// must not depend on the http handler layer.
func normalizeForServiceSearch(s string) string {
	s = strings.ToLower(s)
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	)
	return replacer.Replace(s)
}

// SearchServices returns customer-safe service listings (name + minimum lead
// time only — never the pricing formula or insumo recipe) matching a
// free-text query against the service name. Empty query returns all services.
func (uc *ServiceUsecases) SearchServices(ctx context.Context, userID, query string) ([]dto.ServiceSearchResult, error) {
	services, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	q := normalizeForServiceSearch(strings.TrimSpace(query))
	results := make([]dto.ServiceSearchResult, 0, len(services))
	for _, s := range services {
		if q != "" && !strings.Contains(normalizeForServiceSearch(s.Name), q) {
			continue
		}
		results = append(results, dto.ServiceSearchResult{
			ID:              s.ID,
			Name:            s.Name,
			MinimumLeadTime: s.MinimumLeadTime,
		})
	}
	return results, nil
}

// CheckFeasibility reports whether a service is offered by this tenant
// (offered) and, given its insumo recipe, whether current stock could
// produce the requested quantity (feasible). A service with no recipe
// declared is always feasible — matches validateInsumoStock's own
// nil-stock convention in order_usecases.go: an insumo the owner never
// tracked stock for must never block anything. Never returns the formula
// or the recipe itself — only the two booleans.
func (uc *ServiceUsecases) CheckFeasibility(ctx context.Context, userID, serviceID string, quantity float64) (offered bool, feasible bool, err error) {
	service, err := uc.repo.GetByID(ctx, serviceID, userID)
	if err != nil {
		return false, false, err
	}
	if service == nil {
		return false, false, nil
	}

	recipe, err := uc.repo.ListServiceInsumos(ctx, serviceID)
	if err != nil {
		return true, false, err
	}

	for _, line := range recipe {
		insumo, err := uc.insumoRepo.GetByID(ctx, line.InsumoID, userID)
		if err != nil {
			return true, false, err
		}
		if insumo == nil {
			continue
		}
		needed := line.QuantityPerUnit * quantity
		if insumo.Stock != nil && *insumo.Stock < needed {
			return true, false, nil
		}
	}
	return true, true, nil
}

// GetServiceVariables returns the customer-safe variable schema for a
// service — name/label/type/unit/options/required only, never unit_cost or
// the formula. Used by the AI agent to know what to ask a customer before
// quoting a custom order.
func (uc *ServiceUsecases) GetServiceVariables(ctx context.Context, userID, serviceID string) ([]dto.ServiceVariableResponse, error) {
	s, err := uc.repo.GetByID(ctx, serviceID, userID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, ErrServiceNotFoundForPricing
	}

	results := make([]dto.ServiceVariableResponse, len(s.VariablesSchema))
	for i, def := range s.VariablesSchema {
		options := make([]dto.ServiceVariableOption, len(def.Options))
		for j, opt := range def.Options {
			options[j] = dto.ServiceVariableOption{Label: opt.Label, Value: opt.Value}
		}
		results[i] = dto.ServiceVariableResponse{
			Name:     def.Name,
			Label:    def.Label,
			Type:     string(def.Type),
			Unit:     def.Unit,
			Options:  options,
			Required: def.Required,
		}
	}
	return results, nil
}

// GetServicePrice coerces a flat map of customer-provided variable values
// into the typed OrderItemVariable shape PreviewPrice needs, using the
// service's own VariablesSchema as the single source of truth for how to
// interpret each value — the AI service never needs to know these types
// itself. Returns ErrServiceNotFoundForPricing, ErrUnknownVariableName,
// ErrMissingRequiredVariable, ErrInvalidVariableOption, or
// ErrInvalidVariableValue for data problems; never returns a partial price.
func (uc *ServiceUsecases) GetServicePrice(ctx context.Context, userID, serviceID string, variables map[string]any) (float64, error) {
	s, err := uc.repo.GetByID(ctx, serviceID, userID)
	if err != nil {
		return 0, err
	}
	if s == nil {
		return 0, ErrServiceNotFoundForPricing
	}

	defsByName := make(map[string]catalog.VariableDefinition, len(s.VariablesSchema))
	for _, def := range s.VariablesSchema {
		defsByName[def.Name] = def
	}

	items := make([]entity.OrderItemVariable, 0, len(variables))
	for name, raw := range variables {
		def, ok := defsByName[name]
		if !ok {
			items = append(items, entity.OrderItemVariable{Name: name})
			continue
		}

		item := entity.OrderItemVariable{Name: name, Type: string(def.Type)}
		switch def.Type {
		case catalog.VariableTypeSelect:
			v, ok := raw.(string)
			if !ok {
				return 0, fmt.Errorf("%w: %s", ErrInvalidVariableValue, name)
			}
			item.OptionValue = &v
		case catalog.VariableTypeNumber:
			v, ok := raw.(float64)
			if !ok {
				return 0, fmt.Errorf("%w: %s", ErrInvalidVariableValue, name)
			}
			item.NumberValue = &v
		case catalog.VariableTypeText:
			v, ok := raw.(string)
			if !ok {
				return 0, fmt.Errorf("%w: %s", ErrInvalidVariableValue, name)
			}
			item.TextValue = &v
		}
		items = append(items, item)
	}

	return uc.PreviewPrice(ctx, userID, serviceID, items)
}
