package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/satori/go.uuid"
)

func IsValidUUID(fl validator.FieldLevel) bool {
	_, err := uuid.FromString(fl.Field().String())
	return err == nil
}
