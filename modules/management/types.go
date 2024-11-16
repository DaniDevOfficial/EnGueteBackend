package management

type ManagementError struct {
	Error string `json:"error"`
}

type ManagementSuccess struct {
	Message string `json:"message"`
}

type RequestKickUser struct {
	UserId  string `json:"userId" binding:"required,uuid"`
	GroupId string `json:"groupId" binding:"required,uuid"`
}
type RequestRoleData struct {
	UserId  string `json:"userId" binding:"required,uuid"`
	GroupId string `json:"groupId" binding:"required,uuid"`
	Role    string `json:"role" binding:"required"`
}
