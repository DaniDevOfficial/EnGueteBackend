package validation

import (
	"fmt"
	"log"
	"strconv"
)

const (
	minLength = 8
	maxLength = 127
)

var PasswordFormatNeedsUpperLowerSpecialError = fmt.Errorf("password needs Upper, lower and special characters and at least one number")
var PasswordFormatTooShortError = fmt.Errorf("The Password needs to be at least " + strconv.Itoa(minLength) + " letters long")
var PasswordToLongError = fmt.Errorf("The Password is to long, max length is " + strconv.Itoa(maxLength) + " letters")

func IsValidPassword(password string) error {
	if len(password) < minLength {
		return PasswordFormatTooShortError
	}
	if len(password) > maxLength {
		return PasswordToLongError
	}
	if !checkForCharacters(password) {
		return PasswordFormatNeedsUpperLowerSpecialError
	}

	return nil
}

func checkForCharacters(password string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		ascii := int(char)

		if !hasUpper && ascii >= 65 && ascii <= 90 {
			hasUpper = true
		} else if !hasLower && ascii >= 97 && ascii <= 122 {
			hasLower = true
		} else if !hasDigit && ascii >= 48 && ascii <= 57 {
			hasDigit = true
		} else if !hasSpecial && ((ascii >= 33 && ascii <= 47) || (ascii >= 58 && ascii <= 64) || (ascii >= 91 && ascii <= 96) || (ascii >= 123 && ascii <= 126)) {
			hasSpecial = true
		}
		if hasUpper && hasLower && hasDigit && hasSpecial {
			break
		}
	}
	fmt.Printf("Upper: %v, Lower: %v, Digit: %v, Special: %v\n", hasUpper, hasLower, hasDigit, hasSpecial)
	log.Println(hasUpper, hasLower, hasDigit, hasSpecial)
	return hasUpper && hasLower && hasDigit && hasSpecial
}
