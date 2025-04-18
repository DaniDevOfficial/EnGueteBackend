package meal

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"time"
)

//General

func CreateNewMealInDBWithTransaction(newMeal RequestNewMeal, userId string, db *sql.DB) (string, error) {
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
    		u.user_id,
    		$1 AS meal_id,
    		u.username,
    		COALESCE(mp.preference, 'undecided') AS preference,
    		CASE
        		WHEN mc.user_id IS NOT NULL THEN TRUE
        		ELSE FALSE
    		END AS is_cook
		FROM users u
		LEFT JOIN meal_preferences mp ON u.user_id = mp.user_id AND mp.meal_id = $1
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
			&mealParticipant.UserId,
			&mealParticipant.MealId,
			&mealParticipant.Username,
			&mealParticipant.Preference,
			&mealParticipant.IsCook,
		)
		if err != nil {
			return mealParticipants, err
		}
		mealParticipants = append(mealParticipants, mealParticipant)

	}
	return mealParticipants, nil
}

func GetGroupMembersNotParticipatingInMeal(mealId string, groupId string, db *sql.DB) ([]MealParticipant, error) {
	query := `
SELECT 
    u.user_id,
    $1 AS meal_id,
    u.username,
    'undecided' AS preference,
    false AS is_cook
FROM user_groups ug
INNER JOIN users u ON u.user_id = ug.user_id
LEFT JOIN meal_preferences mp 
    ON mp.user_id = u.user_id 
    AND mp.meal_id = $1
LEFT JOIN meal_cooks mc 
    ON mc.user_id = u.user_id 
    AND mc.meal_id = $1
WHERE ug.group_id = $2
  AND mp.user_id IS NULL
	AND mc.user_id IS NULL
GROUP BY u.user_id, u.username;

`
	rows, err := db.Query(query, mealId, groupId)
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
			&mealParticipant.UserId,
			&mealParticipant.MealId,
			&mealParticipant.Username,
			&mealParticipant.Preference,
			&mealParticipant.IsCook,
		)
		if err != nil {
			return mealParticipants, err
		}
		mealParticipants = append(mealParticipants, mealParticipant)

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

func ChangeOptInStatusMealInDB(userId string, mealId string, preference string, db *sql.DB) error {
	query := `
        INSERT INTO meal_preferences (meal_id, user_id, preference, last_updated)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (meal_id, user_id)
        DO UPDATE 
        SET preference = EXCLUDED.preference,
            last_updated = EXCLUDED.last_updated;`

	_, err := db.Exec(query, mealId, userId, preference, time.Now())

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserAlreadyHasAPreferenceInSpecificMeal
		}
		return err
	}

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
