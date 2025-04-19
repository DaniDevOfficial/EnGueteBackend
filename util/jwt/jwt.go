package jwt

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("capybara") // TODO: add secret key via .env or some rotation

type NewRefreshTokenDataDB struct {
	UserId       string     `json:"userId"`
	RefreshToken string     `json:"refresh_token"`
	LifeTime     *time.Time `json:"lifeTime"`
}

func CreateToken(userData JWTUser) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"UserId":   userData.UserId,
			"UserName": userData.Username,
			"Exp":      time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, nil
	}
	return true, nil
}

var RefreshTokenNotInDbError = errors.New("refresh token not found in database")
var RefreshTokenIsNotValidDueToExpirationDate = errors.New("refresh token not valid due to expiration date")

func VerifyRefreshToken(tokenString string, db *sql.DB) (JWTPayload, error) {

	splitToken := strings.Split(tokenString, ".")
	if len(splitToken) != 3 {
		return JWTPayload{}, fmt.Errorf("invalid token format")
	}

	payloadSegment := splitToken[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSegment)
	if err != nil {
		return JWTPayload{}, fmt.Errorf("failed to decode payload: %v", err)
	}

	var payload JWTPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return JWTPayload{}, fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	if payload.Exp > 0 {
		if payload.Exp < time.Now().Unix()-3600*24*8 {
			return payload, RefreshTokenIsNotValidDueToExpirationDate
		}
	}

	inDB, err := VerifyRefreshTokenInDB(tokenString, payload.UserId, db)
	if err != nil {
		return payload, err
	}
	if !inDB {
		return payload, RefreshTokenNotInDbError
	}

	return payload, nil
}

func VerifyRefreshTokenInDB(token string, userId string, db *sql.DB) (bool, error) {
	var count int64

	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE refresh_token = ?
		AND user_id = ?
	`

	err := db.QueryRow(query, token, userId).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func CreateRefreshToken(userData JWTUser, isTimeBased bool, db *sql.DB) (string, error) {
	var dateTime *time.Time
	if isTimeBased {
		t := time.Now().Add(time.Hour * 24 * 14)
		dateTime = &t
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"UserId":   userData.UserId,
			"Username": userData.Username,
			"Exp": func() int64 {
				if dateTime != nil {
					return dateTime.Unix()
				}
				return 0
			}(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	data := NewRefreshTokenDataDB{
		UserId:       userData.UserId,
		RefreshToken: tokenString,
		LifeTime:     dateTime,
	}
	err = PushRefreshTokenToDB(data, db)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return tokenString, nil
}

func DecodeBearer(tokenString string) (JWTPayload, error) {
	splitToken := strings.Split(tokenString, ".")
	if len(splitToken) != 3 {
		return JWTPayload{}, fmt.Errorf("invalid token format")
	}

	payloadSegment := splitToken[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSegment)
	if err != nil {
		return JWTPayload{}, fmt.Errorf("failed to decode payload: %v", err)
	}

	var payload JWTPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return JWTPayload{}, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	return payload, nil
}

func PushRefreshTokenToDB(data NewRefreshTokenDataDB, db *sql.DB) error {
	if data.LifeTime != nil && data.LifeTime.IsZero() {
		data.LifeTime = nil
	}
	sqlString := `
	INSERT INTO refresh_tokens 
	(user_id, refresh_Token) 
	VALUES ($1, $2)
`
	log.Println(data)
	_, err := db.Exec(sqlString, data.UserId, data.RefreshToken)

	return err
}

func VoidRefreshTokenInDB(token string, db *sql.DB) error {
	sqlString := `
	DELETE FROM refresh_tokens 
	WHERE refresh_token = $1
`
	_, err := db.Exec(sqlString, token)
	return err
}
