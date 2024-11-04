package validator

import (
	"github.com/go-playground/validator/v10"
	"time"
)

func ValidDateTime(fl validator.FieldLevel) bool {
	_, err := time.Parse(time.RFC3339, fl.Field().String()) // Attempts to parse any valid date-time
	return err == nil
}
