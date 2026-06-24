package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct runs struct tag validation and returns a human-readable error.
func ValidateStruct(s any) error {
	if err := validate.Struct(s); err != nil {
		var msgs []string
		for _, fe := range err.(validator.ValidationErrors) {
			msgs = append(msgs, formatFieldError(fe))
		}
		return fmt.Errorf("%s", strings.Join(msgs, "; "))
	}
	return nil
}

func formatFieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + fe.Param()
	case "max":
		return field + " must be at most " + fe.Param()
	case "gte":
		return field + " must be >= " + fe.Param()
	case "lte":
		return field + " must be <= " + fe.Param()
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "uuid":
		return field + " must be a valid UUID"
	case "url":
		return field + " must be a valid URL"
	default:
		return field + " failed validation on " + fe.Tag()
	}
}
