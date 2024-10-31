package hashing

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckHashedString(hash string, plainText string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainText))
	return err == nil
}
