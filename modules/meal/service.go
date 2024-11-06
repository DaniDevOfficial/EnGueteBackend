package meal

import (
	"database/sql"
	"enguete/modules/group"
	"enguete/util/auth"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// CreateNewMeal @Summary Create a new meal
// @Description Creates a new meal within a specified group. The requesting user must be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param newMeal body RequestNewMeal true "Payload to create a new meal"
// @Success 201 {object} ResponseNewMeal "Successfully created new meal with meal ID"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/create [post]
func CreateNewMeal(c *gin.Context, db *sql.DB) {
	var newMeal RequestNewMeal
	err := c.ShouldBindJSON(&newMeal)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	err = group.CheckIfUserIsAdminOrOwnerOfGroupInDB(newMeal.GroupId, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}

	mealId, err := CreateNewMealInDB(newMeal, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	log.Println("New meal created with id:", mealId)
	c.JSON(http.StatusCreated, ResponseNewMeal{MealId: mealId})
}

// AddCookToMeal @Summary Add a cook to a meal
// @Description Adds a user as a cook to a specific meal within a group. Requires the user to be an admin or owner of the group.
// @Tags meals
// @Accept json
// @Produce json
// @Param addCookToMealData body RequestAddCookToMeal true "Payload to add a cook to a meal"
// @Success 201 {object} MealSuccess "Cook successfully added to meal"
// @Failure 400 {object} MealError "Invalid request body"
// @Failure 401 {object} MealError "Unauthorized user or insufficient permissions"
// @Failure 500 {object} MealError "Internal server error"
// @Router /meals/cook [post]
func AddCookToMeal(c *gin.Context, db *sql.DB) {
	var addCookToMealData RequestAddCookToMeal

	if err := c.ShouldBindJSON(&addCookToMealData); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	err = group.CheckIfUserIsAdminOrOwnerOfGroupInDB(addCookToMealData.GroupId, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	err = AddCookToMealInDB(addCookToMealData.UserId, addCookToMealData.UserId, db)

	if err != nil {
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	//TODO: Send notification to user who got added as a cook
	//TODO: Send an updated list of users in the meal
	c.JSON(http.StatusCreated, MealSuccess{Message: "Cook added to meal"})
}

func RemoveCookFromMeal(c *gin.Context, db *sql.DB) {
	var removeCookFromMealData RequestRemoveCook
	if err := c.ShouldBindJSON(&removeCookFromMealData); err != nil {
		c.JSON(http.StatusBadRequest, MealError{Error: "Invalid request body"})
		return
	}
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	err = group.CheckIfUserIsAdminOrOwnerOfGroupInDB(removeCookFromMealData.GroupId, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
		return
	}
	err = RemoveCookFromMealInDB(removeCookFromMealData.UserId, removeCookFromMealData.GroupId, db)
	if err != nil {
		if errors.Is(err, ErrUserWasntACook) {
			c.JSON(http.StatusUnauthorized, MealError{Error: "Unauthorized"})
			return
		}
		c.JSON(http.StatusInternalServerError, MealError{Error: "Internal server error"})
		return
	}
	//TODO: Send notification to user who got added as a cook
	//TODO: Send an updated list of users in the meal
	c.JSON(http.StatusOK, MealSuccess{Message: "Cook removed from meal"})
}
