package domain

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details"`
	RequestID int    `json:"request_id"`
}

type ErrorDetail struct {
	Field   int    `json:"field"`
	Message string `json:"message"`
}

func NewPagination(page, limit, total int) Pagination {
	totalPgs := total / limit

	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPgs,
	}
}
