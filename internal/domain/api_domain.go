package domain

type PaginateRequest struct {
	Page  int
	Limit int
}

type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalPages int  `json:"total_pages"`
	Total      int  `json:"total"`
	HasPrev    bool `json:"has_prev"`
	HasNext    bool `json:"has_next"`
}

type SuccessResponse[T any] struct {
	Status int `json:"status"`
	Data   T   `json:"data"`
}

type PaginateResponse[T any] struct {
	Status     int        `json:"status"`
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type CursorPagination struct {
	NextCursor *int `json:"next_cursor"`
	HasNext    bool `json:"has_next"`
	Limit      int  `json:"limit"`
}

type CursorPaginateResponse[T any] struct {
	Status     int              `json:"status"`
	Data       []T              `json:"data"`
	Pagination CursorPagination `json:"pagination"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
