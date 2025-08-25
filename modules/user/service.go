package user

import (
	"database/sql"
	"enguete/util/auth"
	"enguete/util/frontendErrors"
	"enguete/util/hashing"
	"enguete/util/jwt"
	"enguete/util/responses"
	"enguete/util/validation"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// SignUp godoc
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
func SignUp(c *gin.Context, db *sql.DB) {

	var newUser RequestNewUser
	if err := c.ShouldBindJSON(&newUser); err != nil {
		log.Println(err)
		responses.GenericBadRequestError(c.Writer)
		return
	}
	err := validation.IsValidPassword(newUser.Password)
	if err != nil {
		if errors.Is(err, validation.PasswordFormatNeedsUpperLowerSpecialError) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.PasswordFormatNeedsUpperLowerSpecialError, "Password needs Upper, lower and special characters and at least one number")
			return
		}
		if errors.Is(err, validation.PasswordFormatTooShortError) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.PasswordFormatTooShortError, "The Password needs to be at least 8 letters long")
			return
		}
		if errors.Is(err, validation.PasswordToLongError) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.PasswordFormatTooLongError, "The Password is to long, max length is 127 letters")
			return
		}
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	emailOrUsernameInUse, err := CheckIfUserExistsByEmailOrUsername(newUser.Email, newUser.Username, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if emailOrUsernameInUse {
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.UsernameOrEmailIsAlreadyTakenError, "Username is already taken")
		return
	}

	hashedPassword, err := hashing.HashPassword(newUser.Password)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericInternalServerError(c.Writer)
		return
	}
	jwtUserData := jwt.JWTUser{
		Username: userInDB.username,
		UserId:   newUserId,
	}

	refreshToken, err := jwt.CreateRefreshToken(jwtUserData, true, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}

	jwtToken, err := jwt.CreateToken(jwtUserData)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	userData, err := GetUserByName(credentials.Username, db)
	if err != nil {
		log.Println(err)
		if errors.Is(err, ErrUserNotFound) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.UserDoesNotExistError, "User does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	if !hashing.CheckHashedString(userData.PasswordHash, credentials.Password) {
		responses.HttpErrorResponse(c.Writer, http.StatusUnauthorized, frontendErrors.WrongUsernameOrPasswordError, "Wrong Username Or Password")
		return
	}

	jwtUserData := jwt.JWTUser{
		Username: userData.Username,
		UserId:   userData.UserId,
	}

	refreshToken, err := jwt.CreateRefreshToken(jwtUserData, false, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	jwtToken, err := jwt.CreateToken(jwtUserData)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}
	c.Header("Authorization", jwtToken)
	c.Header("RefreshToken", refreshToken)

	c.JSON(http.StatusOK, MessageResponse{Message: "Sign in successfully"})
}

func Logout(c *gin.Context, db *sql.DB) {
	refreshToken, err := auth.GetRefreshTokenFromHeader(c)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	err = jwt.VoidRefreshTokenInDB(refreshToken, db)
	if err != nil {
		log.Println(err)
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Logout successfully"})
}

func CheckAuth(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		log.Println(err)
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	userData, err := GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		log.Println(err)
		if errors.Is(err, ErrUserNotFound) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.UserDoesNotExistError, "User does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	userData, err := GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.UserDoesNotExistError, "User does not exist")
			return
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}

	groupData, err := GetUsersGroupByUserIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	response := ResponseUserData{
		Username: userData.Username,
		UserID:   userData.UserId,
		Groups:   groupData,
	}

	c.JSON(http.StatusOK, response)
}

func GetUserGroups(c *gin.Context, db *sql.DB) {
	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		log.Println(err)
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	groupData, err := GetUsersGroupByUserIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}
	c.JSON(http.StatusOK, groupData)
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
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	// TODO: Do a email for validation and then handle the delete in another function
	_, err = DeleteUserInDB(decodedJWT.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	jwtTokenData, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}
	userId, err := GetUserIdByName(changeUsernameData.Username, db)

	if userId != "" || err != nil {
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.UsernameIsAlreadyTakenError, "Username is already taken")
		return
	}
	err = UpdateUsernameInDB(changeUsernameData.Username, jwtTokenData.UserId, db)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			responses.GenericNotFoundError(c.Writer)
			return
		}
		responses.GenericInternalServerError(c.Writer)
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
		responses.GenericBadRequestError(c.Writer)
		return
	}

	err := validation.IsValidPassword(updatePasswordData.NewPassword)
	if err != nil {
		if errors.Is(err, validation.PasswordFormatNeedsUpperLowerSpecialError) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.PasswordFormatNeedsUpperLowerSpecialError, "Password needs Upper, lower and special characters and at least one number")
			return
		}
		if errors.Is(err, validation.PasswordFormatTooShortError) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.PasswordFormatTooShortError, "The Password needs to be at least 8 letters long")
			return
		}
		if errors.Is(err, validation.PasswordToLongError) {
			responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.PasswordFormatTooLongError, "The Password is to long, max length is 127 letters")
			return
		}
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	jwtPayload, err := auth.GetJWTPayloadFromHeader(c, db)
	if err != nil {
		responses.GenericUnauthorizedError(c.Writer)
		return
	}

	userData, err := GetUserByIdFromDB(jwtPayload.UserId, db)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			responses.HttpErrorResponse(c.Writer, http.StatusNotFound, frontendErrors.UserDoesNotExistError, "User does not exist")
		}
		responses.GenericInternalServerError(c.Writer)
		return
	}
	if !hashing.CheckHashedString(userData.PasswordHash, updatePasswordData.OldPassword) {
		responses.HttpErrorResponse(c.Writer, http.StatusBadRequest, frontendErrors.PasswordDoesNotMatchError, "Wrong Password")
		return
	}

	hashedPassword, err := hashing.HashPassword(updatePasswordData.NewPassword)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	err = UpdatePasswordInDb(hashedPassword, jwtPayload.UserId, db)
	if err != nil {
		responses.GenericInternalServerError(c.Writer)
		return
	}

	c.JSON(http.StatusOK, UserSuccess{Message: "Password updated Successfully"})
}
