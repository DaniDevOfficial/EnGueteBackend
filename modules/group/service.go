package group

import (
	"database/sql"
	"enguete/modules/user"
	"enguete/util/auth"
	"enguete/util/roles"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// CreateNewGroup godoc
// @Summary Create a new group
// @Description Creates a new group and assigns the requester as the group creator with admin and member roles.
// @Tags Groups
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization"
// @Param group body RequestNewGroup true "Request payload for creating a new group"
// @Success 201 {object} ResponseNewGroup "Group successfully created"
// @Failure 400 {object} GroupError "Bad request - error decoding request"
// @Failure 401 {object} GroupError "Unauthorized - invalid authorization token"
// @Failure 404 {object} GroupError "Not Found - resource not found"
// @Failure 500 {object} GroupError "Internal server error - error during group creation or transaction handling"
// @Router /groups [post]
func CreateNewGroup(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Authorisation is not valid"})
		return
	}

	var newGroupData RequestNewGroup
	if err := c.ShouldBindJSON(&newGroupData); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, GroupError{Error: "Error decoding request"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal server error"})
	}

	newGroupId, err := CreateNewGroupInDBWithTransaction(newGroupData, jwtPayload.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Creating Group"})
		return
	}

	userGroupId, err := AddUserToGroupWithTransaction(newGroupId, jwtPayload.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Adding User to Group"})
		return
	}

	err = AddRoleToUserInGroupWithTransaction(newGroupId, jwtPayload.UserId, roles.AdminRole, userGroupId, tx)
	if err != nil {
		_ = tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Adding User to Group1"})
		return
	}

	err = AddRoleToUserInGroupWithTransaction(newGroupId, jwtPayload.UserId, roles.MemberRole, userGroupId, tx)
	if err != nil {
		_ = tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Adding User to Group2"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Adding User to Group"})
		return
	}

	c.JSON(http.StatusCreated, ResponseNewGroup{
		GroupId: newGroupId,
	})
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
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Authorization is not valid"})
		return
	}

	groupId := c.Param("groupId")

	inDB, err := IsUserInGroup(groupId, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal Server error"})
		return
	}
	if !inDB {
		c.JSON(http.StatusForbidden, GroupError{Error: "You are not in this group or it doesnt exist"})
		return
	}

	groupInformation, err := GetGroupInformationFromDb(groupId, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal Server error"})
		return
	}

	groupInformation.UserRoleRights = roles.GetAllAllowedActionsForRoles(groupInformation.UserRoleRights)
	mealCards, err := GetMealsInGroupDB(groupId, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal Server Error"})
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
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Authorization is not valid"})
		return
	}

	groupId := c.Param("groupId")
	inGroup, err := IsUserInGroup(groupId, jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal Server error"})
		return
	}
	if !inGroup {
		c.JSON(http.StatusForbidden, GroupError{Error: "You are not in this group or it doesnt exist"})
		return
	}

	members, err := GetGroupMembersFromDb(groupId, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal Server error"})
		return
	}

	c.JSON(http.StatusOK, members)
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
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Authorisation is not valid"})
		return
	}
	var inviteRequest InviteLinkGenerationRequest
	if err := c.ShouldBindJSON(&inviteRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, GroupError{Error: "Error decoding request"})
		return
	}
	canPerformAction, _, err := CheckIfUserIsAllowedToPerformAction(inviteRequest.GroupId, jwtPayload.UserId, roles.CanCreateInviteLinks, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusForbidden, GroupError{Error: "You are not allowed to perform this action"})
	}
	log.Println(4)

	tx, err := db.Begin()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Internal server error"})
		return
	}

	token, err := CreateNewInviteInDBWithTransaction(inviteRequest.GroupId, tx)
	if err != nil {
		log.Println(err)
		_ = tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Creating Invite1"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Creating Invite2"})
		return
	}

	fullLink := auth.GenerateInviteLink(token)
	c.JSON(http.StatusCreated, InviteLinkGenerationResponse{
		InviteLink: fullLink,
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
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Invalid jwt token"})
		return
	}

	inviteToken := c.Param("inviteToken")
	groupId, err := ValidateInviteTokenInDB(inviteToken, db)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Invalid invite token"})
		return
	}

	_, err = user.GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusNotFound, GroupError{Error: "User not found"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal server error"})
	}

	userGroupId, err := AddUserToGroupWithTransaction(groupId, jwtPayload.UserId, tx)
	if err != nil {
		_ = tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Adding User to Group"})
		return
	}

	err = AddRoleToUserInGroupWithTransaction(groupId, jwtPayload.UserId, roles.MemberRole, userGroupId, tx)
	if err != nil {
		_ = tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Adding User to Group"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error Adding User to Group"})
		return
	}

	//TODO: this will maybe just return the groupId and then in the frontend the redirection will get handled
	c.JSON(http.StatusOK, GroupSuccess{Message: "User added to group"})
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
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Invalid jwt token"})
		return
	}

	inviteToken := c.Param("inviteToken")
	groupId, err := ValidateInviteTokenInDB(inviteToken, db)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error validating invite token"})
		return
	}

	canPerformAction, _, err := CheckIfUserIsAllowedToPerformAction(groupId, jwtPayload.UserId, roles.CanVoidInviteLinks, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GroupError{Error: "Internal server error"})
		return
	}
	if !canPerformAction {
		c.JSON(http.StatusForbidden, GroupError{Error: "You are not allowed to perform this action"})
	}

	err = VoidInviteTokenIfAllowedInDB(groupId, jwtPayload.UserId, db)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error deleting invite token"})
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
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, GroupError{Error: "Authorisation is not valid"})
		return
	}

	groupId := c.Param("groupId")
	err = LeaveGroupInDB(groupId, jwtPayload.UserId, db) //TODO: some check for if a user was eiter the last user in a group or if there are no admins left. If he was the last one delete the group and if he was the last admin pick a new one by join-date
	if err != nil {
		if errors.Is(err, ErrNoMatchingGroupOrUser) {
			c.AbortWithStatusJSON(http.StatusBadRequest, GroupError{Error: "Cant leave a group your not in or that doesnt exist"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, GroupError{Error: "Error leaving group"})
		return
	}

	c.JSON(http.StatusOK, GroupSuccess{Message: "User left group"})
}
