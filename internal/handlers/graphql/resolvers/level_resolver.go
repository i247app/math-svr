package resolvers

import (
	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type LevelResolver struct {
	levelService di.ILevelService
}

func NewLevelResolver(levelService di.ILevelService) *LevelResolver {
	return &LevelResolver{
		levelService: levelService,
	}
}

// GetLevel resolves a single level by ID
func (r *LevelResolver) GetLevel(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return nil, nil
	}

	statusCode, level, err := r.levelService.GetLevelByID(params.Context, id)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return level, nil
}

// GetLevelByLabel resolves a single level by label
func (r *LevelResolver) GetLevelByLabel(params graphql.ResolveParams) (interface{}, error) {
	label, ok := params.Args["label"].(string)
	if !ok {
		return nil, nil
	}

	statusCode, level, err := r.levelService.GetLevelByLabel(params.Context, label)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return level, nil
}

// ListLevels resolves a list of levels
func (r *LevelResolver) ListLevels(params graphql.ResolveParams) (interface{}, error) {
	req := &dto.ListLevelRequest{
		Page:    1,
		Limit:   100,
		TakeAll: true,
	}

	statusCode, levels, _, err := r.levelService.ListLevels(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return levels, nil
}

// CreateLevel resolves level creation
func (r *LevelResolver) CreateLevel(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.CreateLevelRequest{
		Label:        input["label"].(string),
		Description:  input["description"].(string),
		DisplayOrder: int8(input["display_order"].(int)),
	}

	if iconURL, ok := input["icon_url"].(string); ok {
		req.IconURL = &iconURL
	}

	statusCode, level, err := r.levelService.CreateLevel(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return level, nil
}

// UpdateLevel resolves level updates
func (r *LevelResolver) UpdateLevel(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	id, ok := input["id"].(string)
	if !ok {
		return nil, nil
	}

	req := &dto.UpdateLevelRequest{
		ID: id,
	}

	if label, ok := input["label"].(string); ok {
		req.Label = &label
	}
	if description, ok := input["description"].(string); ok {
		req.Description = &description
	}
	if iconURL, ok := input["icon_url"].(string); ok {
		req.IconURL = &iconURL
	}
	if displayOrder, ok := input["display_order"].(int); ok {
		order := int8(displayOrder)
		req.DisplayOrder = &order
	}

	statusCode, level, err := r.levelService.UpdateLevel(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return level, nil
}

// DeleteLevel resolves level soft deletion
func (r *LevelResolver) DeleteLevel(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return false, nil
	}

	statusCode, err := r.levelService.DeleteLevel(params.Context, id)
	if err != nil || statusCode != status.SUCCESS {
		return false, err
	}

	return true, nil
}

// ForceDeleteLevel resolves level hard deletion
func (r *LevelResolver) ForceDeleteLevel(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return false, nil
	}

	statusCode, err := r.levelService.ForceDeleteLevel(params.Context, id)
	if err != nil || statusCode != status.SUCCESS {
		return false, err
	}

	return true, nil
}
