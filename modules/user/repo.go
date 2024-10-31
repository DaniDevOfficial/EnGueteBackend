package user

import (
	"database/sql"
	"errors"
)

func GetUserIdByName(username string, db *sql.DB) (int, error) {
	query := "SELECT id FROM users WHERE username = $1"
	row := db.QueryRow(query, username)
	var userId int
	err := row.Scan(&userId)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, err
	}
	return userId, err
}

func GetUserByName(username string, db *sql.DB) (UserFromDB, error) {
	query := `SELECT 
				username,
				email,
				password,
				id,
				uuid
			FROM
				users
			WHERE
				username = $1`
	row := db.QueryRow(query, username)
	var userData UserFromDB
	err := row.Scan(&userData.username, &userData.email, &userData.password_hash, &userData.user_id, &userData.uuid)
	return userData, err
}

func GetUserById(id int, db *sql.DB) (UserFromDB, error) {

	query := `SELECT
    			username,
    			email,
    			password_hash,
    			user_id,
    			uuid
    		FROM
    		    users
    		WHERE 
        		user_id = $1`
	row := db.QueryRow(query, id)

	var userData UserFromDB
	err := row.Scan(&userData.username, &userData.email, &userData.password_hash, &userData.user_id)
	return userData, err
}

func GetUserByUUIDFromDB(uuid string, db *sql.DB) (UserFromDB, error) {

	query := `SELECT
    			username,
    			email,
    			password_hash,
    			user_id,
    			uuid
    		FROM
    		    users
    		WHERE 
        		uuid = $1`
	row := db.QueryRow(query, uuid)

	var userData UserFromDB
	err := row.Scan(&userData.username, &userData.email, &userData.password_hash, &userData.user_id)
	return userData, err
}

func CreateUserInDB(userData DBNewUser, db *sql.DB) (int64, string, error) {
	query := `	INSERT INTO users 
    				(username, email, password)
				VALUES
    				($1, $2, $3)
     			RETURNING id, uuid`

	var id int64
	var uuid string

	err := db.QueryRow(query, userData.username, userData.email, userData.password_hash).Scan(&id, &uuid)
	if err != nil {
		return -1, "", err
	}

	return id, uuid, nil
}

func UpdateUsernameInDB(newUsername string, userId int, db *sql.DB) error {
	query := `	UPDATE users
					username = $1
				WHERE id = $2
`
	_, err := db.Exec(query, newUsername, userId)
	return err
}

func UpdatePasswordInDb(newPassword string, userId int, db *sql.DB) error {
	query := `	UPDATE users
					password = $1
				WHERE id = $2
`
	_, err := db.Exec(query, newPassword, userId)
	return err
}

func DeleteUserInDB(id int, db *sql.DB) (bool, error) {
	query := `	DELETE FROM 
	           		users
				WHERE 
				    id = $1
				`
	_, err := db.Exec(query, id)
	if err != nil {
		return false, err
	}
	return true, nil

}
