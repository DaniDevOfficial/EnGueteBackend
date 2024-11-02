package group

import (
	"database/sql"
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

func AddUserToGroup(groupId string, userId string, db *sql.DB) error {
	query := `INSERT INTO group_users (group_id, user_id) VALUES ($1, $2)`
	_, err := db.Exec(query, groupId, userId)
	return err
}

func AddUserToGroupWithTransaction(groupId string, userId string, tx *sql.Tx) error {
	query := `INSERT INTO group_users (group_id, user_id) VALUES ($1, $2)`
	_, err := tx.Exec(query, groupId, userId)
	return err
}
