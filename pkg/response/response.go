package response

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// APIResponse represents a standardized API response
type APIResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	APIResponse
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse creates a success response
func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Status:    "success",
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// ErrorResponse creates an error response
func ErrorResponse(message string) *APIResponse {
	return &APIResponse{
		Status:    "error",
		Error:     message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// ErrorResponseWithCode creates an error response with error code
func ErrorResponseWithCode(message, code string) *APIResponse {
	return &APIResponse{
		Status:    "error",
		Error:     message,
		Code:      code,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// StatusResponse creates a status response
func StatusResponse(status, message string) *APIResponse {
	return &APIResponse{
		Status:    status,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// PaginatedSuccessResponse creates a paginated success response
func PaginatedSuccessResponse(data interface{}, pagination Pagination) *PaginatedResponse {
	return &PaginatedResponse{
		APIResponse: APIResponse{
			Status:    "success",
			Data:      data,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
		Pagination: pagination,
	}
}

// RespondJSON sends a JSON response
func RespondJSON(ctx echo.Context, statusCode int, response *APIResponse) error {
	// Add request ID if available
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	// Log response for debugging
	zap.L().Debug("sending response",
		zap.Int("status_code", statusCode),
		zap.String("status", response.Status),
		zap.String("request_id", response.RequestID),
	)

	return ctx.JSON(statusCode, response)
}

// RespondPaginatedJSON sends a paginated JSON response
func RespondPaginatedJSON(ctx echo.Context, statusCode int, response *PaginatedResponse) error {
	// Add request ID if available
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	// Log response for debugging
	zap.L().Debug("sending paginated response",
		zap.Int("status_code", statusCode),
		zap.String("status", response.Status),
		zap.String("request_id", response.RequestID),
		zap.Int("page", response.Pagination.Page),
		zap.Int("per_page", response.Pagination.PerPage),
		zap.Int64("total", response.Pagination.Total),
	)

	return ctx.JSON(statusCode, response)
}

// RespondError sends an error response
func RespondError(ctx echo.Context, statusCode int, message string) error {
	return RespondJSON(ctx, statusCode, ErrorResponse(message))
}

// RespondErrorWithCode sends an error response with error code
func RespondErrorWithCode(ctx echo.Context, statusCode int, message, code string) error {
	return RespondJSON(ctx, statusCode, ErrorResponseWithCode(message, code))
}

// RespondStatus sends a status response
func RespondStatus(ctx echo.Context, status, message string) error {
	return RespondJSON(ctx, http.StatusOK, StatusResponse(status, message))
}

// RespondSuccess sends a success response
func RespondSuccess(ctx echo.Context, data interface{}) error {
	return RespondJSON(ctx, http.StatusOK, SuccessResponse(data))
}

// RespondCreated sends a created response
func RespondCreated(ctx echo.Context, data interface{}) error {
	return RespondJSON(ctx, http.StatusCreated, SuccessResponse(data))
}

// RespondAccepted sends an accepted response
func RespondAccepted(ctx echo.Context, message string) error {
	return RespondJSON(ctx, http.StatusAccepted, StatusResponse("accepted", message))
}

// RespondNoContent sends a no content response
func RespondNoContent(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

// RespondBadRequest sends a bad request response
func RespondBadRequest(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusBadRequest, message)
}

// RespondUnauthorized sends an unauthorized response
func RespondUnauthorized(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusUnauthorized, message)
}

// RespondForbidden sends a forbidden response
func RespondForbidden(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusForbidden, message)
}

// RespondNotFound sends a not found response
func RespondNotFound(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusNotFound, message)
}

// RespondInternalError sends an internal server error response
func RespondInternalError(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusInternalServerError, message)
}

// RespondServiceUnavailable sends a service unavailable response
func RespondServiceUnavailable(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusServiceUnavailable, message)
}

// RespondTooManyRequests sends a too many requests response
func RespondTooManyRequests(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusTooManyRequests, message)
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Message   string            `json:"message"`
	Features  []string          `json:"features"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
}

// RespondHealthCheck sends a health check response
func RespondHealthCheck(ctx echo.Context, status, version, message string, features []string, services map[string]string) error {
	response := &HealthCheckResponse{
		Status:    status,
		Version:   version,
		Message:   message,
		Features:  features,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}

	// Add request ID if available
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		// We can extend HealthCheckResponse to include RequestID if needed
	}

	return ctx.JSON(http.StatusOK, response)
}
