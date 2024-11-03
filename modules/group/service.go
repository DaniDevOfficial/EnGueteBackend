package group

import (
	"database/sql"
	"enguete/modules/user"
	"enguete/util/auth"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// CreateNewGroup @Summary Create a new group
// @Description Create a new Group and put the requesters id as the creators id
// @Tags groups
// @Accept json
// @Produce json
// @Param group body group.RequestNewGroup true "Request payload for a new group"
// @Success 201 {object} group.ResponseNewGroup
// @Failure 400 {object} group.GroupError
// @Failure 404 {object} group.GroupError
// @Failure 500 {object} group.GroupError
// @Router /groups [post]
func CreateNewGroup(c *gin.Context, db *sql.DB) {
	decodedJWT, err := auth.GetJWTPayloadFromHeader(c)

	if err != nil {
		errorMessage := GroupError{
			Error: "Authorisation is not valid",
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
		return
	}

	var newGroupData RequestNewGroup
	if err := c.ShouldBindJSON(&newGroupData); err != nil {
		log.Println(err)
		errorMessage := GroupError{
			Error: "Error decoding request",
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, errorMessage)
		return
	}
	tx, err := db.Begin()
	newGroupId, err := CreateNewGroupInDBWithTransaction(newGroupData, decodedJWT.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		errorMessage := GroupError{
			Error: "Error Creating Group",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	err = AddUserToGroupWithTransaction(newGroupId, decodedJWT.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		errorMessage := GroupError{
			Error: "Error Adding User to Group",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	err = tx.Commit()
	if err != nil {
		errorMessage := GroupError{
			Error: "Error Adding User to Group",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	response := ResponseNewGroup{
		GroupId: newGroupId,
	}
	c.JSON(http.StatusCreated, response)
}

func GetGroupById(c *gin.Context, db *sql.DB) { // TODO: this will be implemented later
	decodedJWT, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		errorMessage := GroupError{
			Error: "Authorisation is not valid",
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
	}
	groupId := c.Param("groupId")
	// DO this later
	log.Println(groupId)
	log.Println(decodedJWT.UserId)
}

// GenerateInviteLink @Summary Generate an invite link for a group
// @Description Generates a unique invite link for a specified group. Only users with admin or owner roles can generate an invitation link.
// @Tags groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param inviteRequest body InviteLinkGenerationRequest true "Group ID and optional expiration settings for invite link generation"
// @Success 201 {object} InviteLinkGenerationResponse
// @Failure 400 {object} group.GroupError "Bad request - error decoding request"
// @Failure 401 {object} group.GroupError "Unauthorized - authorization is not valid or user lacks permissions"
// @Failure 500 {object} group.GroupError "Internal server error - issues with invite creation or transaction handling"
// @Router /groups/invite [post]
func GenerateInviteLink(c *gin.Context, db *sql.DB) {
	decodedJWT, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		errorMessage := GroupError{
			Error: "Authorisation is not valid",
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
		return
	}

	var inviteRequest InviteLinkGenerationRequest

	if err := c.ShouldBindJSON(&inviteRequest); err != nil {
		errorMessage := GroupError{
			Error: "Error decoding request",
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, errorMessage)
		return
	}

	err = CheckIfUserIsAdminOrOwnerOfGroupInDB(inviteRequest.GroupId, decodedJWT.UserId, db)
	if err != nil {
		if err == sql.ErrNoRows {
			errorMessage := GroupError{
				Error: "You are not allowed to do this",
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
			return
		}
		log.Println(err)
		errorMessage := GroupError{
			Error: "Error while checkin authentication",
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		errorMessage := GroupError{
			Error: "Internal server error",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}

	token, err := CreateNewInviteInDBWithTransaction(inviteRequest.GroupId, tx)
	if err != nil {
		log.Println(err)
		_ = tx.Rollback()
		errorMessage := GroupError{
			Error: "Error Creating Invite1",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	err = tx.Commit()
	if err != nil {
		errorMessage := GroupError{
			Error: "Error Creating Invite2",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	fullLink := auth.GenerateInviteLink(token)
	inviteLinkResponse := InviteLinkGenerationResponse{
		InviteLink: fullLink,
	}
	c.JSON(http.StatusCreated, inviteLinkResponse)
}

// JoinGroupWithInviteToken handles joining a group via an invite token.
// @Summary Join a group using an invite token
// @Description Allows a user to join a specified group by validating an invite token. The user must have a valid token and necessary permissions.
// @Tags groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param inviteToken path string true "Invite token for joining the group"
// @Success 200 {object} GroupSuccess "User successfully added to group"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid invite token or lack of permissions"
// @Failure 404 {object} GroupError "Not Found - user not found"
// @Failure 500 {object} GroupError "Internal server error - error adding user to group"
// @Router /groups/invite/join/:inviteToken [post]
func JoinGroupWithInviteToken(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		errorMessage := GroupError{
			Error: "Invalid jwt token",
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
		return
	}
	inviteToken := c.Param("inviteToken")
	log.Println(inviteToken)
	groupId, err := ValidateInviteTokenInDB(inviteToken, db)
	if err != nil {
		log.Println(err)
		errorMessage := GroupError{
			Error: "Invalid invite token",
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
		return
	}

	_, err = user.GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		errorMessage := GroupError{
			Error: "User not found",
		}
		c.AbortWithStatusJSON(http.StatusNotFound, errorMessage)
		return
	}

	result, err := AddUserToGroupInDB(groupId, jwtPayload.UserId, db)
	if !result {
		errorMessage := GroupError{
			Error: "User already in group",
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, errorMessage)
		return
	}
	if err != nil {
		errorMessage := GroupError{
			Error: "Error adding user to group",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	response := GroupSuccess{
		Message: "User added to group",
	}
	c.JSON(http.StatusOK, response)
}

// DeleteInviteToken handles deleting an invite token.
// @Summary Delete an invite token.
// @Description Allows a user to delete an invite token. The user must have a valid token and necessary permissions.
// @Tags groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param inviteToken path string true "Invite token for deleting it"
// @Success 200 {object} GroupSuccess "Successfully deleted the token"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid invite token or lack of permissions"
// @Failure 500 {object} GroupError "Internal server error - error deleting invite token"
// @Router /groups/invite/join/:inviteToken [post]
func DeleteInviteToken(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		errorMessage := GroupError{
			Error: "Invalid jwt token",
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
		return
	}
	inviteToken := c.Param("inviteToken")
	groupId, err := ValidateInviteTokenInDB(inviteToken, db)
	if err != nil {
		errorMessage := GroupError{
			Error: "Error validating invite token",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	err = DeleteInviteTokenIfAllowedInDB(groupId, jwtPayload.UserId, db)
	if err != nil {
		errorMessage := GroupError{
			Error: "Error deleting invite token",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	response := GroupSuccess{
		Message: "Invite token deleted",
	}
	c.JSON(http.StatusOK, response)
}

// LeaveGroup handles a user leaving a group.
// @Summary Leave a group.
// @Description Allows a user to leave a group. The user must have a valid token and necessary permissions.
// @Tags groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param groupId path string true "Group ID for leaving the group"
// @Success 200 {object} GroupSuccess "User successfully left the group"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid group id or lack of permissions"
// @Failure 500 {object} GroupError "Internal server error - error leaving group"
// @Router /groups/leave/{groupId} [delete]
func LeaveGroup(c *gin.Context, db *sql.DB) {
	decodedJWT, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		errorMessage := GroupError{
			Error: "Authorisation is not valid",
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage)
		return
	}
	groupId := c.Param("groupId")
	err = LeaveGroupInDB(groupId, decodedJWT.UserId, db)
	if err != nil {
		if errors.Is(err, ErrNoMatchingGroupOrUser) {
			errorMessage := GroupError{
				Error: "Cant leave a group your not in or that doesnt exist",
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, errorMessage)
			return
		}

		errorMessage := GroupError{
			Error: "Error leaving group",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	response := GroupSuccess{
		Message: "User left group",
	}
	c.JSON(http.StatusOK, response)
}
