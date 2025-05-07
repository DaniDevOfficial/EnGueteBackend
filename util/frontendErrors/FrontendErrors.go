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

	FiltersAreNotValidError = "filtersAreNotValidError"
)
