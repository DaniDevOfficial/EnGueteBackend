package meal

type MealError struct {
	Error string `json:"error"`
}

type MealSuccess struct {
	Message string `json:"message"`
}

type RequestNewMeal struct {
	Title       string `json:"title" binding:"required"`
	Type        string `json:"type" binding:"required"`
	ScheduledAt string `json:"scheduledAt" binding:"required,dateTime"`
	Notes       string `json:"notes" `
	GroupId     string `json:"groupId" binding:"required,uuid"`
}

type RequestOptInMeal struct {
	MealId     string `json:"mealId" binding:"required,uuid"`
	UserId     string `json:"userId" binding:"required,uuid"`
	Preference string `json:"preference" binding:"required"`
}

type RequestUpdateClosedFlag struct {
	MealId    string `json:"mealId" binding:"required,uuid"`
	CloseFlag bool   `json:"closeFlag" binding:"required"`
}

type RequestUpdateFulfilledFlag struct {
	MealId    string `json:"mealId" binding:"required,uuid"`
	Fulfilled bool   `json:"fulfilled" binding:"required"`
}

type RequestUpdateTitle struct {
	NewTitle string `json:"newTitle" binding:"required"`
	MealId   string `json:"mealId" binding:"required,uuid"`
}
type RequestUpdateType struct {
	NewType string `json:"newType" binding:"required"`
	MealId  string `json:"mealId" binding:"required,uuid"`
}
type RequestUpdateNotes struct {
	NewNotes string `json:"newNotes" binding:"required"`
	MealId   string `json:"mealId" binding:"required,uuid"`
}
type RequestUpdateScheduledAt struct {
	NewScheduledAt string `json:"newScheduledAt" binding:"required,dateTime"`
	MealId         string `json:"mealId" binding:"required,uuid"`
}

type RequestAddCookToMeal struct {
	UserId string `json:"userId" binding:"required"`
	MealId string `json:"mealId" binding:"required,uuid"`
}
type RequestRemoveCook struct {
	UserId string `json:"userId" binding:"required"`
	MealId string `json:"mealId" binding:"required,uuid"`
}

type ResponseNewMeal struct {
	MealId string `json:"mealId"`
}

type MealCard struct {
	MealID           string `json:"mealId"`
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
