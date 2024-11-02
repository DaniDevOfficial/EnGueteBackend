package group

type GroupError struct {
	Error string `json:"error"`
}

type GroupSuccess struct {
	Message string `json:"message"`
}

type RequestNewGroup struct {
	GroupName string `json:"groupName"`
}
type RequestChangeUsername struct {
	Username string `json:"username"`
}
type RequestChangePassword struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}
type DBNewUser struct {
	username      string
	email         string
	password_hash string
}

type SimpleLoginUser struct {
	Username     string
	PasswordHash string
}

type SignInCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserFromDB struct {
	userId       string
	username     string
	email        string
	passwordHash string
}
