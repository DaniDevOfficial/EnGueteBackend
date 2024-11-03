package jwt

type JWTUser struct {
	Username string
	UserId   string
}

type JWTPayload struct {
	UserId   string
	UserName string
	Exp      int64
}

type JWTTokenResponse struct {
	Token string `json:"token"`
}
