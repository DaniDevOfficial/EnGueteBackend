package user

type UserError struct {
	Error string `json:"error"`
}

type UserSuccess struct {
	Message string `json:"message"`
}

type RequestNewUser struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type RequestChangeUsername struct {
	Username string `json:"username" binding:"required"`
}
type RequestChangePassword struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type ResponseUserGroups struct {
	Groups []GroupCard `json:"userGroups" `
}

type ResponseUserData struct {
	Username string      `json:"username"`
	UserID   string      `json:"userId"`
	Groups   []GroupCard `json:"groups"`
}

type GroupCard struct {
	GroupName    string `json:"groupName"`
	GroupId      string `json:"groupId"`
	UserCount    int    `json:"userCount"`
	NextMealDate string `json:"nextMealDate"`
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
	UserId       string `json:"userId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

type UserGroupsFromDB struct {
	userCount int
	groupName string
	groupId   string
}

type MessageResponse struct {
	Message string `json:"message"`
}
