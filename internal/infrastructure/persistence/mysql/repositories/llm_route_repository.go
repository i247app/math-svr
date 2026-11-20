package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/domain/llmroute"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	llmRouteTable = "ma_llm_routes"

	llmRouteActiveWhere = `r.status IN (?) AND r.deleted_dt IS NULL`
)

var llmRouteColumns = `r.llm_route_id, r.route_code, r.llm_provider, r.llm_model, r.description, r.llm_route_status, ` + models.AuditColumns("r")

func llmRouteActiveArgs() []any {
	return []any{enum.StatusActive}
}

type LLMRouteRepository struct {
	db database.Executor
}

func NewLLMRouteRepository(db database.Executor) llmroute.IRepository {
	return &LLMRouteRepository{db: db}
}

func scanLLMRoute(s database.RowScanner) (*models.LLMRouteModel, error) {
	var m models.LLMRouteModel
	if err := s.Scan(append([]any{&m.LLMRouteId, &m.RouteCode, &m.LLMProvider, &m.LLMModel, &m.Description, &m.LLMRouteStatus}, m.Audit.Dest()...)...); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *LLMRouteRepository) FindByCode(ctx context.Context, code llmroute.Code) (*llmroute.LLMRoute, error) {
	query := `SELECT ` + llmRouteColumns + ` FROM ` + llmRouteTable + ` r WHERE (r.route_code = ?) AND ` + llmRouteActiveWhere
	args := slices.Concat([]any{code.String()}, llmRouteActiveArgs())

	m, err := scanLLMRoute(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("llm route repo find %s: %w", code, err)
	}
	return ModelToDomainLLMRoute(m), nil
}

func (r *LLMRouteRepository) List(ctx context.Context) ([]*llmroute.LLMRoute, error) {
	query := `SELECT ` + llmRouteColumns + ` FROM ` + llmRouteTable + ` r WHERE ` + llmRouteActiveWhere + ` ORDER BY r.llm_route_id`

	rows, err := r.db.Query(ctx, query, llmRouteActiveArgs()...)
	if err != nil {
		return nil, fmt.Errorf("llm route repo list: %w", err)
	}
	defer rows.Close()

	var out []*llmroute.LLMRoute
	for rows.Next() {
		m, err := scanLLMRoute(rows)
		if err != nil {
			return nil, fmt.Errorf("llm route repo list scan: %w", err)
		}
		out = append(out, ModelToDomainLLMRoute(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("llm route repo list rows: %w", err)
	}
	return out, nil
}

// UpdateTarget sets both columns as given — no COALESCE, a nil clears the
// column. modify_dt is set explicitly so the row always changes: MySQL
// reports CHANGED rows by default (no clientFoundRows on this driver), and
// re-saving the same target would otherwise read as "no such row".
func (r *LLMRouteRepository) UpdateTarget(ctx context.Context, code llmroute.Code, provider, model *string, modifyId *int64) (bool, error) {
	query := `UPDATE ` + llmRouteTable + ` r SET r.llm_provider = ?, r.llm_model = ?, r.modify_id = ?, r.modify_dt = CURRENT_TIMESTAMP(6)` +
		` WHERE (r.route_code = ?) AND ` + llmRouteActiveWhere
	args := slices.Concat([]any{provider, model, modifyId, string(code)}, llmRouteActiveArgs())

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("llm route repo update %s: %w", code, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("llm route repo update %s rows affected: %w", code, err)
	}
	return n > 0, nil
}

func ModelToDomainLLMRoute(m *models.LLMRouteModel) *llmroute.LLMRoute {
	e := llmroute.NewLLMRoute()
	e.SetLLMRouteId(m.LLMRouteId)
	e.SetRouteCode(llmroute.Code(m.RouteCode))
	e.SetLLMProvider(m.LLMProvider)
	e.SetLLMModel(m.LLMModel)
	e.SetDescription(m.Description)
	e.SetLLMRouteStatus(m.LLMRouteStatus)
	m.Audit.ApplyTo(&e.Audit)
	return e
}
