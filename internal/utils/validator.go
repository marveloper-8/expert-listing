package utils

import (
	"errors"
	"fmt"
	"strings"

	"expertlisting/internal/models"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) []models.FieldError {
	var errs []models.FieldError

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			namespace := fe.Namespace()
			parts := strings.Split(namespace, ".")
			var fieldName string
			if len(parts) > 1 {
				var cleanParts []string
				for _, p := range parts[1:] {
					cleanParts = append(cleanParts, toSnakeCase(p))
				}
				fieldName = strings.Join(cleanParts, ".")
			} else {
				fieldName = toSnakeCase(fe.Field())
			}

			message := formatMessage(fieldName, fe.Tag(), fe.Param())
			errs = append(errs, models.FieldError{
				Field:   fieldName,
				Message: message,
			})
		}
		return errs
	}

	errs = append(errs, models.FieldError{
		Field:   "body",
		Message: err.Error(),
	})
	return errs
}

func formatMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters or value", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters or value", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude (-90 to 90)", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude (-180 to 180)", field)
	default:
		return fmt.Sprintf("%s failed validation rule '%s'", field, tag)
	}
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
