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
	GroupId            string `json:"groupId" binding:"required,uuid"`
	ExpirationDateTime string `json:"expirationDateTime"`
}

type InviteLinkGenerationResponse struct {
	InviteLink string `json:"inviteLink"`
}

type ResponseNewGroup struct {
	GroupId string `json:"groupId"`
}

type GroupInfo struct {
	GroupId   string   `json:"groupId"`
	GroupName string   `json:"groupName"`
	UserCount int      `json:"userCount"`
	UserRoles []string `json:"userRoles"`
}

type Group struct {
	GroupInfo  GroupInfo  `json:"groupInfo"`
	GroupMeals []MealCard `json:"meals"`
}

type MealCard struct {
	MealId           string `json:"mealId"`
	Title            string `json:"title"`
	Closed           bool   `json:"closed"`
	Fulfilled        bool   `json:"fulfilled"`
	DateTime         string `json:"dateTime"`
	MealType         string `json:"mealType"`
	Notes            string `json:"notes"`
	ParticipantCount int    `json:"participantCount"`
	UserPreference   string `json:"userPreference"`
	IsCook           bool   `json:"isCook"`
}

type Member struct {
	Username  string   `json:"username"`
	UserId    string   `json:"userId"`
	UserRoles []string `json:"userRoles"`
}
