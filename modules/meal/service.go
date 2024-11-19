package meal

import (
	"database/sql"
	"enguete/modules/group"
	"enguete/util/auth"
	"enguete/util/roles"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// basic meal functions

// CreateNewMeal godoc
// @Summary Create a new meal
// @Description Creates a new meal within a specified group. The requesting user must be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param newMeal body RequestNewMeal true "Payload to create a new meal"
// @Success 201 {object} ResponseNewMeal "Successfully created new meal with meal ID"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals [post]
func CreateNewMeal(c *gin.Context, db *sql.DB) {
	var newMeal RequestNewMeal
	err := c.ShouldBindJSON(&newMeal)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformAction(newMeal.GroupId, jwtPayload.UserId, roles.CanCreateMeal, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	mealId, err := CreateNewMealInDB(newMeal, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	log.Println("New meal created with id:", mealId)
	c.JSON(http.StatusCreated, ResponseNewMeal{MealId: mealId})
}

func GetMealById(c *gin.Context, db *sql.DB) {
	mealId := c.Param("mealId")

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	isGroupMember, err := group.IsUserInGroupViaMealId(mealId, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !isGroupMember {
		c.JSON(http.StatusNotFound, MealError{Error: "Group Not found"})
	}

	mealInformation, err := GetSingularMealInformation(mealId, jwtPayload.UserId, db)
	if err != nil {
		if errors.Is(err, ErrNoData) {
			c.JSON(http.StatusNotFound, MealError{Error: "Meal Not found"})
		}
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	participationInformation, err := GetMealParticipationInformationFromDB(mealId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	meal := Meal{
		MealInformation:            mealInformation,
		MealParticipantInformation: participationInformation,
	}
	c.JSON(http.StatusOK, meal)
}

// DeleteMeal godoc
// @Summary Delete a meal
// @Description Deletes a meal within a specified group. The requesting user must be an admin or owner of the group to perform this action.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param mealId path string true "ID of the meal to delete"
// @Success 200 {object} MealSuccess "Meal successfully deleted"
// @Failure 400 {object} MealError "Invalid meal ID"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/{mealId} [delete]
func DeleteMeal(c *gin.Context, db *sql.DB) {
	mealId := c.Param("mealId")
	if mealId == "" {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid meal id"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(mealId, jwtPayload.UserId, roles.CanDeleteMeal, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	err = DeleteMealInDB(mealId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	// here we dont se
	//nd a update, because the user will be redirected to the all page, where a api request will happen regardeless
	c.JSON(http.StatusOK, MealSuccess{Message: "Meal Sucessfuly deleted"})
}

// ChangeMealClosedFlag godoc
// @Summary Change a meal's open status
// @Description Updates a meal's open or closed status within a specified group. The requesting user must be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param updateClosedFlag body RequestUpdateClosedFlag true "Payload to update the meal status"
// @Success 200 {object} MealSuccess "Meal status successfully updated"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/open [post]
func ChangeMealClosedFlag(c *gin.Context, db *sql.DB) {
	var updateClosedFlag RequestUpdateClosedFlag
	if c.ShouldBindJSON(&updateClosedFlag) != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(updateClosedFlag.MealId, jwtPayload.UserId, roles.CanChangeMealFlags, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	err = UpdateClosedBoolInDB(updateClosedFlag.MealId, updateClosedFlag.CloseFlag, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	//TODO: send a notification to all the members of the group.

	c.JSON(http.StatusOK, MealSuccess{Message: "Meal Successfully updated"})
}

// ChangeMealFulfilledFlag godoc
// @Summary Change a meal's fulfilled status
// @Description Updates a meal's fulfilled status within a specified group. The requesting user must be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param updateFulfilledFlag body RequestUpdateFulfilledFlag true "Payload to update the meal's fulfilled status"
// @Success 200 {object} MealSuccess "Meal status successfully updated"
// @Failure 400 {object} MealError "Invalid meal ID or request body"
// @Failure 401 {object} MealError "Unauthorized - insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/fulfilled [post]
func ChangeMealFulfilledFlag(c *gin.Context, db *sql.DB) {
	var updateFulfilledFlag RequestUpdateFulfilledFlag
	if c.ShouldBindJSON(&updateFulfilledFlag) != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid meal id"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(updateFulfilledFlag.MealId, jwtPayload.UserId, roles.CanChangeMealFlags, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	err = UpdateMealFulfilledStatus(updateFulfilledFlag.MealId, updateFulfilledFlag.Fulfilled, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	//TODO: send a notification to all the members of the group.

	c.JSON(http.StatusOK, MealSuccess{Message: "Meal Successfully updated"})
}

// Meal Status per user

// OptInMeal godoc
// @Summary Opt-in to a meal
// @Description Allows a user to opt-in to a specific meal within a group. The requesting user must be a member of the group associated with the meal.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param requestOptInMeal body RequestOptInMeal true "Payload to opt-in to a meal"
// @Success 200 {object} MealSuccess "User successfully opted in to the meal"
// @Failure 400 {object} MealError "Invalid request body or user already has a preference set for this meal"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/optin [post]
func OptInMeal(c *gin.Context, db *sql.DB) {
	var requestOptInMeal RequestOptInMeal
	if err := c.ShouldBindJSON(&requestOptInMeal); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	isSelfAction := requestOptInMeal.UserId == jwtPayload.UserId
	isGroupMember, err := group.IsUserInGroupViaMealId(requestOptInMeal.MealId, requestOptInMeal.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !isGroupMember {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	if !isSelfAction {
		canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(requestOptInMeal.MealId, jwtPayload.UserId, roles.CanForceOptIn, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
			return
		}
		if !canPerformAction {
			c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
		}
	}

	err = OptInMealInDB(jwtPayload.UserId, requestOptInMeal, db)
	if err != nil {
		if errors.Is(err, ErrDataCouldNotBeUpdated) {
			c.JSON(http.StatusBadRequest, MealError{Error: "This user already has a Preference in this Meal"})
			return
		}
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	if !isSelfAction {
		//TODO: send notification to user whose opt in status got changed
	}

	c.JSON(http.StatusOK, MealSuccess{Message: "Meal Successfully OptIn"}) // TODO: later return entire meal preferences for this meal, to have valid frontend.
}

// ChangeOptInMeal godoc
// @Summary Change Opt-in status in a meal
// @Description Allows a user to change their opt-in status for a specific meal within a group. The requesting user must be a member of the group associated with the meal.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param requestOptInMeal body RequestOptInMeal true "Payload to change opt-in status in a meal"
// @Success 200 {object} MealSuccess "User's opt-in status for the meal successfully changed"
// @Failure 400 {object} MealError "Invalid request body or user already has a preference set for this meal"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/optin [put]
func ChangeOptInMeal(c *gin.Context, db *sql.DB) {
	var requestOptInMeal RequestOptInMeal

	if err := c.ShouldBindJSON(&requestOptInMeal); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	isSelfAction := requestOptInMeal.UserId == jwtPayload.UserId
	isGroupMember, err := group.IsUserInGroupViaMealId(requestOptInMeal.MealId, requestOptInMeal.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !isGroupMember {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	if !isSelfAction {
		canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(requestOptInMeal.MealId, jwtPayload.UserId, roles.CanForceOptIn, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
			return
		}
		if !canPerformAction {
			c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
		}
	}

	err = ChangeOptInStatusMealInDB(requestOptInMeal.UserId, requestOptInMeal, db)
	if err != nil {
		if errors.Is(err, ErrDataCouldNotBeUpdated) {
			c.JSON(http.StatusBadRequest, MealError{Error: "This user already has a Preference in this Meal"})
			return
		}
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	if !isSelfAction {
		//TODO: send notification to user whose opt in status got changed
	}

	c.JSON(http.StatusOK, MealSuccess{Message: "Meal Successfully OptIn changed"}) // TODO: later return entire meal preferences for this meal, to have valid frontend.
}

// Meal Cook

// AddCookToMeal godoc
// @Summary Add a cook to a meal
// @Description Adds a user as a cook to a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param addCookToMealData body RequestAddCookToMeal true "Payload to add a cook to a meal"
// @Success 201 {object} MealSuccess "Cook successfully added to meal"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/cook [post]
func AddCookToMeal(c *gin.Context, db *sql.DB) {
	var addCookToMealData RequestAddCookToMeal

	if err := c.ShouldBindJSON(&addCookToMealData); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	isSelfAdd := addCookToMealData.UserId == jwtPayload.UserId
	isGroupMember, err := group.IsUserInGroupViaMealId(addCookToMealData.MealId, addCookToMealData.UserId, db)

	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !isGroupMember {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	if !isSelfAdd {
		canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(addCookToMealData.MealId, jwtPayload.UserId, roles.CanForceAddCook, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
			return
		}
		if !canPerformAction {
			c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
		}
	}

	err = AddCookToMealInDB(addCookToMealData.UserId, addCookToMealData.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	if !isSelfAdd {
		//TODO: Send notification to user who got added as a cook
	}

	//TODO: Send an updated list of users in the meal

	c.JSON(http.StatusCreated, MealSuccess{Message: "Cook added to meal"})
}

// RemoveCookFromMeal godoc
// @Summary Remove a cook from a meal
// @Description Removes a specific user from the list of cooks in a meal. Requires the user to be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param removeCookFromMealData body RequestRemoveCook true "Payload to remove a cook from a meal"
// @Success 200 {object} MealSuccess "Cook successfully removed from meal"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/cooks [delete]
func RemoveCookFromMeal(c *gin.Context, db *sql.DB) {
	var removeCookFromMealData RequestRemoveCook
	if err := c.ShouldBindJSON(&removeCookFromMealData); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	isSelfAction := removeCookFromMealData.UserId == jwtPayload.UserId
	isGroupMember, err := group.IsUserInGroupViaMealId(removeCookFromMealData.MealId, removeCookFromMealData.UserId, db)

	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !isGroupMember {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	if !isSelfAction {
		canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(removeCookFromMealData.MealId, jwtPayload.UserId, roles.CanForceAddCook, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
			return
		}
		if !canPerformAction {
			c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
		}
	}

	err = RemoveCookFromMealInDB(removeCookFromMealData.UserId, removeCookFromMealData.MealId, db)
	if err != nil {
		if errors.Is(err, ErrUserWasntACook) {
			c.JSON(http.StatusUnauthorized, MealError{Error: "User was not a cook"})
			return
		}
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	if !isSelfAction {
		// TODO: Send notification to user who was removed as a cook
	}

	// TODO: Send an updated list of users in the meal
	c.JSON(http.StatusOK, MealSuccess{Message: "Cook removed from meal"})
}

// Update Meal Info

// UpdateMealTitle godoc
// @Summary Update a meal's title
// @Description Updates the title of a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param newTitle body RequestUpdateTitle true "Payload to update the title of a meal"
// @Success 200 {object} MealSuccess "Meal successfully updated"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/title [post]
func UpdateMealTitle(c *gin.Context, db *sql.DB) {
	var newTitle RequestUpdateTitle
	if err := c.ShouldBindJSON(&newTitle); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newTitle.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	err = UpdateMealTitleIdDB(newTitle.MealId, newTitle.NewTitle, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	//TODO: Send an updated meal information to the frontend
	c.JSON(http.StatusOK, MealSuccess{Message: "Meal updated successfully"})
}

// UpdateMealType godoc
// @Summary Update a meal's type
// @Description Update the type of a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param newType body RequestUpdateType true "Payload to update the meal type"
// @Success 200 {object} MealSuccess "Meal type successfully updated"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/type [post]
func UpdateMealType(c *gin.Context, db *sql.DB) {
	var newType RequestUpdateType
	if err := c.ShouldBindJSON(&newType); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newType.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	//TODO: Send an updated meal information to the frontend
	c.JSON(http.StatusOK, MealSuccess{Message: "Meal updated successfully"})
}

// UpdateMealNotes godoc
// @Summary Update a meal's notes
// @Description Update the notes for a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param newNotes body RequestUpdateNotes true "Payload to update the meal notes"
// @Success 200 {object} MealSuccess "Meal notes successfully updated"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/notes [put]
func UpdateMealNotes(c *gin.Context, db *sql.DB) {
	var newNotes RequestUpdateNotes
	if err := c.ShouldBindJSON(&newNotes); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newNotes.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	err = UpdateMealNotesInDB(newNotes.MealId, newNotes.NewNotes, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	//TODO: Send an updated meal information to the frontend
	c.JSON(http.StatusOK, MealSuccess{Message: "Meal updated successfully"})
}

// UpdateMealScheduledAt godoc
// @Summary Update a meal's scheduled date and time
// @Description Update the date and time when a meal will take place within a group. Requires the user to be an admin or owner of the group.
// @Tags Meals
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param newScheduledAt body RequestUpdateScheduledAt true "Payload to update the meal scheduled date and time"
// @Success 200 {object} MealSuccess "Meal scheduled date and time successfully updated"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/scheduledAt [put]
func UpdateMealScheduledAt(c *gin.Context, db *sql.DB) {
	var newScheduledAt RequestUpdateScheduledAt
	if err := c.ShouldBindJSON(&newScheduledAt); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newScheduledAt.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, MealError{Error: "You are not allowed to perform this action"})
	}

	err = UpdateMealScheduledAtInDB(newScheduledAt.MealId, newScheduledAt.NewScheduledAt, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}

	//TODO: Send an updated meal information to the frontend
	//TODO: Send a push notification to all not opt out or undecided users
	c.JSON(http.StatusOK, MealSuccess{Message: "Meal updated successfully"})
}
