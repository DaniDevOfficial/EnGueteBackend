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
