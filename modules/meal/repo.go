package meal

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"time"
)

//General

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

func DeleteMealInDB(mealId string, db *sql.DB) error {
	query := `DELETE FROM meals WHERE meal_id=$1`
	_, err := db.Exec(query, mealId)
	return err
}

var ErrNoData = errors.New("no data found")

func GetSingularMealInformation(mealId string, userId string, db *sql.DB) (MealInformation, error) {
	query := `
        SELECT 
            m.meal_id,
            m.title,
            m.closed,
            m.fulfilled,
            m.date_time,
            m.meal_type,
            m.notes,
            COUNT(CASE WHEN mp.preference = 'opt-in' OR mp.preference = 'eat later' THEN 1 END) AS participant_count,
            CASE WHEN mc.user_id IS NOT NULL THEN true ELSE false END AS is_cook
        FROM meals m
        LEFT JOIN meal_preferences mp ON mp.meal_id = m.meal_id
        LEFT JOIN meal_cooks mc ON mc.meal_id = m.meal_id
        WHERE m.meal_id = $1
        GROUP BY m.meal_id, mc.user_id
        ORDER BY m.date_time
`
	var mealInformation MealInformation
	err := db.QueryRow(query, mealId).Scan(
		&mealInformation.MealID,
		&mealInformation.Title,
		&mealInformation.Closed,
		&mealInformation.Fulfilled,
		&mealInformation.DateTime,
		&mealInformation.MealType,
		&mealInformation.Notes,
		&mealInformation.ParticipantCount,
		&mealInformation.IsCook,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mealInformation, ErrNoData
		}
		return mealInformation, err
	}

	return mealInformation, nil
}

func GetMealParticipationInformationFromDB(mealId string, db *sql.DB) ([]MealParticipant, error) {
	query := `
		SELECT 
    		u.username,
    		u.user_id,
    		COALESCE(mp.preference, 'undecided') AS preference,
    		CASE
        		WHEN mc.user_id IS NOT NULL THEN TRUE
        		ELSE FALSE
    		END AS is_cook
		FROM users u
		LEFT JOIN meal_preferences mp ON u.user_id = mp.user_id AND mp.meal_id = $1 AND mp.preference = 'opt-in'
		LEFT JOIN meal_cooks mc ON u.user_id = mc.user_id AND mc.meal_id = $1
		WHERE mp.meal_id = $1 OR mc.meal_id = $1;
`
	rows, err := db.Query(query, mealId)
	var mealParticipants []MealParticipant
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mealParticipants, nil
		}
		return mealParticipants, err
	}
	defer rows.Close()

	for rows.Next() {
		var mealParticipant MealParticipant
		err := rows.Scan(
			&mealParticipant.Username,
			&mealParticipant.UserId,
			&mealParticipant.Preference,
			&mealParticipant.IsCook,
		)
		if err != nil {
			return mealParticipants, err
		}
	}
	return mealParticipants, nil
}

var ErrUserAlreadyHasAPreferenceInSpecificMeal = errors.New("user already has A Preference")

// Flags

func UpdateClosedBoolInDB(mealId string, isClosed bool, db *sql.DB) error {
	query := `UPDATE meals SET closed=$1 WHERE meal_id=$2 RETURNING closed` // TODO Swap the closed bool from what it currently is
	var tmp bool
	err := db.QueryRow(query, isClosed, mealId).Scan(&tmp)
	return err
}

func UpdateMealFulfilledStatus(mealId string, isFulfilled bool, db *sql.DB) error {
	query := `UPDATE meals SET fulfilled=$1 WHERE meal_id=$2 RETURNING fulfilled` // TODO Swap the closed bool from what it currently is
	var tmp bool
	err := db.QueryRow(query, isFulfilled, mealId).Scan(&tmp)
	return err
}

//OptIn Status

func OptInMealInDB(userId string, optData RequestOptInMeal, db *sql.DB) error {
	query := `
	INSERT INTO meal_preferences
	(meal_id, user_id, preference)
	VALUES 
	    ($1, $2, $3)`

	_, err := db.Exec(query, optData.MealId, userId, optData.Preference)
	if err != nil {
		// Check for unique constraint violation using pq's error code
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserAlreadyHasAPreferenceInSpecificMeal
		}
		// Return other errors as is
		return err
	}
	return err
}

func ChangeOptInStatusMealInDB(userId string, optData RequestOptInMeal, db *sql.DB) error {
	query := `
		UPDATE meal_preferences
		SET preference = $1, changed_at = $2
		WHERE meal_id = $3
		AND user_id = $4
		RETURNING meal_id`

	var updatedMealID string
	err := db.QueryRow(query, optData.Preference, time.Now(), optData.MealId, userId).Scan(&updatedMealID)

	if err != nil {
		// Check for unique constraint violation using pq's error code
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserAlreadyHasAPreferenceInSpecificMeal
		}
		// Return other errors as is
		return err
	}

	// If we reach this point, the update was successful
	return nil
}

// Meal Cooks

var ErrUserWasntACook = errors.New("user wasn't a Cook")

func RemoveCookFromMealInDB(userId string, meal_id string, db *sql.DB) error {
	query := `
        DELETE FROM meal_cooks 
        WHERE user_id = $1 AND meal_id = $2 
        RETURNING user_id
    `
	var deletedUserId string
	err := db.QueryRow(query, userId, meal_id).Scan(&deletedUserId)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserWasntACook
	}
	return err
}

func AddCookToMealInDB(userId string, mealId string, db *sql.DB) error {
	query := `INSERT INTO meal_cooks
    				(user_id, meal_id)
    				VALUES
    				($1, $2)`
	_, err := db.Exec(query, userId, mealId)
	return err
}

var ErrDataCouldNotBeUpdated = errors.New("data couldn't be updated")

// Meal Update

func UpdateMealTitleIdDB(mealId string, newTitle string, db *sql.DB) error {
	query := `
	UPDATE meals
	SET title = $1
	WHERE meal_id = $2
	RETURNING meal_id
`
	var updatedMealId string
	err := db.QueryRow(query, newTitle, mealId).Scan(&updatedMealId)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDataCouldNotBeUpdated
	}
	//TODO: maybe implement a second query, which selects all the required data again from the db for real time data
	return err
}

func UpdateMealTypeInDB(mealId string, newType string, db *sql.DB) error {
	query := `
	UPDATE meals
	SET meal_type = $1
	WHERE meal_id = $2
	RETURNING meal_id
`
	var updatedMealId string
	err := db.QueryRow(query, newType, mealId).Scan(&updatedMealId)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDataCouldNotBeUpdated
	}
	//TODO: maybe implement a second query, which selects all the required data again from the db for real time data
	return err
}

func UpdateMealNotesInDB(mealId string, newNotes string, db *sql.DB) error {
	query := `
	UPDATE meals
	SET notes = $1
	WHERE meal_id = $2
	RETURNING meal_id
`
	var updatedMealId string
	err := db.QueryRow(query, newNotes, mealId).Scan(&updatedMealId)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDataCouldNotBeUpdated
	}
	//TODO: maybe implement a second query, which selects all the required data again from the db for real time data
	return err
}

func UpdateMealScheduledAtInDB(mealId string, newScheduledAt string, db *sql.DB) error {
	query := `
	UPDATE meals
	SET date_time = $1
	WHERE meal_id = $2
	RETURNING meal_id
`
	var updatedMealId string
	err := db.QueryRow(query, newScheduledAt, mealId).Scan(&updatedMealId)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDataCouldNotBeUpdated
	}
	//TODO: maybe implement a second query, which selects all the required data again from the db for real time data
	return err
}
