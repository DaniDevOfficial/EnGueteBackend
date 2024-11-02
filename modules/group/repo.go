package group

import (
	"database/sql"
)

func CreateNewGroupInDB(groupData RequestNewGroup, userId string, db *sql.DB) (string, error) {
	query := `	INSERT INTO groups 
    				(group_name, created_by)
				VALUES
    				($1, $2)
     			RETURNING group_id`

	var groupId string

	err := db.QueryRow(query, groupData.GroupName, userId).Scan(&groupId)
	if err != nil {
		return "", err
	}

	return groupId, nil
}
