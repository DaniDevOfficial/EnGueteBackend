package user

import (
	"database/sql"
	"errors"
)

func GetUserIdByName(username string, db *sql.DB) (string, error) {
	query := "SELECT user_id FROM users WHERE username = $1"
	row := db.QueryRow(query, username)
	var userId string
	err := row.Scan(&userId)
	if errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return userId, err
}

func GetUserByName(username string, db *sql.DB) (UserFromDB, error) {
	query := `SELECT 
				username,
				email,
				password,
				id
			FROM
				users
			WHERE
				username = $1`
	row := db.QueryRow(query, username)
	var userData UserFromDB
	err := row.Scan(&userData.username, &userData.email, &userData.passwordHash, &userData.userId)
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
        		uuid = $1`
	row := db.QueryRow(query, userId)

	var userData UserFromDB
	err := row.Scan(&userData.username, &userData.email, &userData.passwordHash, &userData.userId)
	return userData, err
}

func CreateUserInDB(userData DBNewUser, db *sql.DB) (string, error) {
	query := `	INSERT INTO users 
    				(username, email, password)
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
					username = $1
				WHERE id = $2
`
	_, err := db.Exec(query, newUsername, userId)
	return err
}

func UpdatePasswordInDb(newPassword string, userId string, db *sql.DB) error {
	query := `	UPDATE users
					password = $1
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
