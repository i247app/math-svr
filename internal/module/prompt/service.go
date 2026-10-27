package prompt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	command "math-ai.com/math-ai/internal/application/command/prompt"
	dto "math-ai.com/math-ai/internal/application/dto/prompt"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/prompt"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	sctx "math-ai.com/math-ai/internal/shared/context"
)

// Service is the admin surface over ma_prompts: read every prompt, rewrite
// the text of one. Rows are never created or deleted here — the catalog of
// prompt kinds is code (domain/prompt/catalog.go), and the seed plants one
// row per scope.
type Service struct {
	repo      domain.IRepository
	updateCmd *command.UpdatePromptCommandHandler
}

func NewService(repo domain.IRepository, uow transaction.UnitOfWork) *Service {
	return &Service{repo: repo, updateCmd: command.NewUpdatePromptCommandHandler(uow)}
}

var ErrAdminGateMissing = errors.New("prompt: route did not pass the admin gate")

// requireAdminGate fails closed when a route was registered without the
// admin middleware: every prompt is sent to the model for every child.
func requireAdminGate(ctx context.Context) (sctx.AdminVia, error) {
	via := sctx.GetAdminVia(ctx)
	if via == "" {
		return "", errs.NewError(ctx, status.FORBIDDEN, nil, ErrAdminGateMissing)
	}
	return via, nil
}

// adminActor is the admin's uid for modify_id, or nil when the request came
// in with the X-Api-Admin-Key header and has no account behind it.
func adminActor(ctx context.Context) *int64 {
	if uid := sctx.UID(ctx); uid > 0 {
		return &uid
	}
	return nil
}

func (s *Service) ListPrompts(ctx context.Context, req *dto.ListPromptsReq) (*dto.ListPromptsRes, error) {
	if _, err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	filter := domain.ListFilter{Grade: req.Grade, Language: req.Language}
	if req.PromptCode != nil {
		code := domain.Code(strings.ToUpper(strings.TrimSpace(*req.PromptCode)))
		filter.Code = &code
	}
	prompts, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	res := &dto.ListPromptsRes{Prompts: dto.PromptsToResponse(prompts)}
	return res, nil
}

func (s *Service) GetPrompt(ctx context.Context, req *dto.GetPromptReq) (*dto.PromptRes, error) {
	if _, err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	p, err := s.findPrompt(ctx, req.PromptID)
	if err != nil {
		return nil, err
	}
	res := &dto.PromptRes{Prompt: dto.PromptToResponse(p)}
	return res, nil
}

// UpdatePrompt rewrites one row. The new text is used from the next call
// on; for exam generation the version it bumps moves the cache tag, so sets
// generated from the old text stop being served.
func (s *Service) UpdatePrompt(ctx context.Context, req *dto.UpdatePromptReq) (*dto.PromptRes, error) {
	via, err := requireAdminGate(ctx)
	if err != nil {
		return nil, err
	}
	if err := ValidateUpdatePrompt(ctx, req); err != nil {
		return nil, err
	}
	// The JSON keys a text must keep depend on the row's code, so the row is
	// read first. A code never changes, so reading it outside the write's
	// transaction cannot go stale.
	current, err := s.findPrompt(ctx, req.PromptID)
	if err != nil {
		return nil, err
	}
	if err := ValidatePromptKeys(ctx, current.PromptCode(), req.SystemPrompt); err != nil {
		return nil, err
	}
	saved, err := s.updateCmd.Handle(ctx, command.UpdatePromptCommand{
		PromptID:     req.PromptID,
		SystemPrompt: req.SystemPrompt,
		AdminUID:     adminActor(ctx),
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("prompt.updated admin_uid=%d via=%s prompt_id=%d code=%s prompt_version=%d runes=%d",
		sctx.UID(ctx), via, saved.PromptId(), saved.PromptCode(), saved.PromptVersion(), utf8.RuneCountInString(saved.SystemPrompt()))
	res := &dto.PromptRes{Prompt: dto.PromptToResponse(saved)}
	return res, nil
}

func (s *Service) findPrompt(ctx context.Context, promptID int64) (*domain.Prompt, error) {
	if promptID <= 0 {
		return nil, errs.NewError(ctx, status.PROMPT_MISSING_ID, nil, ErrPromptIDRequired)
	}
	p, err := s.repo.FindByPromptId(ctx, promptID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROMPT_NOT_FOUND, nil, fmt.Errorf("prompt: no row %d", promptID))
	}
	return p, nil
}
