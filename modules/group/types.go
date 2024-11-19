package group

import "enguete/modules/meal"

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
	GroupName string   `json:"groupName"`
	UserCount int      `json:"userCount"`
	UserRoles []string `json:"userRoles"`
}

type Group struct {
	GroupInfo  GroupInfo       `json:"groupInfo"`
	GroupMeals []meal.MealCard `json:"meals"`
}

type Member struct {
	Username  string   `json:"username"`
	UserId    string   `json:"userId"`
	UserRoles []string `json:"userRoles"`
}
