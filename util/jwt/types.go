package jwt

type JWTUser struct {
	Username string
	UserId   int
	UUID     string
}

type JWTPayload struct {
	UserId   int
	UUID     string
	UserName string
	Exp      int64
}
