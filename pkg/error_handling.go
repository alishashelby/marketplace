package pkg

// ErrorResponse represents a standardized error response
// @Description This is the standard error response format for all API endpoints
type ErrorResponse struct {
	Error string `json:"error" example:"error description"`
}

// ValidationErrorResponse represents validation errors
// @Description This is the validation error response format
type ValidationErrorResponse struct {
	Errors map[string]string `json:"errors" example:"error: error description"`
}
