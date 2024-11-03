package user

import (
	"database/sql"
	"enguete/util/auth"
	"enguete/util/hashing"
	"enguete/util/jwt"
	"enguete/util/validation"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

// CreateNewUser @Summary Create a new user
// @Description Create a new user in the system with password and username validation
// @Tags users
// @Accept json
// @Produce json
// @Param user body user.RequestNewUser true "Request payload for creating a new user"
// @Success 201 {object} jwt.JWTTokenResponse
// @Failure 400 {object} user.UserError
// @Router /auth/signup [post]
func CreateNewUser(c *gin.Context, db *sql.DB) {

	var newUser RequestNewUser
	if err := c.ShouldBindJSON(&newUser); err != nil {
		log.Println(err)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding request"})
		return
	}
	isValid, err := validation.IsValidPassword(newUser.Password)
	if err != nil {
		userError := UserError{
			Error: "invalid Password Struct",
		}
		c.JSON(http.StatusBadRequest, userError)
		return
	}
	if !isValid {
		userError := UserError{
			Error: "invalid Password Struct",
		}
		c.JSON(http.StatusBadRequest, userError)
		return
	}

	userId, err := GetUserIdByName(newUser.Username, db)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for users"})
		return
	}

	if userId != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "User Does already exist"})
		return
	}

	hashedPassword, err := hashing.HashPassword(newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hashing error"})
		return
	}

	userInDB := DBNewUser{
		username:      newUser.Username,
		email:         newUser.Email,
		password_hash: hashedPassword,
	}

	newUserId, err := CreateUserInDB(userInDB, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Creating User"})
		return
	}
	jwtUserData := jwt.JWTUser{
		Username: userInDB.username,
		UserId:   newUserId,
	}
	jwtToken, err := jwt.CreateToken(jwtUserData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating JWT"})
		return
	}
	response := jwt.JWTTokenResponse{
		Token: jwtToken,
	}
	c.JSON(http.StatusCreated, response)
}

// SignIn @Summary Sign in to account
// @Description Sign in to a previously created account. Cheks for correct password and username
// @Tags users
// @Accept json
// @Produce json
// @Param user body user.RequestNewUser true "Request payload for sign in into an account"
// @Success 201 {object} jwt.JWTTokenResponse
// @Failure 400 {object} user.UserError
// @Router /auth/signin [post]
func SignIn(c *gin.Context, db *sql.DB) {

	var credentials SignInCredentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error decoding request"})
		return
	}

	userData, err := GetUserByName(credentials.Username, db)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong USERNAME Or Password"})
		return
	}

	if !hashing.CheckHashedString(userData.passwordHash, credentials.Password) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong Username Or Password"})
		return
	}

	jwtUserData := jwt.JWTUser{
		Username: userData.username,
		UserId:   userData.userId,
	}
	jwtToken, err := jwt.CreateToken(jwtUserData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error creating JWT"})
		return
	}

	response := jwt.JWTTokenResponse{
		Token: jwtToken,
	}
	c.JSON(http.StatusOK, response)
}

// GetUserById @Summary Get a user by his id
// @Description Get a user by his id or return an error
// @Tags users
// @Accept json
// @Produce json
// @Param user body string true "UUID"
// @Success 201 {object} user.ResponseUserData
// @Failure 400 {object} user.UserError
// @Failure 404 {object} user.UserError
// @Failure 500 {object} user.UserError
// @Router /users/:id [get]
func GetUserById(c *gin.Context, db *sql.DB) {

	userId := c.Param("id")
	userId = strings.Trim(userId, " ")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No userId attatched"})
		return
	}

	userData, err := GetUserByIdFromDB(userId, db)
	if errors.Is(err, sql.ErrNoRows) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "shit hit the fan"})
		return
	}

	response := ResponseUserData{
		Username: userData.username,
		UserID:   userData.userId,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserGroupsById @Summary Get a users groups his in, by his id
// @Description Get a users groups he's in by his id or return an error
// @Tags users
// @Accept json
// @Produce json
// @Param user body string true "UUID"
// @Success 201 {object} user.ResponseUserGroups
// @Failure 400 {object} user.UserError
// @Failure 404 {object} user.UserError
// @Failure 500 {object} user.UserError
// @Router /users/:id/groups [get]
func GetUserGroupsById(c *gin.Context, db *sql.DB) {

	userId := c.Param("id")
	userId = strings.Trim(userId, " ")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No userId attatched"})
		return
	}

	userGroups, err := GetUsersGroupByUserIdFromDB(userId, db)
	if errors.Is(err, sql.ErrNoRows) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "shit hit the fan"})
		return
	}

	response := ResponseUserGroups{
		Groups: userGroups,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUserWithJWT @Summary Delete a user
// @Description Delete a user with his jwt token
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Success 201 {object} user.UserSuccess
// @Failure 401 {object} user.UserError
// @Failure 500 {object} user.UserError
// @Router /users/:uuid [delete]
func DeleteUserWithJWT(c *gin.Context, db *sql.DB) {

	decodedJWT, err := auth.GetJWTPayloadFromHeader(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "JWT Token is not valid"})
		return
	}

	// TODO: Do a email for validation and then handle the delete in another function
	_, err = DeleteUserInDB(decodedJWT.UserId, db)
	if err != nil {
		errorMessage := UserError{
			Error: "user wasnt deleted",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}

	successResponse := UserSuccess{
		Message: "user deleted sucessfuly",
	}

	c.JSON(http.StatusOK, successResponse)
}

// UpdateUsername @Summary Update a users username
// @Description Update a users username with his jwt token
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Param user body user.RequestChangeUsername true "Request payload for changing username"
// @Success 201 {object} user.UserSuccess
// @Failure 400 {object} user.UserError
// @Router /users/username [post]
func UpdateUsername(c *gin.Context, db *sql.DB) {
	var changeUsernameData RequestChangeUsername

	if err := c.ShouldBindJSON(&changeUsernameData); err != nil {
		log.Println(err)
		errorMessage := UserError{
			Error: "Error decoding request",
		}
		c.JSON(http.StatusBadRequest, errorMessage)
		return
	}

	jwtTokenData, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		return
	}
	userId, err := GetUserIdByName(changeUsernameData.Username, db)

	if userId == "" || err != nil {
		errorMessage := UserError{
			Error: "Username is already in use",
		}
		c.JSON(http.StatusBadRequest, errorMessage)
		return
	}
	err = UpdateUsernameInDB(changeUsernameData.Username, jwtTokenData.UserId, db)
	if err != nil {
		errorMessage := UserError{
			Error: "we fucked up",
		}
		c.JSON(http.StatusInternalServerError, errorMessage)
		return
	}
	successMessage := UserSuccess{
		Message: "Username updated Successfully",
	}
	c.JSON(http.StatusOK, successMessage)
}

// UpdateUserPassword @Summary Update a users username
// @Description Update a users username with his jwt token
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Param user body user.RequestChangePassword true "Request payload for Changing password"
// @Success 201 {object} user.UserSuccess
// @Failure 400 {object} user.UserError
// @Router /users/password [post]
func UpdateUserPassword(c *gin.Context, db *sql.DB) {
	var updatePasswordData RequestChangePassword
	if err := c.ShouldBindJSON(&updatePasswordData); err != nil {
		log.Println(err)
		errorMessage := UserError{
			Error: "Error decoding request",
		}
		c.JSON(http.StatusBadRequest, errorMessage)
		return
	}

	isValid, err := validation.IsValidPassword(updatePasswordData.NewPassword)
	if err != nil {
		errorMessage := UserError{
			Error: "Password isnt valid",
		}
		c.JSON(http.StatusBadRequest, errorMessage)
		return
	}

	if !isValid {
		errorMessage := UserError{
			Error: "Password isnt valid",
		}
		c.JSON(http.StatusBadRequest, errorMessage)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c)
	if err != nil {
		errorMessage := UserError{
			Error: "JWT Token is not valid",
		}
		c.JSON(http.StatusUnauthorized, errorMessage)
		return
	}

	userData, err := GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		errorMessage := UserError{
			Error: "User not found",
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, errorMessage)
		return
	}
	if !hashing.CheckHashedString(userData.passwordHash, updatePasswordData.OldPassword) {

		errorMessage := UserError{
			Error: "Your Old password doesnt match",
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, errorMessage)
		return
	}
	hashedPassword, err := hashing.HashPassword(updatePasswordData.NewPassword)
	if err != nil {
		errorMessage := UserError{
			Error: "Error hashing password",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorMessage)
		return
	}
	err = UpdatePasswordInDb(hashedPassword, jwtPayload.UserId, db)
	if err != nil {
		return
	}

	successMessage := UserSuccess{
		Message: "Password updated Successfully",
	}
	c.JSON(http.StatusOK, successMessage)
}
