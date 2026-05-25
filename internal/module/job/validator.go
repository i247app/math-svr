package job

import (
	"context"
	"errors"
	"strings"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func validateJobName(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return errs.NewError(ctx, status.JOB_MISSING_NAME, nil, errors.New("job name is required"))
	}
	return nil
}

func validateTaskName(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return errs.NewError(ctx, status.TASK_MISSING_NAME, nil, errors.New("task name is required"))
	}
	return nil
}
