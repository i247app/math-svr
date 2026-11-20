package llmroute

import "context"

// IRepository owns ma_llm_routes. There is no Create or Delete: rows are
// seeded, and an admin can only retarget one in place.
type IRepository interface {
	// FindByCode is the read every routed LLM call makes; (nil, nil) when
	// the code has no live row.
	FindByCode(ctx context.Context, code Code) (*LLMRoute, error)
	// List returns every live row, ordered by llm_route_id.
	List(ctx context.Context) ([]*LLMRoute, error)
	// UpdateTarget overwrites llm_provider and llm_model (nil clears either)
	// in one statement. It reports false when no live row has that code.
	UpdateTarget(ctx context.Context, code Code, provider, model *string, modifyId *int64) (bool, error)
}
