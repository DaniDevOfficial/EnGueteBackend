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
// @Success 201 {object} group.NewGroup
// @Failure 400 {object} group.GroupError
// @Failure 404 {object} group.GroupError
// @Failure 500 {object} group.GroupError
// @Router /groups [get]
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
	newGroupId, err := CreateNewGroupInDB(newGroupData, decodedJWT.UserId, db)
	if err != nil {
		errorMessage := GroupError{
			Error: "Error Creating Group",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	response := struct {
		GroupId string `json:"groupId"`
	}{
		GroupId: newGroupId,
	}
	c.JSON(http.StatusCreated, response)
}
