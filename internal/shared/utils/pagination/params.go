package pagination

// Params represents common pagination and sorting parameters
// This struct can be embedded in repository-specific parameter structs
type Params struct {
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

// NewParams creates a new Params with default values
func NewParams(page, limit int64) Params {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > 1000 {
		limit = 1000 // Max limit cap
	}

	return Params{
		Page:      page,
		Limit:     limit,
		OrderBy:   "",
		OrderDesc: false,
		TakeAll:   false,
	}
}

// WithOrderBy sets the order by field
func (p Params) WithOrderBy(field string, desc bool) Params {
	p.OrderBy = field
	p.OrderDesc = desc
	return p
}

// WithTakeAll sets take all flag
func (p Params) WithTakeAll(takeAll bool) Params {
	p.TakeAll = takeAll
	return p
}
