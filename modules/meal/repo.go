package meal

import (
	"database/sql"
	"errors"
)

func CreateNewMealInDB(newMeal RequestNewMeal, userId string, db *sql.DB) (string, error) {
	query := `INSERT INTO meals
				(title, notes, date_time, meal_type, created_by, group_id)
			VALUES
				($1, $2, $3, $4, $5, $6)
			RETURNING
				meal_id`
	row := db.QueryRow(query, newMeal.Title, newMeal.Notes, newMeal.ScheduledAt, newMeal.Type, userId, newMeal.GroupId)
	var mealId string
	err := row.Scan(&mealId)
	return mealId, err
}

func AddCookToMealInDB(userId string, groupId string, db *sql.DB) error {
	query := `INSERT INTO meal_cooks
    				(user_id, meal_id)
    				VALUES
    				($1, $2)`
	_, err := db.Exec(query, userId, groupId)
	return err
}

var ErrUserWasntACook = errors.New("user wasn't a Cook")

func RemoveCookFromMealInDB(userId string, groupId string, db *sql.DB) error {
	query := `
        DELETE FROM meal_cooks 
        WHERE user_id = $1 AND group_id = $2 
        RETURNING user_id
    `
	var deletedUserId string
	err := db.QueryRow(query, userId, groupId).Scan(&deletedUserId)
	if err == sql.ErrNoRows {
		return ErrUserWasntACook
	}
	return err
}
