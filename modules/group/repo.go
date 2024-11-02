package group

import (
	"database/sql"
	"errors"
	"fmt"
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

func AddUserToGroupInDB(groupId string, userId string, db *sql.DB) error {
	query := `INSERT INTO users_group (group_id, user_id) VALUES ($1, $2)`
	_, err := db.Exec(query, groupId, userId)
	return err
}

func AddUserToGroupWithTransaction(groupId string, userId string, tx *sql.Tx) error {
	query := `INSERT INTO user_groups (group_id, user_id) VALUES ($1, $2)`
	_, err := tx.Exec(query, groupId, userId)
	return err
}

func CheckIfUserIsAdminOrOwnerOfGroupInDB(groupId string, userId string, db *sql.DB) error {
	query := `
	SELECT 
		1
	FROM groups g
	LEFT JOIN user_groups gu ON gu.group_id = g.group_id
	WHERE gu.user_id = $1
	AND g.created_by = $1
	AND g.group_id = $2
` //TODO: do some table for group admins

	row := db.QueryRow(query, userId, groupId)
	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user is neither admin nor owner of the group")
		}
		return err
	}

	return nil
}

func CreateNewInviteInDBWithTransaction(groupId string, userId string, tx *sql.Tx) (string, error) {
	query := `
		INSERT INTO group_invites 
		    (group_id, expires_at)
		VALUES 
		    ($1, $2)
		RETURNING invite_token`
	var inviteToken string
	currentTime := time.Now()

	err := tx.QueryRow(query, groupId, currentTime).Scan(&inviteToken)
	if err != nil {
		return "", err
	}
	return inviteToken, nil
}

func ValidateInviteTokenInDB(inviteToken string, db *sql.DB) (string, error) {
	query := `SELECT group_id FROM group_invites WHERE invite_token = $1 AND expires_at > NOW()`
	var groupId string
	err := db.QueryRow(query, inviteToken).Scan(&groupId)
	return groupId, err
}
