package management

import "database/sql"

func UnBanUserFromGroupInDB(userId string, groupId string, db *sql.DB) error {
	query := `
	DELETE FROM user_groups_blacklist
	WHERE user_id = $1 AND group_id = $2
`
	_, err := db.Exec(query, userId, groupId)
	return err
}
