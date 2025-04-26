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

// CreateNewUser godoc
// @Summary Create a new user
// @Description Create a new user with password and username validation.
// @Tags users
// @Accept json
// @Produce json
// @Param user body RequestNewUser true "Request payload for creating a new user"
// @Success 201 {object} jwt.JWTTokenResponse
// @Failure 400 {object} UserError "Invalid request data or username already exists"
// @Failure 500 {object} UserError "Server error creating user"
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

	refreshToken, err := jwt.CreateRefreshToken(jwtUserData, false, db)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error creating RefreshToken"})
		return
	}

	jwtToken, err := jwt.CreateToken(jwtUserData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error creating JWT"})
		return
	}

	c.Header("Authorization", jwtToken)
	c.Header("RefreshToken", refreshToken)

	c.JSON(http.StatusOK, MessageResponse{Message: "Sign up successfully"})
}

// SignIn godoc
// @Summary Sign in to an account
// @Description Sign in to an account. Checks for valid username and password.
// @Tags users
// @Accept json
// @Produce json
// @Param user body SignInCredentials true "Sign-in credentials"
// @Success 200 {object} jwt.JWTTokenResponse
// @Failure 401 {object} UserError "Invalid username or password"
// @Failure 500 {object} UserError "Server error during sign-in"
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong USERNAME Or Password"})
		return
	}

	if !hashing.CheckHashedString(userData.PasswordHash, credentials.Password) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong Username Or Password"})
		return
	}

	jwtUserData := jwt.JWTUser{
		Username: userData.Username,
		UserId:   userData.UserId,
	}

	refreshToken, err := jwt.CreateRefreshToken(jwtUserData, false, db)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error creating RefreshToken"})
		return
	}

	jwtToken, err := jwt.CreateToken(jwtUserData)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error creating JWT"})
		return
	}
	c.Header("Authorization", jwtToken)
	c.Header("RefreshToken", refreshToken)

	c.JSON(http.StatusOK, MessageResponse{Message: "Sign in successfully"})
}

func Logout(c *gin.Context, db *sql.DB) {
	refreshToken, err := auth.GetRefreshTokenFromHeader(c)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, UserError{Error: "Error getting JWT Payload"})
		return
	}

	err = jwt.VoidRefreshTokenInDB(refreshToken, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, UserError{Error: "Error logging out"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Logout successfully"})
}

func CheckAuth(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, UserError{Error: "Error getting JWT Payload"})
		return
	}

	userData, err := GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, UserError{Error: "Error getting JWT Payload"})
		return
	}

	response := ResponseUserData{
		UserID:   userData.UserId,
		Username: userData.Username,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserInformationById godoc
// @Summary Get a user by ID
// @Description Fetch user details by ID.
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Success 200 {object} ResponseUserData
// @Failure 400 {object} UserError "Invalid user ID"
// @Failure 404 {object} UserError "User not found"
// @Failure 500 {object} UserError "Server error retrieving user"
// @Router /users/{userId} [get]
func GetUserInformationById(c *gin.Context, db *sql.DB) {

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, UserError{Error: "Error getting JWT Payload"})
		return
	}

	userData, err := GetUserByIdFromDB(jwtPayload.UserId, db)
	if errors.Is(err, sql.ErrNoRows) {
		c.AbortWithStatusJSON(http.StatusNotFound, UserError{Error: "user not found"})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, UserError{Error: "internal server error"})
		return
	}

	groupData, err := GetUsersGroupByUserIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserError{Error: "internal server error"})
	}

	response := ResponseUserData{
		Username: userData.Username,
		UserID:   userData.UserId,
		Groups:   groupData,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserGroupsById godoc
// @Summary Get a user's groups by ID
// @Description Fetch all groups that a user belongs to by their ID.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} ResponseUserGroups
// @Failure 400 {object} UserError "Invalid user ID"
// @Failure 404 {object} UserError "User not found"
// @Failure 500 {object} UserError "Server error retrieving user groups"
// @Router /users/{userId}/groups [get]
func GetUserGroupsById(c *gin.Context, db *sql.DB) {

	userId := c.Param("userId")
	userId = strings.Trim(userId, " ")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, UserError{Error: "No userId attached"})
		return
	}

	userGroups, err := GetUsersGroupByUserIdFromDB(userId, db)
	if errors.Is(err, sql.ErrNoRows) {
		c.AbortWithStatusJSON(http.StatusNotFound, UserError{Error: "user not found"})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, UserError{Error: "shit hit the fan"})
		return
	}

	response := ResponseUserGroups{
		Groups: userGroups,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUserWithJWT godoc
// @Summary Delete a user
// @Description Delete a user based on JWT token.
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Success 200 {object} UserSuccess "User successfully deleted"
// @Failure 401 {object} UserError "Invalid JWT token"
// @Failure 500 {object} UserError "Server error deleting user"
// @Router /users [delete]
func DeleteUserWithJWT(c *gin.Context, db *sql.DB) {

	decodedJWT, err := auth.GetJWTPayloadFromHeader(c, db)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, UserError{Error: "JWT Token is not valid"})
		return
	}

	// TODO: Do a email for validation and then handle the delete in another function
	_, err = DeleteUserInDB(decodedJWT.UserId, db)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, UserError{Error: "user wasn't deleted"})
		return
	}

	successResponse := UserSuccess{
		Message: "user deleted successfully",
	}

	c.JSON(http.StatusOK, successResponse)
}

// UpdateUsername godoc
// @Summary Update a user's username
// @Description Update a user's username using their JWT token.
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Param user body RequestChangeUsername true "Username update payload"
// @Success 200 {object} UserSuccess "Username updated successfully"
// @Failure 400 {object} UserError "Invalid username or already in use"
// @Failure 500 {object} UserError "Server error updating username"
// @Router /users/username [put]
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

	jwtTokenData, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		return
	}
	userId, err := GetUserIdByName(changeUsernameData.Username, db)

	if userId != "" || err != nil {
		c.JSON(http.StatusBadRequest, UserError{Error: "Username is already in use"})
		return
	}
	err = UpdateUsernameInDB(changeUsernameData.Username, jwtTokenData.UserId, db)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, UserError{Error: "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, UserError{Error: "Internal server error"})
		return
	}
	successMessage := UserSuccess{
		Message: "Username updated Successfully",
	}
	c.JSON(http.StatusOK, successMessage)
}

// UpdateUserPassword godoc
// @Summary Update a user's password
// @Description Update a user's password using their JWT token.
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT Token"
// @Param user body RequestChangePassword true "Password update payload"
// @Success 200 {object} UserSuccess "Password updated successfully"
// @Failure 400 {object} UserError "Invalid password"
// @Failure 500 {object} UserError "Server error updating password"
// @Router /users/password [put]
func UpdateUserPassword(c *gin.Context, db *sql.DB) {
	var updatePasswordData RequestChangePassword
	if err := c.ShouldBindJSON(&updatePasswordData); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, UserError{Error: "Error decoding request"})
		return
	}

	isValid, err := validation.IsValidPassword(updatePasswordData.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, UserError{Error: "Password isn't valid"})
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, UserError{Error: "Password isnt valid"})
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, UserError{Error: "JWT Token is not valid"})
		return
	}

	userData, err := GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, UserError{Error: "User not found"})
		return
	}
	if !hashing.CheckHashedString(userData.PasswordHash, updatePasswordData.OldPassword) {

		c.AbortWithStatusJSON(http.StatusBadRequest, UserError{Error: "Your Old password doesnt match"})
		return
	}

	hashedPassword, err := hashing.HashPassword(updatePasswordData.NewPassword)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, UserError{Error: "Error hashing password"})
		return
	}

	err = UpdatePasswordInDb(hashedPassword, jwtPayload.UserId, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UserError{Error: "Error updating password"})
		return
	}

	successMessage := UserSuccess{
		Message: "Password updated Successfully",
	}
	c.JSON(http.StatusOK, successMessage)
}
