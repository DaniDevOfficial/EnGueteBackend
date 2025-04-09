package user

import (
	"database/sql"
	"errors"
	"log"
)

func GetUserIdByName(username string, db *sql.DB) (string, error) {
	query := "SELECT user_id FROM users WHERE username = $1"
	row := db.QueryRow(query, username)
	var userId string
	err := row.Scan(&userId)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return userId, err
}

func GetUserByName(username string, db *sql.DB) (UserFromDB, error) {
	query := `SELECT 
				username,
				email,
				password_hash,
				user_id
 			FROM 
				users
			WHERE
				username = $1`
	row := db.QueryRow(query, username)
	var userData UserFromDB
	err := row.Scan(&userData.Username, &userData.Email, &userData.PasswordHash, &userData.UserId)
	return userData, err
}

func GetUserByIdFromDB(userId string, db *sql.DB) (UserFromDB, error) {

	query := `SELECT
    			username,
    			email,
    			password_hash,
    			user_id
    		FROM
    		    users
    		WHERE 
        		user_id = $1`
	row := db.QueryRow(query, userId)

	var userData UserFromDB
	err := row.Scan(&userData.Username, &userData.Email, &userData.PasswordHash, &userData.UserId)
	return userData, err
}

func GetUsersGroupByUserIdFromDB(userId string, db *sql.DB) ([]GroupCard, error) {

	query := `
		SELECT
			g.group_id,
			g.group_name,
			COUNT(DISTINCT ug.user_id) AS user_count
		FROM
			groups g
		LEFT JOIN 
			user_groups ug ON g.group_id = ug.group_id
		LEFT JOIN 
			users u ON ug.user_id = u.user_id
		WHERE 
			u.user_id = $1
		GROUP BY
			g.group_id
	`
	rows, err := db.Query(query, userId)
	var userGroups []GroupCard
	if err != nil {
		log.Println(err)
		return userGroups, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var thisUserGroup GroupCard
		err := rows.Scan(&thisUserGroup.GroupId, &thisUserGroup.GroupName, &thisUserGroup.AmountOfPeopleInGroup)
		if err != nil {
			return userGroups, err
		}
		thisUserGroup.NextMealDate = "123"
		userGroups = append(userGroups, thisUserGroup)
	}

	return userGroups, err
}

func CreateUserInDB(userData DBNewUser, db *sql.DB) (string, error) {
	query := `	INSERT INTO users 
    				(username, email, password_hash)
				VALUES
    				($1, $2, $3)
     			RETURNING user_id`

	var userId string

	err := db.QueryRow(query, userData.username, userData.email, userData.password_hash).Scan(&userId)
	if err != nil {
		return "", err
	}

	return userId, nil
}

func UpdateUsernameInDB(newUsername string, userId string, db *sql.DB) error {
	query := `	UPDATE users
				SET	username = $1
				WHERE user_id = $2
`
	_, err := db.Exec(query, newUsername, userId)
	return err
}

func UpdatePasswordInDb(newPassword string, userId string, db *sql.DB) error {
	query := `	UPDATE users
				SET password_hash = $1
				WHERE user_id = $2
`
	_, err := db.Exec(query, newPassword, userId)
	return err
}

func DeleteUserInDB(userId string, db *sql.DB) (bool, error) {
	query := `	DELETE FROM 
	           		users
				WHERE 
				    user_id = $1
				`
	_, err := db.Exec(query, userId)
	if err != nil {
		return false, err
	}
	return true, nil

}
