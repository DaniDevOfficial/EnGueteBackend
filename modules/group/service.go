package group

import (
	"database/sql"
	"enguete/modules/user"
	"enguete/util/auth"
	"enguete/util/dates"
	"enguete/util/frontendErrors"
	"enguete/util/responses"
	"enguete/util/roles"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// CreateNewGroup godoc
// @Summary Create a new group
// @Description Creates a new group and assigns the requester as the group creator with admin and member roles.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param group body RequestNewGroup true "Request payload for creating a new group"
// @Success 201 {object} ResponseGroupId "Group successfully created"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid authorization token"
// @Failure 404 {object} GroupError "Not Found - resource not found"
// @Failure 500 {object} GroupError "Internal server error - error during group creation or transaction handling"
// @Router /groups [post]
func CreateNewGroup(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Authorisation is not valid"})
		return
	}

	var newGroupData RequestNewGroup
	if err := c.ShouldBindJSON(&newGroupData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	newGroupId, err := CreateNewGroupInDBWithTransaction(newGroupData, jwtPayload.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		responses.HttpErrorResponse(c.Writer, http.StatusInternalServerError, frontendErrors.CreateGroupError, "Error creating group")
		return
	}

	userGroupId, err := AddUserToGroupWithTransaction(newGroupId, jwtPayload.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		responses.HttpErrorResponse(c.Writer, http.StatusInternalServerError, frontendErrors.CreateGroupError, "Error creating group")
		return
	}

	err = AddRoleToUserInGroupWithTransaction(newGroupId, jwtPayload.UserId, roles.AdminRole, userGroupId, tx)
	if err != nil {
		_ = tx.Rollback()
		responses.HttpErrorResponse(c.Writer, http.StatusInternalServerError, frontendErrors.CreateGroupError, "Error creating group")
		return
	}

	err = AddRoleToUserInGroupWithTransaction(newGroupId, jwtPayload.UserId, roles.MemberRole, userGroupId, tx)
	if err != nil {
		_ = tx.Rollback()
		responses.HttpErrorResponse(c.Writer, http.StatusInternalServerError, frontendErrors.CreateGroupError, "Error creating group")
		return
	}

	err = tx.Commit()
	if err != nil {
		responses.HttpErrorResponse(c.Writer, http.StatusInternalServerError, frontendErrors.CreateGroupError, "Error creating group")
		return
	}

	c.JSON(http.StatusCreated, ResponseGroupId{
		GroupId: newGroupId,
	})
}

func DeleteGroup(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var groupData RequestIdGroup
	if err := c.ShouldBindQuery(&groupData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	canPerformAction, _, err := CheckIfUserIsAllowedToPerformAction(groupData.GroupId, jwtPayload.UserId, roles.CanDeleteGroup, db)
	if err != nil {
		if errors.Is(err, ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.HttpErrorResponse(c.Writer, http.StatusForbidden, frontendErrors.NotAllowedToDeleteGroupError, "You are not allowed to delete this group")
		return
	}

	err = DeleteGroupInDB(groupData.GroupId, db)
	if err != nil {
		if errors.Is(err, ErrNothingHappened) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}

		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, GroupSuccess{Message: "Group deleted successfully"})
}

func UpdateGroupName(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var groupData RequestUpdateGroupName
	if err := c.ShouldBindJSON(&groupData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	canPerformAction, _, err := CheckIfUserIsAllowedToPerformAction(groupData.GroupId, jwtPayload.UserId, roles.CanUpdateGroup, db)
	if err != nil {
		if errors.Is(err, ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.HttpErrorResponse(c.Writer, http.StatusForbidden, frontendErrors.NotAllowedToUpdateGroupError, "You are not allowed to update this group")
		return
	}

	err = UpdateGroupNameInDB(groupData, tx)
	if err != nil {
		_ = tx.Rollback()

		if errors.Is(err, ErrNothingHappened) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}

		responses.GenericInternalServerError(c.Writer)
		return
	}

	err = tx.Commit()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, GroupSuccess{Message: "Group name updated successfully"})
}

// GetGroupById godoc
// @Summary Retrieve group information
// @Description Fetches detailed information about a specific group, including group metadata and associated meals.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param groupId path string true "Group ID to fetch information for"
// @Success 200 {object} Group "Group information retrieved successfully"
// @Failure 400 {object} GroupError "Bad request - invalid group ID or request format"
// @Failure 401 {object} GroupError "Unauthorized - invalid authorization token"
// @Failure 403 {object} GroupError "Forbidden - user is not a member of the group"
// @Failure 404 {object} GroupError "Not Found - group not found"
// @Failure 500 {object} GroupError "Internal server error - database error or failure in retrieving group data"
// @Router /groups/{groupId} [get]
func GetGroupById(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		log.Println(err)
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var filterRequest FilterGroupRequest
	if err = c.ShouldBindQuery(&filterRequest); err != nil {
		log.Println(err)
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.FiltersAreNotValidError, "Error decoding request")
		return
	}

	inDB, err := IsUserInGroup(filterRequest.GroupId, jwtPayload.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !inDB {
		c.JSON(http.StatusForbidden, GroupError{Error: "You are not in this group or it doesnt exist"})
		return
	}

	groupInformation, err := GetGroupInformationFromDb(filterRequest.GroupId, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}

	groupInformation.UserRoleRights = roles.GetAllAllowedActionsForRoles(groupInformation.UserRoles)

	if filterRequest.WeekFilter != nil {

		weekTime, err := time.Parse(time.RFC3339, *filterRequest.WeekFilter)
		if err != nil {
			log.Println(err)
			weekTime = time.Now()

		}

		startOfWeek, endOfWeek := dates.GetStartAndEndOfWeek(weekTime)

		filterRequest.StartDateFilter = &startOfWeek
		filterRequest.EndDateFilter = &endOfWeek
	}

	mealCards, err := GetMealsInGroupDB(filterRequest, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}

	response := Group{
		GroupInfo:  groupInformation,
		GroupMeals: mealCards,
	}
	c.JSON(http.StatusOK, response)
}

// GetGroupMembers godoc
// @Summary Retrieve group members
// @Description Fetches a list of members for a specific group, ensuring the user is authorized and a member of the group.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param groupId path string true "Group ID to fetch members for"
// @Success 200 {array} Member "List of group members retrieved successfully"
// @Failure 400 {object} GroupError "Bad request - invalid group ID or request format"
// @Failure 401 {object} GroupError "Unauthorized - invalid authorization token"
// @Failure 403 {object} GroupError "Forbidden - user is not a member of the group"
// @Failure 404 {object} GroupError "Not Found - group not found"
// @Failure 500 {object} GroupError "Internal server error - database error or failure in retrieving group members"
// @Router /groups/{groupId}/members [get]
func GetGroupMembers(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		log.Println(err)
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var groupData RequestIdGroup

	if err := c.ShouldBindQuery(&groupData); err != nil {
		log.Println(err)
		responses.GenericBadRequestError(c.Writer)
		return
	}

	inGroup, err := IsUserInGroup(groupData.GroupId, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)

		return
	}
	if !inGroup {
		responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "This group doesnt exist")
		return
	}

	members, err := GetGroupMembersFromDb(groupData.GroupId, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, members)
}

func GetGroupMeals(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		log.Println(err)
		responses.GenericUnauthorizedError(c.Writer)
		return
	}
	var groupData RequestGroupMeals
	if err := c.ShouldBindQuery(&groupData); err != nil {
		log.Println(err)
		responses.GenericBadRequestError(c.Writer)
		return
	}
	inGroup, err := IsUserInGroup(groupData.GroupId, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !inGroup {
		responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "This group doesnt exist")
		return
	}
	var filterRequest FilterGroupRequest

	filterRequest.GroupId = groupData.GroupId

	if groupData.FilterDate != nil {
		weekTime, err := time.Parse(time.RFC3339, *groupData.FilterDate)

		if err != nil {
			log.Println(err)
			weekTime = time.Now()
		}

		startOfWeek, endOfWeek := dates.GetStartAndEndOfWeek(weekTime)

		filterRequest.StartDateFilter = &startOfWeek
		filterRequest.EndDateFilter = &endOfWeek
	}

	mealCards, err := GetMealsInGroupDB(filterRequest, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}
	c.JSON(http.StatusOK, mealCards)
}

// GenerateInviteLink godoc
// @Summary Generate an invite link for a group
// @Description Generates a unique invite link for a specified group. Only users with admin or owner roles can generate an invitation link.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param inviteRequest body InviteLinkGenerationRequest true "Group ID and optional expiration settings for invite link generation"
// @Success 201 {object} InviteLinkGenerationResponse "Invite link successfully created"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid authorization or insufficient permissions"
// @Failure 500 {object} GroupError "Internal server error - issues with invite creation or transaction handling"
// @Router /groups/invite [post]
func GenerateInviteLink(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var inviteRequest InviteLinkGenerationRequest
	if err := c.ShouldBindJSON(&inviteRequest); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	canPerformAction, _, err := CheckIfUserIsAllowedToPerformAction(inviteRequest.GroupId, jwtPayload.UserId, roles.CanCreateInviteLinks, db)
	if err != nil {
		if errors.Is(err, ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "This group doesnt exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	token, err := CreateNewInviteInDBWithTransaction(inviteRequest, tx)
	if err != nil {
		_ = tx.Rollback()
		responses.GenericInternalServerError(c.Writer)
		return
	}

	err = tx.Commit()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusCreated, InviteToken{
		ExpiresAt:   inviteRequest.ExpirationDateTime,
		InviteToken: token,
	})
}

// JoinGroupWithInviteToken godoc
// @Summary Join a group using an invite token
// @Description Allows a user to join a specified group by validating an invite token. The user must have a valid token and necessary permissions.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param inviteToken path string true "Invite token for joining the group"
// @Success 200 {object} GroupSuccess "User successfully added to group"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid invite token or lack of permissions"
// @Failure 404 {object} GroupError "Not Found - user not found"
// @Failure 500 {object} GroupError "Internal server error - error adding user to group"
// @Router /groups/invite/join/{inviteToken} [post]
func JoinGroupWithInviteToken(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var inviteData RequestInviteToken
	if err := c.ShouldBindQuery(&inviteData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	groupId, err := ValidateInviteTokenInDB(inviteData.InviteToken, db)
	if err != nil {
		log.Println(err)
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.InvalidInviteTokenError, "Invalid invite token")
		return
	}

	_, err = user.GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.UserDoesNotExistError, "User not found")
		return
	}

	isMember, err := IsUserInGroup(groupId, jwtPayload.UserId, db)
	if err != nil {
		responses.HttpErrorResponse(c.Writer, http.StatusInternalServerError, frontendErrors.InternalServerError, "Internal server error")
		return
	}
	if isMember {
		// This is a happy path, the user is already in the group
		c.JSON(http.StatusOK, ResponseGroupId{GroupId: groupId})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal server error"})
		return
	}

	userGroupId, err := AddUserToGroupWithTransaction(groupId, jwtPayload.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		responses.GenericInternalServerError(c.Writer)
		return
	}

	err = AddRoleToUserInGroupWithTransaction(groupId, jwtPayload.UserId, roles.MemberRole, userGroupId, tx)
	if err != nil {
		_ = tx.Rollback()
		responses.GenericInternalServerError(c.Writer)
		return
	}

	err = tx.Commit()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	//TODO: this will maybe just return the groupId and then in the frontend the redirection will get handled
	c.JSON(http.StatusOK, ResponseGroupId{GroupId: groupId})
}

func GetAllInviteTokensInAGroup(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var groupData RequestIdGroup
	if err := c.ShouldBindQuery(&groupData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	canPerformAction, _, err := CheckIfUserIsAllowedToPerformAction(groupData.GroupId, jwtPayload.UserId, roles.CanViewInviteLinks, db)
	if err != nil {
		if errors.Is(err, ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
		return
	}

	inviteTokens, err := GetAllInviteTokensInAGroupFromDB(groupData.GroupId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, inviteTokens)
}

// VoidInviteToken godoc
// @Summary Delete an invite token
// @Description Allows a user to delete an invite token. The user must have a valid token and the necessary permissions.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param inviteToken path string true "Invite token to be deleted"
// @Success 200 {object} GroupSuccess "Successfully deleted the token"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid invite token or lack of permissions"
// @Failure 500 {object} GroupError "Internal server error - error deleting invite token"
// @Router /groups/invite/join/{inviteToken} [delete]
func VoidInviteToken(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var inviteData RequestInviteToken
	if err := c.ShouldBindQuery(&inviteData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	groupId, err := ValidateInviteTokenInDB(inviteData.InviteToken, db)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.InvalidInviteTokenError, "Invalid invite token")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	canPerformAction, _, err := CheckIfUserIsAllowedToPerformAction(groupId, jwtPayload.UserId, roles.CanVoidInviteLinks, db)
	if err != nil {
		if errors.Is(err, ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
		return
	}

	err = VoidInviteTokenInDB(inviteData.InviteToken, db)
	if err != nil {
		if errors.Is(err, ErrNothingHappened) {
			c.JSON(http.StatusOK, GroupSuccess{Message: "Invite token deleted"})
			return
		}

		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, GroupSuccess{Message: "Invite token deleted"})
}

// LeaveGroup godoc
// @Summary Leave a group
// @Description Allows a user to leave a specified group. If the user is the last member or the last admin, additional handling is performed (e.g., delete the group or assign a new admin).
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param groupId path string true "Group ID"
// @Success 200 {object} GroupSuccess "User successfully left the group"
// @Failure 400 {object} GroupError "User is not in the group or the group does not exist"
// @Failure 401 {object} GroupError "Authorization is not valid"
// @Failure 500 {object} GroupError "Error leaving group"
// @Router /groups/leave/{groupId} [delete]
func LeaveGroup(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	var groupData RequestIdGroup
	if err := c.ShouldBindQuery(&groupData); err != nil {

		responses.GenericBadRequestError(c.Writer)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	err = LeaveGroupInDB(groupData.GroupId, jwtPayload.UserId, tx) //TODO: some check for if a user was eiter the last user in a group or if there are no admins left. If he was the last one delete the group and if he was the last admin pick a new one by join-date
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, ErrNoMatchingGroupOrUser) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	err = RemovePreferencesInOpenMealsInGroup(jwtPayload.UserId, groupData.GroupId, tx)

	if err != nil {
		err = tx.Rollback()
		responses.GenericInternalServerError(c.Writer)
		return
	}
	//TODO: delete roles
	//TODO: delete isCook
	err = tx.Commit()
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, GroupSuccess{Message: "User left group"})
}
