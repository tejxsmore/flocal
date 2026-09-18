package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

const (
	CodeValidation      = "VALIDATION_ERROR"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeRateLimited     = "RATE_LIMITED"
	CodeInternal        = "INTERNAL_ERROR"
	CodeBadRequest      = "BAD_REQUEST"
	CodeUpstreamFailure = "UPSTREAM_FAILURE"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func WithMeta(c *gin.Context, data, meta any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, Envelope{Success: false, Error: &Error{Code: code, Message: message}})
}

func FailWithDetails(c *gin.Context, status int, code, message string, details any) {
	c.JSON(status, Envelope{Success: false, Error: &Error{Code: code, Message: message, Details: details}})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, message)
}

func ValidationFailed(c *gin.Context, fieldErrors map[string]string) {
	FailWithDetails(c, http.StatusUnprocessableEntity, CodeValidation, "One or more fields are invalid.", fieldErrors)
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Authentication is required to access this resource."
	}
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, CodeForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "The requested resource was not found."
	}
	Fail(c, http.StatusNotFound, CodeNotFound, message)
}

func Conflict(c *gin.Context, message string) {
	Fail(c, http.StatusConflict, CodeConflict, message)
}

func RateLimited(c *gin.Context, message string) {
	if message == "" {
		message = "Too many requests. Please slow down and try again shortly."
	}
	Fail(c, http.StatusTooManyRequests, CodeRateLimited, message)
}

func Internal(c *gin.Context, message string) {
	if message == "" {
		message = "Something went wrong on our end. Please try again."
	}
	Fail(c, http.StatusInternalServerError, CodeInternal, message)
}

func UpstreamFailure(c *gin.Context, message string) {
	Fail(c, http.StatusBadGateway, CodeUpstreamFailure, message)
}
