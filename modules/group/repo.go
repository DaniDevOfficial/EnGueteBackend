package group

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
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

func DeleteGroupInDB(groupId string, db *sql.DB) error {
	query := `
		DELETE FROM groups
		WHERE group_id = $1
	`

	result, err := db.Exec(query, groupId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNothingHappened
	}

	return nil
}

func UpdateGroupNameInDB(groupInfo RequestUpdateGroupName, db *sql.Tx) error {
	query := `UPDATE groups SET group_name = $1 WHERE group_id = $2`

	result, err := db.Exec(query, groupInfo.GroupName, groupInfo.GroupId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNothingHappened
	}

	return nil
}

func GetGroupInformationFromDb(groupId string, userId string, db *sql.DB) (GroupInfo, error) {
	query := `
	SELECT
	    g.group_id,
	    g.group_name,
		COUNT(DISTINCT ug.user_id) AS user_count,
	    ARRAY_AGG(ur.role) AS user_roles
	FROM groups g 
	LEFT JOIN user_groups ug ON ug.group_id = g.group_id
	LEFT JOIN user_group_roles ur ON ur.group_id = g.group_id AND ur.user_id = $2
	WHERE g.group_id = $1
	GROUP BY g.group_id
`

	var info GroupInfo
	var userRoles pq.StringArray

	if err := db.QueryRow(query, groupId, userId).Scan(&info.GroupId, &info.GroupName, &info.UserCount, &userRoles); err != nil {
		return info, err
	}

	info.UserRoles = userRoles
	return info, nil
}

func GetGroupMembersFromDb(groupId string, db *sql.DB) ([]Member, error) {
	query := `
		SELECT 
    		ug.group_id,
    		u.user_id,
    		u.username,
    		ARRAY_AGG(ur.role) AS user_roles
		FROM user_groups ug
		INNER JOIN users u ON ug.user_id = u.user_id
		INNER JOIN user_group_roles ur ON ur.group_id = ug.group_id AND ur.user_id = u.user_id
		WHERE ug.group_id = $1
		GROUP BY ug.group_id, u.user_id, u.username;
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
		var userRoles pq.StringArray

		err = rows.Scan(&member.GroupId, &member.UserId, &member.Username, &userRoles)
		if err != nil {
			return nil, err
		}
		member.UserRoles = userRoles
		members = append(members, member)
	}

	return members, nil
}

func GetMealsInGroupDB(filters FilterGroupRequest, userId string, db *sql.DB) ([]MealCard, error) {
	query := `
        SELECT 
            m.meal_id,
            m.title,
            m.closed,
            m.fulfilled,
            m.date_time,
            m.meal_type,
            m.notes,
            COUNT(CASE WHEN mp.preference = 'opt-in' OR mp.preference = 'eat later' THEN 1 END) AS participant_count,
            COALESCE(user_pref.preference, 'undecided') AS user_preference,
            CASE WHEN mc.user_id IS NOT NULL THEN true ELSE false END AS is_cook
        FROM meals m
        LEFT JOIN meal_preferences mp ON mp.meal_id = m.meal_id
        LEFT JOIN meal_preferences user_pref ON user_pref.meal_id = m.meal_id AND user_pref.user_id = $2
        LEFT JOIN meal_cooks mc ON mc.meal_id = m.meal_id AND mc.user_id = $2
        WHERE m.group_id = $1
        AND ($3::timestamp IS NULL OR $4::timestamp IS NULL OR m.date_time BETWEEN $3 AND $4)

        GROUP BY m.meal_id, user_pref.preference, mc.user_id
        ORDER BY m.date_time desc 
`
	rows, err := db.Query(query, filters.GroupId, userId, filters.StartDateFilter, filters.EndDateFilter)
	var mealCards []MealCard
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mealCards, nil
		}
		return mealCards, err
	}

	defer rows.Close()

	for rows.Next() {
		var mealCard MealCard
		err := rows.Scan(
			&mealCard.MealId,
			&mealCard.Title,
			&mealCard.Closed,
			&mealCard.Fulfilled,
			&mealCard.DateTime,
			&mealCard.MealType,
			&mealCard.Notes,
			&mealCard.ParticipantCount,
			&mealCard.UserPreference,
			&mealCard.IsCook,
		)
		if err != nil {
			return mealCards, err
		}
		mealCards = append(mealCards, mealCard)
	}

	return mealCards, nil

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

func IsUserMemberOfGroupViaMealId(mealId string, userId string, db *sql.DB) (string, error) {
	query := `
	SELECT 
		g.group_id
	FROM groups g
	INNER JOIN meals m ON m.group_id = g.group_id
	INNER JOIN user_groups gu ON gu.group_id = g.group_id
	WHERE m.meal_id = $2
	AND gu.user_id = $1
`
	var exists string
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
	LEFT JOIN meals m ON m.group_id = ugr.group_id
	LEFT JOIN user_groups gu ON gu.group_id = ugr.group_id
	WHERE m.meal_id = $2
	AND ugr.user_id = $1
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

func CreateNewInviteInDBWithTransaction(inviteData InviteLinkGenerationRequest, tx *sql.Tx) (string, error) {
	query := `
		INSERT INTO group_invites 
		    (group_id, expires_at)
		VALUES 
		    ($1, $2)
		RETURNING invite_token`
	var inviteToken string

	err := tx.QueryRow(query, inviteData.GroupId, inviteData.ExpirationDateTime).Scan(&inviteToken)
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

func GetAllInviteTokensInAGroupFromDB(groupId string, db *sql.DB) ([]InviteToken, error) {
	query := `
		WITH deleted AS (
			DELETE FROM group_invites
			WHERE expires_at <= NOW()
		)
		SELECT invite_token, expires_at
		FROM group_invites
		WHERE group_id = $1
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY expires_at;
	`

	rows, err := db.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inviteTokens []InviteToken
	for rows.Next() {
		var inviteToken InviteToken
		err := rows.Scan(&inviteToken.InviteToken, &inviteToken.ExpiresAt)
		if err != nil {
			return nil, err
		}
		inviteTokens = append(inviteTokens, inviteToken)
	}

	return inviteTokens, nil
}

func VoidInviteTokenInDB(inviteToken string, db *sql.DB) error {
	query := `
	DELETE FROM group_invites gi
	WHERE gi.invite_token = $1
`
	result, err := db.Exec(query, inviteToken)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNothingHappened
	}

	return err
}

var ErrNoMatchingGroupOrUser = errors.New("no matching group or user found for deletion")

func LeaveGroupInDB(groupId string, userId string, tx *sql.Tx) error {
	query := `
		DELETE FROM user_groups
		WHERE group_id = $1
		AND user_id = $2
	`

	result, err := tx.Exec(query, groupId, userId)
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

func RemovePreferencesInOpenMealsInGroup(userId string, groupId string, tx *sql.Tx) error {
	query := `
		DELETE FROM meal_preferences mp
		USING meals m
		WHERE m.meal_id = mp.meal_id
		  AND m.group_id = $2
		  AND mp.user_id = $1
		  AND NOT m.closed
		  AND NOT m.fulfilled;
`
	_, err := tx.Exec(query, userId, groupId)
	return err
}
