package management

import (
	"database/sql"
	"enguete/modules/group"
	"enguete/util/auth"
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

	// TODO: some sort of smarter role check, than just a db query to check if user has this role
	err = group.CheckIfUserIsAdminOrOwnerOfGroupInDB(kickUserData.GroupId, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
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

	// TODO: some sort of smarter role check, than just a db query to check if user has this role
	err = group.CheckIfUserIsAdminOrOwnerOfGroupInDB(kickUserData.GroupId, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
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

	// TODO: some sort of smarter role check, than just a db query to check if user has this role
	err = group.CheckIfUserIsAdminOrOwnerOfGroupInDB(kickUserData.GroupId, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ManagementError{Error: "Unauthorized"})
		return
	}
	err = UnBanUserFromGroupInDB(kickUserData.GroupId, kickUserData.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ManagementError{Error: "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, ManagementSuccess{Message: "user successfully unbanned"})
}
