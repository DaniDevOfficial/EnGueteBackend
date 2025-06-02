package meal

import (
	"database/sql"
	"enguete/modules/group"
	"enguete/util/auth"
	"enguete/util/frontendErrors"
	"enguete/util/responses"
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformAction(newMeal.GroupId, jwtPayload.UserId, roles.CanCreateMeal, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
		return
	}

	mealId, err := CreateNewMealInDBWithTransaction(newMeal, jwtPayload.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusCreated, ResponseNewMeal{MealId: mealId})
}

func GetMealById(c *gin.Context, db *sql.DB) {
	var mealInfo RequestMealId
	if err := c.ShouldBindQuery(&mealInfo); err != nil {
		log.Println(err)
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	groupId, err := group.IsUserInGroupViaMealId(mealInfo.MealId, jwtPayload.UserId, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	mealInformation, err := GetSingularMealInformation(mealInfo.MealId, jwtPayload.UserId, db)
	if err != nil {
		if errors.Is(err, ErrNoData) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.MealDoesNotExistError, "Meal does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	participationInformationWithPreference, err := GetMealParticipationInformationFromDB(mealInfo.MealId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	//TODO: if meal is closed dont get this data and just say that the ones with a preference are sent back (this is so we have accurate historical data)
	participationInformationWithoutPreference, err := GetGroupMembersNotParticipatingInMeal(mealInfo.MealId, groupId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	participationInformation := MergeAndSortParticipants(participationInformationWithPreference, participationInformationWithoutPreference)
	meal := Meal{
		MealInformation:           mealInformation,
		MealPreferenceInformation: participationInformation,
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
	var requestData RequestMealId
	if err := c.ShouldBindQuery(&requestData); err != nil {
		log.Println(err)
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(requestData.MealId, jwtPayload.UserId, roles.CanDeleteMeal, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
		return
	}

	err = DeleteMealInDB(requestData.MealId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	// here we dont send a update, because the user will be redirected to the all page, where a api request will happen regardeless
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
func ChangeMealClosedFlag(c *gin.Context, db *sql.DB) { //TODO: Functionality of this needs to be rew
	var updateClosedFlag RequestUpdateClosedFlag
	if c.ShouldBindJSON(&updateClosedFlag) != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(updateClosedFlag.MealId, jwtPayload.UserId, roles.CanChangeMealFlags, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
		return
	}

	err = UpdateClosedBoolInDB(updateClosedFlag.MealId, updateClosedFlag.CloseFlag, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)

		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(updateFulfilledFlag.MealId, jwtPayload.UserId, roles.CanChangeMealFlags, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
		return
	}

	err = UpdateMealFulfilledStatus(updateFulfilledFlag.MealId, updateFulfilledFlag.Fulfilled, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}
	//TODO: send a notification to all the members of the group.

	c.JSON(http.StatusOK, MealSuccess{Message: "Meal Successfully updated"})
}

// Meal Status per user

// Preferences

func UpdatePreference(c *gin.Context, db *sql.DB) {
	var updatePreference RequestUpdatePreference
	if err := c.ShouldBindJSON(&updatePreference); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	if updatePreference.Preference == nil && updatePreference.IsCook == nil {
		c.JSON(http.StatusOK, MealSuccess{Message: "No changes made"})
		return
	}

	isSelfAction := updatePreference.UserId == jwtPayload.UserId

	_, err = group.IsUserInGroupViaMealId(updatePreference.MealId, updatePreference.UserId, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	if !isSelfAction {
		canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(updatePreference.MealId, jwtPayload.UserId, roles.CanForceMealPreferenceAndCooking, db)
		if err != nil {
			if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
				responses.GenericGroupDoesNotExistError(c.Writer)
				return
			}
			responses.GenericInternalServerError(c.Writer)
			return
		}
		if !canPerformAction {
			responses.GenericNotAllowedToPerformActionError(c.Writer)
			return
		}
	}

	if updatePreference.Preference != nil {
		err = ChangeOptInStatusMealInDB(updatePreference.UserId, updatePreference.MealId, *updatePreference.Preference, db)
		if err != nil {
			responses.GenericInternalServerError(c.Writer)
			return
		}
	}

	if updatePreference.IsCook != nil {
		if *updatePreference.IsCook {
			err = AddCookToMealInDB(updatePreference.UserId, updatePreference.MealId, db)
		} else {
			err = RemoveCookFromMealInDB(updatePreference.UserId, updatePreference.MealId, db)
		}

		if err != nil && !errors.Is(err, ErrUserWasntACook) {
			responses.GenericInternalServerError(c.Writer)
			return
		}
	}

	if !isSelfAction {
		//TODO: Send notification to user whose preference was changed
	}

	c.JSON(http.StatusOK, MealSuccess{Message: "Preference successfully updated"})
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newTitle.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
		return
	}

	err = UpdateMealTitleIdDB(newTitle.MealId, newTitle.NewTitle, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newType.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newNotes.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}

		responses.GenericInternalServerError(c.Writer)
		return

	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
		return
	}

	err = UpdateMealNotesInDB(newNotes.MealId, newNotes.NewNotes, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformActionViaMealId(newScheduledAt.MealId, jwtPayload.UserId, roles.CanUpdateMeal, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericNotAllowedToPerformActionError(c.Writer)
		return
	}

	err = UpdateMealScheduledAtInDB(newScheduledAt.MealId, newScheduledAt.NewScheduledAt, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	//TODO: Send an updated meal information to the frontend
	//TODO: Send a push notification to all not opt out or undecided users
	c.JSON(http.StatusOK, MealSuccess{Message: "Meal updated successfully"})
}

func SyncGroupMeals(c *gin.Context, db *sql.DB) {
	var requestSyncGroupMeals RequestSyncGroupMeals
	if err := c.ShouldBindQuery(&requestSyncGroupMeals); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	inGroup, err := group.IsUserInGroup(requestSyncGroupMeals.GroupId, jwtPayload.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !inGroup {
		responses.GenericNotFoundError(c.Writer)
		return
	}

	meals, err := GetAllMealsInGroupInTimeframe(requestSyncGroupMeals.GroupId, jwtPayload.UserId, requestSyncGroupMeals.StartDate, requestSyncGroupMeals.EndDate, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	deletedIds, err := GetDeletedMealsInTimeframe(requestSyncGroupMeals.GroupId, requestSyncGroupMeals.StartDate, requestSyncGroupMeals.EndDate, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, ResponseSyncGroupMeals{
		Meals:      meals,
		DeletedIds: deletedIds,
	})
}

func SyncMealInformation(c *gin.Context, db *sql.DB) {
	var mealInfo RequestMealId
	if err := c.ShouldBindQuery(&mealInfo); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	_, err = group.IsUserInGroupViaMealId(mealInfo.MealId, jwtPayload.UserId, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.GenericGroupDoesNotExistError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	mealInformation, err := GetSingularMealInformation(mealInfo.MealId, jwtPayload.UserId, db)
	if err != nil {
		if errors.Is(err, ErrNoData) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.MealDoesNotExistError, "Meal does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	participationInformation, err := GetMealParticipationInformationFromDB(mealInfo.MealId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}
	meal := ResponseSyncSingularMeal{
		MealInformation: mealInformation,
		MealPreferenceInformation: ResponsePreferenceSync{
			Preferences: participationInformation,
			DeletedIds:  nil,
		},
	}
	c.JSON(http.StatusOK, meal)
}
