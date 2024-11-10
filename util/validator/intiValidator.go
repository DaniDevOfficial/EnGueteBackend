package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("validDateTime", ValidDateTime)
		if err != nil {
			panic("Failed to register custom validator")
			return
		}
		err = v.RegisterStructValidation("validUUID", ValidUUID)
	}
}
