package meal

type MealError struct {
	Error string `json:"error"`
}

type MealSuccess struct {
	Message string `json:"message"`
}

type RequestNewMeal struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	ScheduledAt string `json:"scheduledAt"`
	Notes       string `json:"notes"`
	GroupId     string `json:"groupId"`
}
