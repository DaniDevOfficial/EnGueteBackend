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

func CheckIfUserExistsByEmailOrUsername(email, username string, db *sql.DB) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email ILIKE $1 OR username ILIKE $2`
	row := db.QueryRow(query, email, username)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetUserByName(username string, db *sql.DB) (UserFromDB, error) {
	query := `SELECT 
				username,
				email,
				password_hash,
				user_id
 			FROM 
				users
			WHERE username = $1
			AND deleted_at IS NULL
				`
	row := db.QueryRow(query, username)
	var userData UserFromDB
	err := row.Scan(&userData.Username, &userData.Email, &userData.PasswordHash, &userData.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		return UserFromDB{}, ErrUserNotFound
	}
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
    		WHERE user_id = $1
    		AND deleted_at IS NULL
    		`
	row := db.QueryRow(query, userId)

	var userData UserFromDB
	err := row.Scan(&userData.Username, &userData.Email, &userData.PasswordHash, &userData.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		return UserFromDB{}, ErrUserNotFound
	}
	return userData, err
}

func GetUsersGroupByUserIdFromDB(userId string, db *sql.DB) ([]GroupCard, error) {

	query := `
		SELECT
			g.group_id,
			g.group_name,
			COUNT(DISTINCT ug.user_id) AS user_count
		FROM groups g
		INNER JOIN user_groups ug ON g.group_id = ug.group_id
		INNER JOIN users u ON ug.user_id = u.user_id
		WHERE u.user_id = $1
		AND g.deleted_at IS NULL
		AND u.deleted_at IS NULL
		AND ug.deleted_at IS NULL
		GROUP BY g.group_id
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
		err := rows.Scan(&thisUserGroup.GroupId, &thisUserGroup.GroupName, &thisUserGroup.UserCount)
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

var ErrUserNotFound = errors.New("user not found")

func UpdateUsernameInDB(newUsername string, userId string, db *sql.DB) error {
	query := `	UPDATE users
				SET	username = $1
				WHERE user_id = $2
`
	result, err := db.Exec(query, newUsername, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

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

func DeleteUserInDB(userId string, db *sql.DB) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	updateUserQuery := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	_, err = tx.Exec(updateUserQuery, userId)
	if err != nil {
		transactionErr := tx.Rollback()
		if transactionErr != nil {
			return transactionErr
		}
		return err
	}

	updateUserGroupsQuery := `
		UPDATE user_groups
		SET deleted_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	_, err = tx.Exec(updateUserGroupsQuery, userId)
	if err != nil {
		transactionErr := tx.Rollback()
		if transactionErr != nil {
			return transactionErr
		}
		return err
	}

	updateMealPreferencesQuery := `
		UPDATE meal_preferences
		SET deleted_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	_, err = tx.Exec(updateMealPreferencesQuery, userId)
	if err != nil {
		transactionErr := tx.Rollback()
		if transactionErr != nil {
			return transactionErr
		}
		return err
	}

	deleteRefreshTokensQuery := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`
	_, err = tx.Exec(deleteRefreshTokensQuery, userId)
	if err != nil {
		transactionErr := tx.Rollback()
		if transactionErr != nil {
			return transactionErr
		}
		return err
	}
	deleteUserGroupRolesQuery := `
		DELETE FROM user_group_roles
		WHERE user_id = $1
	`
	_, err = tx.Exec(deleteUserGroupRolesQuery, userId)
	if err != nil {
		transactionErr := tx.Rollback()
		if transactionErr != nil {
			return transactionErr
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
