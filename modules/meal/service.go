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

// CreateNewMeal @Summary Create a new meal
// @Description Creates a new meal within a specified group. The requesting user must be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
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

// DeleteMeal @Summary Delete a new meal
// @Description Delete a meal within a specified group. The requesting user must be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param newMeal params String true "Payload to create a new meal"
// @Success 201 {object} MealSuccess "Successfully created new meal with meal ID"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals [delete]
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

// ChangeMealClosedFlag @Summary Change a meals open status
// @Description Update a meals open status within a specified group. The requesting user must be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param updateClosedFlag body RequestUpdateClosedFlag true "Payload to update the status"
// @Success 201 {object} MealSuccess "Successfully created new meal with meal ID"
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

// OptInMeal @Summary Opt-in to a meal
// @Description Allows a user to opt-in to a specific meal within a group. The requesting user must be a member of the group associated with the meal.
// @Tags meals
// @Accept json
// @Produce json
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

// ChangeOptInMeal @Summary Change Opt-in in a meal
// @Description Allows a user to change opt-in status to a specific meal within a group. The requesting user must be a member of the group associated with the meal.
// @Tags meals
// @Accept json
// @Produce json
// @Param requestOptInMeal body RequestOptInMeal true "Payload to chnage opt-in status in a meal"
// @Success 200 {object} MealSuccess "User successfully opted in to the meal"
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

// AddCookToMeal @Summary Add a cook to a meal
// @Description Adds a user as a cook to a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
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

// RemoveCookFromMeal @Summary Remove a cook from a meal
// @Description Remove a specific user from the list of cooks in a meal. Requires the user to be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param removeCookFromMealData body RequestRemoveCook true "Payload to remove a cook from a meal"
// @Success 200 {object} MealSuccess "Cook successfully removed from meal"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/remove-cook [post]
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

// UpdateMealTitle @Summary Updated a meals Title
// @Description Update the Title on a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param newTitle body RequestUpdateTitle true "Payload to add a cook to a meal"
// @Success 201 {object} MealSuccess "Meal successfully updated"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/ [post]
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

// UpdateMealType @Summary Updated a meals Type
// @Description Update the Type on a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param newTitle body RequestUpdateTitle true "Payload to add a cook to a meal"
// @Success 201 {object} MealSuccess "Meal successfully updated"
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

// UpdateMealNotes @Summary Updated a meals Notes
// @Description Update the Notes on a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param newTitle body RequestUpdateNotes true "Payload to add a cook to a meal"
// @Success 201 {object} MealSuccess "Meal successfully updated"
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

// UpdateMealScheduledAt @Summary Updated a meals happening date
// @Description Update the dateTime when a meal will take place on a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param newTitle body RequestUpdateScheduledAt true "Payload to add a cook to a meal"
// @Success 201 {object} MealSuccess "Meal successfully updated"
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
