package management

import "database/sql"

func KickUSerFromGroupInDB(userId string, groupId string, db *sql.DB) error {
	query := `
	DELETE FROM user_groups
	WHERE user_id = $1 AND group_id = $2
` // TODO: remove any not closed optin out states for this user
	_, err := db.Exec(query, userId, groupId)
	return err
}

func UnBanUserFromGroupInDB(userId string, groupId string, db *sql.DB) error {
	query := `
	DELETE FROM user_groups_blacklist
	WHERE user_id = $1 AND group_id = $2
`
	_, err := db.Exec(query, userId, groupId)
	return err
}
