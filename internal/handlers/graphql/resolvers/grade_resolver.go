package resolvers

import (
	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type GradeResolver struct {
	gradeService di.IGradeService
}

func NewGradeResolver(gradeService di.IGradeService) *GradeResolver {
	return &GradeResolver{
		gradeService: gradeService,
	}
}

// GetGrade resolves a single grade by ID
func (r *GradeResolver) GetGrade(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return nil, nil
	}

	statusCode, grade, err := r.gradeService.GetGradeByID(params.Context, id)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return grade, nil
}

// GetGradeByLabel resolves a single grade by label
func (r *GradeResolver) GetGradeByLabel(params graphql.ResolveParams) (interface{}, error) {
	label, ok := params.Args["label"].(string)
	if !ok {
		return nil, nil
	}

	statusCode, grade, err := r.gradeService.GetGradeByLabel(params.Context, label)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return grade, nil
}

// ListGrades resolves a list of grades
func (r *GradeResolver) ListGrades(params graphql.ResolveParams) (interface{}, error) {
	req := &dto.ListGradeRequest{
		Page:    1,
		Limit:   100,
		TakeAll: true,
	}

	statusCode, grades, _, err := r.gradeService.ListGrades(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return grades, nil
}

// CreateGrade resolves grade creation
func (r *GradeResolver) CreateGrade(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.CreateGradeRequest{
		Label:        input["label"].(string),
		Description:  input["description"].(string),
		DisplayOrder: int8(input["display_order"].(int)),
	}

	if iconURL, ok := input["image_key"].(string); ok {
		req.ImageKey = &iconURL
	}

	statusCode, grade, err := r.gradeService.CreateGrade(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return grade, nil
}

// UpdateGrade resolves grade updates
func (r *GradeResolver) UpdateGrade(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	id, ok := input["id"].(string)
	if !ok {
		return nil, nil
	}

	req := &dto.UpdateGradeRequest{
		ID: id,
	}

	if label, ok := input["label"].(string); ok {
		req.Label = &label
	}
	if description, ok := input["description"].(string); ok {
		req.Description = &description
	}
	if iconURL, ok := input["image_key"].(string); ok {
		req.ImageKey = &iconURL
	}
	if displayOrder, ok := input["display_order"].(int); ok {
		order := int8(displayOrder)
		req.DisplayOrder = &order
	}

	statusCode, grade, err := r.gradeService.UpdateGrade(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return grade, nil
}

// DeleteGrade resolves grade soft deletion
func (r *GradeResolver) DeleteGrade(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return false, nil
	}

	statusCode, err := r.gradeService.DeleteGrade(params.Context, id)
	if err != nil || statusCode != status.SUCCESS {
		return false, err
	}

	return true, nil
}

// ForceDeleteGrade resolves grade hard deletion
func (r *GradeResolver) ForceDeleteGrade(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return false, nil
	}

	statusCode, err := r.gradeService.ForceDeleteGrade(params.Context, id)
	if err != nil || statusCode != status.SUCCESS {
		return false, err
	}

	return true, nil
}
