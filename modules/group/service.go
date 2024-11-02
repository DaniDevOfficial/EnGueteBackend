package group

import (
	"database/sql"
	"enguete/util/auth"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// CreateNewGroup @Summary Create a new group
// @Description Create a new Group and put the requesters id as the creators id
// @Tags groups
// @Accept json
// @Produce json
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

func GetGroupById(c *gin.Context, db *sql.DB) {
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

	token, err := CreateNewInviteInDBWithTransaction(inviteRequest.GroupId, decodedJWT.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		errorMessage := GroupError{
			Error: "Error Creating Invite",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	err = tx.Commit()
	if err != nil {
		errorMessage := GroupError{
			Error: "Error Creating Invite",
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
