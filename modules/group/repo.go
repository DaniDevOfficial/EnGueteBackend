package group

import (
	"database/sql"
	"errors"
	"time"
)

func CreateNewGroupInDBWithTransaction(groupData RequestNewGroup, userId string, tx *sql.Tx) (string, error) {
	query := `INSERT INTO groups (group_name, created_by) VALUES ($1, $2) 
	RETURNING group_id`

	var groupId string

	err := tx.QueryRow(query, groupData.GroupName, userId).Scan(&groupId)
	if err != nil {
		return "", err
	}

	return groupId, nil
}

func AddUserToGroupInDB(groupId string, userId string, db *sql.DB) (bool, error) {
	query := `
		INSERT INTO user_groups (group_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (group_id, user_id) DO NOTHING
		RETURNING true
	`

	var result bool
	err := db.QueryRow(query, groupId, userId).Scan(&result)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return result, err
}

func AddUserToGroupWithTransaction(groupId string, userId string, tx *sql.Tx) (string, error) {
	query := `INSERT INTO user_groups (group_id, user_id) VALUES ($1, $2) RETURNING user_group_id`
	var userGroupId string
	err := tx.QueryRow(query, groupId, userId).Scan(&userGroupId)
	return userGroupId, err
}

func AddRoleToUserInGroupWithTransaction(groupId string, userId string, role string, userGroupId string, tx *sql.Tx) error {
	query := `INSERT INTO user_group_roles (group_id, user_id, role, user_groups_id) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(query, groupId, userId, role, userGroupId)
	return err
}

func AddRoleToUserInGroup(groupId string, userId string, role string, db *sql.DB) error {
	query := `INSERT INTO user_group_roles (group_id, user_id, role) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, groupId, userId, role)
	return err
}

func GetGroupInformationFromDb(groupId string, userId string, db *sql.DB) (GroupInfo, error) {
	query := `
	SELECT
	    g.group_name,
		COUNT(ug.user_id) AS user_count,
	    ARRAY_AGG(ur.role) AS user_roles
	FROM groups g 
	LEFT JOIN user_groups ug ON ug.group_id = g.group_id
	LEFT JOIN user_group_roles ur ON ur.group_id = g.group_id AND ur.user_id = $2
	WHERE g.group_id = $1
	GROUP BY g.group_id
`

	var info GroupInfo
	if err := db.QueryRow(query, groupId, userId).Scan(&info.GroupName, &info.UserCount, &info.UserRoles); err != nil {
		return info, err
	}
	return info, nil
}

func GetGroupMembersFromDb(groupId string, db *sql.DB) ([]Member, error) {
	query := `
		SELECT 
    		u.username,
    		u.user_id,
    		ARRAY_AGG(ur.role) AS user_roles
		FROM user_groups ug
		LEFT JOIN users u ON ug.user_id = u.user_id
		LEFT JOIN user_group_roles ur ON ur.group_id = ug.group_id AND ur.user_id = u.user_id
		WHERE ug.group_id = $1
		GROUP BY u.username, u.user_id;
`
	rows, err := db.Query(query, groupId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Member{}, nil
		}
		return nil, err
	}

	defer rows.Close()

	var members []Member
	for rows.Next() {
		var member Member
		err = rows.Scan(&member.Username, &member.UserId, &member.UserRoles)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

var ErrNothingHappened = errors.New("nothing happened")

func RemoveRoleFromUserInGroup(groupId string, userId string, role string, db *sql.DB) error {
	query := `DELETE FROM user_group_roles WHERE group_id = $1 AND user_id = $2 AND role = $3 RETURNING group_id`
	var groupIdTmp string
	err := db.QueryRow(query, groupId, userId, role).Scan(&groupIdTmp)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNothingHappened
	}
	return err
}

var ErrUserIsNotPartOfThisGroup = errors.New("user is not part of this group")

func IsUserMemberOfGroupViaMealId(mealId string, userId string, db *sql.DB) (int, error) {
	query := `
	SELECT 
		1
	FROM groups g
	LEFT JOIN meals m ON m.group_id = g.group_id
	LEFT JOIN user_groups gu ON gu.group_id = g.group_id
	WHERE m.meal_id = $2
	AND gu.user_id = $1
`
	var exists int
	err := db.QueryRow(query, userId, mealId).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exists, ErrUserIsNotPartOfThisGroup
		}
	}
	return exists, err
}

func IsUserMemberOfGroupInDB(groupId string, userId string, db *sql.DB) (int, error) {
	query := `
	SELECT 
		1
	FROM user_groups gu
	WHERE gu.group_id = $2
	AND gu.user_id = $1
`
	var exists int
	err := db.QueryRow(query, userId, groupId).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exists, ErrUserIsNotPartOfThisGroup
		}
	}
	return exists, err
}

func GetUserRolesInGroup(groupId string, userId string, db *sql.DB) ([]string, error) {
	query := `
	SELECT role
	FROM user_group_roles
	WHERE user_id = $1
	AND group_id = $2
`
	rows, err := db.Query(query, userId, groupId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserIsNotPartOfThisGroup
		}
		return nil, err
	}

	var userRoles []string
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var userRole string
		err := rows.Scan(&userRole)
		if err != nil {
			return userRoles, err
		}
		userRoles = append(userRoles, userRole)
	}

	return userRoles, nil
}

func GetUserRolesInGroupViaMealId(mealId string, userId string, db *sql.DB) ([]string, error) {
	query := `
	SELECT ugr.role
	FROM user_group_roles ugr
	LEFT JOIN meals m ON m.group_id = g.group_id
	LEFT JOIN user_groups gu ON gu.group_id = g.group_id
	WHERE m.meal_id = $2
	AND user_group_roles.user_id = $1
`
	rows, err := db.Query(query, userId, mealId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserIsNotPartOfThisGroup
		}
		return nil, err
	}

	var userRoles []string
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var userRole string
		err := rows.Scan(&userRole)
		if err != nil {
			return userRoles, err
		}
		userRoles = append(userRoles, userRole)
	}
	return userRoles, nil
}

func CreateNewInviteInDBWithTransaction(groupId string, tx *sql.Tx) (string, error) {
	query := `
		INSERT INTO group_invites 
		    (group_id, expires_at)
		VALUES 
		    ($1, $2)
		RETURNING invite_token`
	var inviteToken string
	expirationTime := time.Now().Add(24 * time.Hour)

	err := tx.QueryRow(query, groupId, expirationTime).Scan(&inviteToken)
	if err != nil {
		return "", err
	}
	return inviteToken, nil
}

func ValidateInviteTokenInDB(inviteToken string, db *sql.DB) (string, error) {
	query := `
		WITH deleted AS (
			DELETE FROM group_invites
			WHERE invite_token = $1 AND expires_at <= NOW()
		)
		SELECT group_id FROM group_invites
		WHERE invite_token = $1
	`

	var groupId string
	err := db.QueryRow(query, inviteToken).Scan(&groupId)

	return groupId, err
}

func VoidInviteTokenIfAllowedInDB(inviteToken string, userId string, db *sql.DB) error {
	query := `
	DELETE FROM group_invites gi
	USING groups
	WHERE gi.invite_token = $1
	AND groups.group_id = gi.group_id 
	AND groups.created_by = $2
`
	_, err := db.Exec(query, inviteToken, userId)
	return err
}

var ErrNoMatchingGroupOrUser = errors.New("no matching group or user found for deletion")

func LeaveGroupInDB(groupId string, userId string, db *sql.DB) error {
	query := `
		DELETE FROM user_groups
		WHERE group_id = $1
		AND user_id = $2
	`

	result, err := db.Exec(query, groupId, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoMatchingGroupOrUser
	}

	return nil
}
