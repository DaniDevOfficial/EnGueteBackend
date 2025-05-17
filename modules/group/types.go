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
type RequestIdGroup struct {
	GroupId string `form:"groupId" binding:"required,uuid"`
}

type RequestGroupMeals struct {
	GroupId    string  `form:"groupId" binding:"required,uuid"`
	FilterDate *string `form:"filterDate" binding:"dateTime"`
}

type RequestInviteToken struct {
	InviteToken string `form:"inviteToken" binding:"required,uuid"`
}

type RequestUpdateGroupName struct {
	GroupId   string `json:"groupId" binding:"required,uuid"`
	GroupName string `json:"groupName" binding:"required"`
}

type InviteLinkGenerationRequest struct {
	GroupId            string `json:"groupId" binding:"required,uuid"`
	ExpirationDateTime string `json:"expiresAt" binding:"required,dateTime"`
}

type InviteToken struct {
	InviteToken string `json:"inviteToken"`
	ExpiresAt   string `json:"expiresAt"`
}

type InviteLinkGenerationResponse struct {
	InviteLink  string `json:"inviteLink"`
	InviteToken string `json:"inviteToken"`
}

type ResponseGroupId struct {
	GroupId string `json:"groupId"`
}

type GroupInfo struct {
	GroupId        string   `json:"groupId"`
	GroupName      string   `json:"groupName"`
	UserCount      int      `json:"userCount"`
	UserRoles      []string `json:"userRoles"`
	UserRoleRights []string `json:"userRoleRights"`
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
	GroupId   string   `json:"groupId"`
	UserId    string   `json:"userId"`
	UserRoles []string `json:"userRoles"`
}

type FilterGroupRequest struct {
	GroupId         string  `form:"groupId" binding:"required"`
	IsFulfilled     int     `form:"isFulfilled"`
	IsOpen          int     `form:"isOpen"`
	WeekFilter      *string `form:"weekFilter"`
	StartDateFilter *string
	EndDateFilter   *string
	AmICook         int    `form:"amICook"`
	Preference      string `form:"preference"`
}

type AllGroupsSyncResponse struct {
	Groups     []GroupInfo `json:"groups"`
	DeletedIds []string    `json:"deletedIds"`
}
