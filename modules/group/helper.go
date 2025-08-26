package group

import (
	"database/sql"
	"enguete/util/roles"
	"errors"
)

// IsUserInGroupViaMealId Check if the target user is part of the group
//
// return true => user is in group
func IsUserInGroupViaMealId(mealId string, userId string, db *sql.DB) (string, error) {
	groupId, err := IsUserMemberOfGroupViaMealId(mealId, userId, db)
	if err != nil {
		return groupId, err
	}
	return groupId, nil // User is in group
}

// IsUserInGroup Check if the target user is part of the group
//
// return true => user is in group
//
// return err => internal server error
//
// return false => user is not in group
func IsUserInGroup(groupId string, userId string, db *sql.DB) (bool, error) {
	_, err := IsUserMemberOfGroupInDB(groupId, userId, db)
	if err != nil {
		if errors.Is(err, ErrUserIsNotPartOfThisGroup) {
			return false, nil // Not in group, but no internal error
		}
		return false, err // Internal error occurred
	}
	return true, nil // User is in group
}

// CheckIfUserIsAllowedToPerformActionViaMealId Check if the user is able to perform an action in a group via mealId.
//
// return true => user is in group and can perform action.
//
// return err => internal server error.
//
// return false => user is not in group or cant perform action
func CheckIfUserIsAllowedToPerformActionViaMealId(mealId string, userId string, actionToPerform string, db *sql.DB) (bool, []string, error) {
	userRoles, err := GetUserRolesInGroupViaMealId(mealId, userId, db)
	if err != nil {
		return false, nil, err
	}
	return roles.CanPerformAction(userRoles, actionToPerform), userRoles, nil
}

// CheckIfUserIsAllowedToPerformAction Check if the user is able to perform an action in a group.
//
// return true => user is in group and can perform action.
//
// return err => internal server error.
//
// return false => user is not in group or cant perform action
func CheckIfUserIsAllowedToPerformAction(groupId string, userId string, actionToPerform string, db *sql.DB) (isAllowedToPerformAction bool, userRoles []string, error error) {
	userRoles, err := GetUserRolesInGroup(groupId, userId, db)
	if err != nil {
		return false, userRoles, err
	}
	return roles.CanPerformAction(userRoles, actionToPerform), userRoles, nil
}
