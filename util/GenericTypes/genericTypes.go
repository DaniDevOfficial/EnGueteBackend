package GenericTypes

type LastUpdatedRequest struct {
	LastUpdated *string `form:"lastUpdated" binding:"omitempty,dateTime"`
}

type LastUpdatedAndIdRequest struct {
	LastUpdated string `form:"lastUpdated" binding:"omitempty,dateTime"`
	Id          string `form:"id" binding:"required,uuid"`
}
