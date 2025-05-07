package management

import (
	"database/sql"
	"enguete/modules/group"
	"enguete/util/auth"
	"enguete/util/frontendErrors"
	"enguete/util/responses"
	"enguete/util/roles"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// KickUserFromGroup godoc
// @Summary Kick a user from a group
// @Description Allows a user to remove another user from a group. The requesting user must be a group member with the necessary permissions.
// @Tags Management
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param kickUserData body RequestKickUser true "Payload to kick user from group"
// @Success 200 {object} ManagementSuccess "User successfully kicked from group"
// @Failure 400 {object} ManagementError "Error decoding request" or "You can't kick yourself"
// @Failure 401 {object} ManagementError "Unauthorized" or "You are not allowed to perform this action"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router /management/user/kick [post]
func KickUserFromGroup(c *gin.Context, db *sql.DB) {
	var kickUserData RequestKickUser
	if err := c.ShouldBind(&kickUserData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}
	if jwtPayload.UserId == kickUserData.UserId {
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.YouCantKickOrBanYourselfError, "You can't kick yourself. You need To leave the group")
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanKickUsers, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusForbidden, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
		return
	}

	err = KickUSerFromGroupInDB(kickUserData.GroupId, kickUserData.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	//TODO: send notification to kicked user
	//TODO: update userData for the frontend
	c.JSON(http.StatusOK, ManagementSuccess{Message: "user successfully kicked"})
}

// BanUserFromGroup godoc
// @Summary Ban a user from a group
// @Description Allows a user to ban another user within a group, adding them to the banned list. The requesting user must be a group member with the appropriate permissions to perform this action.
// @Tags Management
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param requestKickUser body RequestKickUser true "Payload to ban user from group"
// @Success 200 {object} ManagementSuccess "User successfully banned from group"
// @Failure 400 {object} ManagementError "Invalid request body or user tries to ban themselves"
// @Failure 401 {object} ManagementError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router /management/user/ban [post]
func BanUserFromGroup(c *gin.Context, db *sql.DB) {
	var kickUserData RequestKickUser
	if err := c.ShouldBind(&kickUserData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	if jwtPayload.UserId == kickUserData.UserId {
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.YouCantKickOrBanYourselfError, "You can't kick/ban yourself. You need To leave the group")
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanBanUsers, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusForbidden, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
		return
	}

	//TODO: either have a seperate function or a follow up, which adds the userId in a blacklist for this specific group
	err = KickUSerFromGroupInDB(kickUserData.GroupId, kickUserData.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	//TODO: send notification to kicked user
	//TODO: update userData for the frontend
	c.JSON(http.StatusOK, ManagementSuccess{Message: "user successfully kicked"})
}

// UnbanUserFromGroup godoc
// @Summary Unban a user from a group
// @Description Allows a user to unban another user within a group, removing them from the banned list. The requesting user must be a group member with the appropriate permissions to perform this action.
// @Tags Management
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param requestKickUser body RequestKickUser true "Payload to unban user from group"
// @Success 200 {object} ManagementSuccess "User successfully unbanned from group"
// @Failure 400 {object} ManagementError "Invalid request body"
// @Failure 401 {object} ManagementError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router /management/user/unban [post]
func UnbanUserFromGroup(c *gin.Context, db *sql.DB) {
	var kickUserData RequestKickUser
	if err := c.ShouldBind(&kickUserData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanUnbanUser, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusForbidden, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
	}

	err = UnBanUserFromGroupInDB(kickUserData.GroupId, kickUserData.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, ManagementSuccess{Message: "user successfully unbanned"})
}

// AddRoleToUser godoc
// @Summary Add a role to a user in a specific group
// @Description This endpoint assigns a specified role to a user within a given group. Requires authorization and appropriate permissions.
// @Tags Roles
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param roleData body RequestRoleData true "Role data containing groupId, userId, and role"
// @Success 200 {object} ManagementSuccess "Role successfully added"
// @Failure 400 {object} ManagementError "Error decoding request" or "Invalid role"
// @Failure 401 {object} ManagementError "Unauthorized" or "You are not allowed to perform this action"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router management/roles/add [post]
func AddRoleToUser(c *gin.Context, db *sql.DB) {
	var roleData RequestRoleData
	if err := c.ShouldBindJSON(&roleData); err != nil {
		c.JSON(http.StatusBadRequest, ManagementError{Error: "Error decoding request"})
		return
	}

	role := roles.GetConstViaString(roleData.Role)
	if role == "" {
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.InvalidRoleError, "Invalid role")
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	action := "can_promote_to_" + role
	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformAction(roleData.GroupId, jwtPayload.UserId, action, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			c.JSON(http.StatusForbidden, ManagementError{Error: "Group does not exist"})
			return
		}
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
		return
	}

	err = group.AddRoleToUserInGroup(roleData.GroupId, roleData.UserId, role, db)
	if err != nil {
		if errors.Is(err, group.ErrNothingHappened) {
			c.JSON(http.StatusOK, ManagementSuccess{Message: "Role successfully added"})
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, ManagementSuccess{Message: "Role successfully added"})
}

// RemoveRoleFromUser godoc
// @Summary Remove a role from a user in a specific group
// @Description This endpoint removes a specified role from a user in a given group. Requires authorization and permissions.
// @Tags Roles
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param roleData body RequestRoleData true "Role data containing groupId, userId, and role"
// @Success 200 {object} ManagementSuccess "Role successfully removed"
// @Failure 400 {object} ManagementError "Error decoding request" or "Invalid role" or "User did not have this role"
// @Failure 401 {object} ManagementError "Unauthorized" or "You are not allowed to perform this action"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router management/roles/remove [delete]
func RemoveRoleFromUser(c *gin.Context, db *sql.DB) {
	var roleData RequestRoleData
	if err := c.ShouldBindJSON(&roleData); err != nil {
		responses.GenericBadRequestError(c.Writer)
		return
	}

	role := roles.GetConstViaString(roleData.Role)
	if role == "" {
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.InvalidRoleError, "Invalid role")
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	action := "can_demote_from_" + role
	canPerformAction, _, err := group.CheckIfUserIsAllowedToPerformAction(roleData.GroupId, jwtPayload.UserId, action, db)
	if err != nil {
		if errors.Is(err, group.ErrUserIsNotPartOfThisGroup) {
			responses.HttpErrorResponse(c.Writer, http.StatusForbidden, frontendErrors.GroupDoesNotExistError, "Group does not exist")
			return
		}

		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !canPerformAction {
		responses.GenericForbiddenError(c.Writer)
		return
	}

	err = group.RemoveRoleFromUserInGroup(roleData.GroupId, roleData.UserId, role, db)
	if err != nil {
		if errors.Is(err, group.ErrNothingHappened) {
			c.JSON(http.StatusOK, ManagementSuccess{Message: "Role successfully removed"})
			return
		}
		c.JSON(http.StatusInternalServerError, ManagementError{Error: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, ManagementSuccess{Message: "Role successfully removed"})
}
