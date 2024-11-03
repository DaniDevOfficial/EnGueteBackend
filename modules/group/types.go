package group

type GroupError struct {
	Error string `json:"error"`
}

type GroupSuccess struct {
	Message string `json:"message"`
}

type RequestNewGroup struct {
	GroupName string `json:"groupName" binding:"required"`
}

type InviteLinkGenerationRequest struct {
	GroupId            string `json:"groupId" binding:"required"`
	ExpirationDateTime string `json:"expirationDateTime"`
}

type InviteLinkGenerationResponse struct {
	InviteLink string `json:"inviteLink"`
}

type ResponseNewGroup struct {
	GroupId string `json:"groupId"`
}
