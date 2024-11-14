package management

import (
	"database/sql"
	"enguete/modules/group"
	"enguete/util/auth"
	"enguete/util/roles"
	"github.com/gin-gonic/gin"
	"net/http"
)

// KickUserFromGroup @Summary Kick a user from a group
// @Description Allows a user to kick another user within a group. The requesting user must be a member of the group and have the required rights to perform this action.
// @Tags meals
// @Accept json
// @Produce json
// @Param requestKickUser body RequestKickUser true "Payload Kick user from group"
// @Success 200 {object} ManagementSuccess "User successfully kicked from group"
// @Failure 400 {object} ManagementError "Invalid request body or user tries to kick himself"
// @Failure 401 {object} ManagementError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router /management/user/kick [post]
func KickUserFromGroup(c *gin.Context, db *sql.DB) {
	var kickUserData RequestKickUser
	if err := c.ShouldBind(&kickUserData); err != nil {
		c.JSON(http.StatusBadRequest, ManagementError{Error: "Error decoding request"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
	}
	if jwtPayload.UserId == kickUserData.UserId {
		c.JSON(http.StatusBadRequest, ManagementError{Error: "You can't kick yourself"})
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanKickUsers, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ManagementError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "You are not allowed to perform this action"})
	}

	err = KickUSerFromGroupInDB(kickUserData.GroupId, kickUserData.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
	}

	//TODO: send notification to kicked user
	//TODO: update userData for the frontend
	c.JSON(http.StatusOK, ManagementSuccess{Message: "user successfully kicked"})
}

// BanUserFromGroup @Summary Ban a user from a group
// @Description Allows a user to ban another user within a group, and add him to the banned List. The requesting user must be a member of the group and have the required rights to perform this action.
// @Tags meals
// @Accept json
// @Produce json
// @Param requestKickUser body RequestKickUser true "Payload to ban user from group"
// @Success 200 {object} ManagementSuccess "User successfully banned from group"
// @Failure 400 {object} ManagementError "Invalid request body or user tries to ban himself"
// @Failure 401 {object} ManagementError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router /management/user/kick [post]
func BanUserFromGroup(c *gin.Context, db *sql.DB) {
	var kickUserData RequestKickUser
	if err := c.ShouldBind(&kickUserData); err != nil {
		c.JSON(http.StatusBadRequest, ManagementError{Error: "Error decoding request"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
	}

	if jwtPayload.UserId == kickUserData.UserId {
		c.JSON(http.StatusBadRequest, ManagementError{Error: "You can't ban yourself"})
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanBanUsers, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ManagementError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "You are not allowed to perform this action"})
	}

	//TODO: either have a seperate function or a follow up, which adds the userId in a blacklist for this specific group
	err = KickUSerFromGroupInDB(kickUserData.GroupId, kickUserData.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
	}

	//TODO: send notification to kicked user
	//TODO: update userData for the frontend
	c.JSON(http.StatusOK, ManagementSuccess{Message: "user successfully kicked"})
}

// UnbanUserFromGroup  @Summary unban a user from a group
// @Description Allows a user to unban another user within a group, and remove him from the banned List. The requesting user must be a member of the group and have the required rights to perform this action.
// @Tags meals
// @Accept json
// @Produce json
// @Param requestKickUser body RequestKickUser true "Payload to unban user from group"
// @Success 200 {object} ManagementSuccess "User successfully unbanned from group"
// @Failure 400 {object} ManagementError "Invalid request body"
// @Failure 401 {object} ManagementError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} ManagementError "Internal server error"
// @Router /management/user/kick [post]
func UnbanUserFromGroup(c *gin.Context, db *sql.DB) {
	var kickUserData RequestKickUser
	if err := c.ShouldBind(&kickUserData); err != nil {
		c.JSON(http.StatusBadRequest, ManagementError{Error: "Error decoding request"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
	}

	canPerformAction, err := group.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanUnbanUser, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ManagementError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "You are not allowed to perform this action"})
	}

	err = UnBanUserFromGroupInDB(kickUserData.GroupId, kickUserData.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ManagementError{Error: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, ManagementSuccess{Message: "user successfully unbanned"})
}
