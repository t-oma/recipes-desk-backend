package pagination

type (
	// Request holds pagination parameters.
	Request struct {
		Page  int
		Limit int
	}
	// Result holds paginated data with metadata.
	Result[T any] struct {
		Items      []T
		Pagination Metadata
	}
	// Metadata holds pagination metadata.
	Metadata struct {
		Page       int
		Limit      int
		Total      int64
		TotalPages int
	}
)

// Skip calculates skip value.
func (r *Request) Skip() int64 {
	return int64((r.Page - 1) * r.Limit)
}

// NewResult creates a paginated result with metadata.
func NewResult[T any](items []T, req *Request, total int64) Result[T] {
	if req.Limit < 1 {
		req.Limit = 1
	}
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	return Result[T]{
		Items: items,
		Pagination: Metadata{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
