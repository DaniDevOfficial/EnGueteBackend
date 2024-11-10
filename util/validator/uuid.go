package validator

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

func ValidUUID(fl validator.FieldLevel) bool {
	uuidRegex := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`
	re := regexp.MustCompile(uuidRegex)
	return re.MatchString(fl.Field().String())
}
