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
	ScheduledAt string `json:"scheduledAt" binding:"required,validDateTime"`
	Notes       string `json:"notes" `
	GroupId     string `json:"groupId" binding:"required"`
}

type RequestOptInMeal struct {
	MealId     string `json:"mealId" binding:"required"`
	Preference string `json:"preference" binding:"required"`
}

type RequestUpdateTitle struct {
	NewTitle string `json:"newTitle" binding:"required"`
	MealId   string `json:"mealId" binding:"required"`
}
type RequestUpdateType struct {
	NewType string `json:"newType" binding:"required"`
	MealId  string `json:"mealId" binding:"required"`
}
type RequestUpdateNotes struct {
	NewNotes string `json:"newNotes" binding:"required"`
	MealId   string `json:"mealId" binding:"required"`
}
type RequestUpdateScheduledAt struct {
	NewScheduledAt string `json:"newScheduledAt" binding:"required,validDateTime"`
	MealId         string `json:"mealId" binding:"required"`
}

type RequestAddCookToMeal struct {
	UserId  string `json:"userId" binding:"required"`
	GroupId string `json:"groupId" binding:"required"`
	MealId  string `json:"mealId" binding:"required"`
}
type RequestRemoveCook struct {
	UserId  string `json:"userId" binding:"required"`
	GroupId string `json:"groupId" binding:"required"`
	MealId  string `json:"mealId" binding:"required"`
}

type ResponseNewMeal struct {
	MealId string `json:"mealId"`
}
