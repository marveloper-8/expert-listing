package models

type StandardResponse struct {
	Success    bool                `json:"success" example:"true"`
	StatusCode int                 `json:"status_code" example:"200"`
	Message    string              `json:"message" example:"Operation successful"`
	Data       interface{}         `json:"data,omitempty"`
	Pagination *PaginationMetadata `json:"pagination,omitempty"`
	Error      string              `json:"error,omitempty" example:""`
	Errors     []FieldError        `json:"errors,omitempty"`
}

type PaginationMetadata struct {
	CurrentPage int   `json:"current_page" example:"1"`
	PageSize    int   `json:"page_size" example:"10"`
	TotalItems  int64 `json:"total_items" example:"42"`
	TotalPages  int   `json:"total_pages" example:"5"`
	HasNextPage bool  `json:"has_next_page" example:"true"`
	HasPrevPage bool  `json:"has_prev_page" example:"false"`
}

type FieldError struct {
	Field   string `json:"field" example:"price"`
	Message string `json:"message" example:"price must be greater than or equal to 0"`
}

func NewSuccessResponse(statusCode int, message string, data interface{}) StandardResponse {
	return StandardResponse{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
}

func NewPaginatedResponse(statusCode int, message string, data interface{}, pagination *PaginationMetadata) StandardResponse {
	return StandardResponse{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

func NewErrorResponse(statusCode int, message string, errDetail string) StandardResponse {
	return StandardResponse{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
		Error:      errDetail,
	}
}

func NewValidationErrorResponse(message string, fieldErrors []FieldError) StandardResponse {
	return StandardResponse{
		Success:    false,
		StatusCode: 422,
		Message:    message,
		Errors:     fieldErrors,
	}
}
