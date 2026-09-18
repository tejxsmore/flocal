package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func BindAndValidate(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		BadRequest(c, "Request body is missing or malformed JSON: "+err.Error())
		return false
	}
	return Validate(c, dst)
}

func BindQueryAndValidate(c *gin.Context, dst any) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		BadRequest(c, "Query parameters are missing or invalid: "+err.Error())
		return false
	}
	return Validate(c, dst)
}

func Validate(c *gin.Context, dst any) bool {
	if err := validate.Struct(dst); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			ValidationFailed(c, formatFieldErrors(verrs))
			return false
		}
		BadRequest(c, "Invalid request.")
		return false
	}
	return true
}

func formatFieldErrors(verrs validator.ValidationErrors) map[string]string {
	out := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		out[lowerFirst(fe.Field())] = fieldErrorMessage(fe)
	}
	return out
}

func fieldErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required."
	case "email":
		return "Must be a valid email address."
	case "min":
		return fmt.Sprintf("Must be at least %s characters.", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters.", fe.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s.", fe.Param())
	case "url":
		return "Must be a valid URL."
	default:
		return fmt.Sprintf("Failed validation: %s.", fe.Tag())
	}
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
