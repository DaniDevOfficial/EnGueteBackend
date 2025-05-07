package frontendErrors

const (
	//Generic Errors
	UnauthorizedError   = "unauthorizedError"
	ForbiddenError      = "forbiddenError"
	NotFoundError       = "notFoundError"
	InternalServerError = "internalServerError"
	BadRequestError     = "badRequestError"

	NotAllowedToDeleteGroupError = "notAllowedToDeleteGroupError"
	NotAllowedToUpdateGroupError = "notAllowedToUpdateGroupError"
	GroupDoesNotExistError       = "groupDoesNotExistError"

	InvalidInviteTokenError = "invalidInviteTokenError"

	UserDoesNotExistError = "userDoesNotExistError"
	CreateGroupError      = "createGroupError"

	YouCantKickOrBanYourselfError = "youCantKickOrBanYourselfError"
	InvalidRoleError              = "invalidRoleError"

	PasswordFormatTooShortError               = "passwordFormatTooShortError"
	PasswordFormatNeedsUpperLowerSpecialError = "passwordFormatNeedsUpperLowerSpecialError"
	PasswordFormatTooLongError                = "passwordFormatTooLongError"
	PasswordDoesNotMatchError                 = "passwordDoesNotMatchError"

	UsernameIsAlreadyTakenError  = "usernameIsAlreadyTakenError"
	WrongUsernameOrPasswordError = "wrongUsernameOrPasswordError"

	MealDoesNotExistError = "mealDoesNotExistError"

	FiltersAreNotValidError = "filtersAreNotValidError"
)
