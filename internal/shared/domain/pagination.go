package domain

type PaginationParams struct {
	Page  int
	Limit int
}

func NewPaginationParams(page, limit int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return PaginationParams{Page: page, Limit: limit}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}
