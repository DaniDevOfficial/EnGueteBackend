package management

import (
	"database/sql"
	"enguete/modules/meal"
	"enguete/util/auth"
	"enguete/util/roles"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

	canPerformAction, err := meal.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanBanUsers, db)
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

	canPerformAction, err := meal.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanBanUsers, db)
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

	canPerformAction, err := meal.CheckIfUserIsAllowedToPerformAction(kickUserData.GroupId, jwtPayload.UserId, roles.CanUnbanUser, db)
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
