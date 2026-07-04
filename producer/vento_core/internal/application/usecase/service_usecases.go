package usecase

import (
	"context"
	"fmt"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/pricing"
)

type ServiceUsecases struct {
	repo    port.ServiceRepository
	tagRepo port.TagRepository
}

func NewServiceUsecases(repo port.ServiceRepository, tagRepo port.TagRepository) *ServiceUsecases {
	return &ServiceUsecases{repo: repo, tagRepo: tagRepo}
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
