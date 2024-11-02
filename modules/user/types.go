package user

type UserError struct {
	Error string `json:"error"`
}

type UserSuccess struct {
	Message string `json:"message"`
}

type RequestNewUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

type UserGroupsFromDB struct {
	userCount int
	groupName string
	groupId   string
}
