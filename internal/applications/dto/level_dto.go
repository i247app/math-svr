package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/level"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type LevelResponse struct {
	ID           string    `json:"id"`
	Label        string    `json:"label"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	DisplayOrder int8      `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type GetLevelResponse struct {
	Level *LevelResponse `json:"result"`
}

type ListLevelRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type ListLevelResponse struct {
	Items      []*LevelResponse       `json:"result"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateLevelRequest struct {
	Label        string `json:"label"`
	Description  string `json:"description"`
	DisplayOrder int8   `json:"display_order"`
}

type CreateLevelResponse struct {
	Level *LevelResponse `json:"result"`
}

type UpdateLevelRequest struct {
	ID           string        `json:"id"`
	Label        *string       `json:"label,omitempty"`
	Description  *string       `json:"description,omitempty"`
	Status       *enum.EStatus `json:"status,omitempty"`
	DisplayOrder *int8         `json:"display_order,omitempty"`
}

type UpdateLevelResponse struct {
	Level *LevelResponse `json:"result"`
}

type DeleteLevelRequest struct {
	ID string `json:"id"`
}

func BuildLevelDomainForCreate(dto *CreateLevelRequest) *domain.Level {
	levelDomain := domain.NewLevelDomain()
	levelDomain.GenerateID()
	levelDomain.SetLabel(dto.Label)
	levelDomain.SetDescription(dto.Description)
	levelDomain.SetStatus(string(enum.StatusActive))
	levelDomain.SetDisplayOrder(dto.DisplayOrder)

	return levelDomain
}

func BuildLevelDomainForUpdate(dto *UpdateLevelRequest) *domain.Level {
	levelDomain := domain.NewLevelDomain()
	levelDomain.SetID(dto.ID)

	if dto.Label != nil {
		levelDomain.SetLabel(*dto.Label)
	}

	if dto.Description != nil {
		levelDomain.SetDescription(*dto.Description)
	}

	if dto.Status != nil {
		levelDomain.SetStatus(string(*dto.Status))
	}

	if dto.DisplayOrder != nil {
		levelDomain.SetDisplayOrder(*dto.DisplayOrder)
	}

	return levelDomain
}

func LevelResponseFromDomain(l *domain.Level) LevelResponse {
	return LevelResponse{
		ID:           l.ID(),
		Label:        l.Label(),
		Description:  l.Description(),
		Status:       l.Status(),
		DisplayOrder: l.DisplayOrder(),
		CreatedAt:    l.CreatedAt(),
		ModifiedAt:   l.ModifiedAt(),
	}
}

func LevelResponseListFromDomain(levels []*domain.Level) []LevelResponse {
	responses := make([]LevelResponse, len(levels))
	for i, l := range levels {
		responses[i] = LevelResponseFromDomain(l)
	}
	return responses
}
