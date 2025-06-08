package meal

import (
	"database/sql"
	"enguete/modules/group"
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
            m.group_id,
            m.title,
            m.closed,
            m.fulfilled,
            m.date_time,
            m.meal_type,
            m.notes,
            COUNT(CASE WHEN mp.preference = 'opt-in' OR mp.preference = 'eat later' THEN 1 END) AS participant_count,
            CASE WHEN mc.user_id IS NOT NULL THEN true ELSE false END AS is_cook,
            COALESCE(user_pref.preference, 'undecided') AS user_preference
        FROM meals m
        LEFT JOIN meal_preferences mp ON mp.meal_id = m.meal_id AND mp.deleted_at IS NULL
        LEFT JOIN meal_preferences user_pref ON user_pref.meal_id = m.meal_id AND user_pref.user_id = $2 AND user_pref.deleted_at IS NULL
        LEFT JOIN meal_cooks mc ON mc.meal_id = m.meal_id AND mp.deleted_at IS NULL AND mc.user_id = $2
        WHERE m.meal_id = $1
        GROUP BY m.meal_id, mc.user_id, user_pref.preference
        ORDER BY m.date_time
`
	var mealInformation MealInformation
	err := db.QueryRow(query, mealId, userId).Scan(
		&mealInformation.MealId,
		&mealInformation.GroupId,
		&mealInformation.Title,
		&mealInformation.Closed,
		&mealInformation.Fulfilled,
		&mealInformation.DateTime,
		&mealInformation.MealType,
		&mealInformation.Notes,
		&mealInformation.ParticipantCount,
		&mealInformation.IsCook,
		&mealInformation.UserPreference,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mealInformation, ErrNoData
		}
		return mealInformation, err
	}

	return mealInformation, nil
}

func GetMealParticipationInformationFromDB(mealId string, db *sql.DB) ([]MealPreferences, error) {
	query := `
		SELECT 
    		u.user_id,
    		$1 AS meal_id,
    		ug.user_group_id,
    		mp.preference_id,
    		u.username,
    		COALESCE(mp.preference, 'undecided') AS preference,
    		CASE
        		WHEN mc.user_id IS NOT NULL THEN TRUE
        		ELSE FALSE
    		END AS is_cook
		FROM users u
		LEFT JOIN meal_preferences mp ON u.user_id = mp.user_id AND mp.meal_id = $1
		LEFT JOIN meal_cooks mc ON u.user_id = mc.user_id AND mc.meal_id = $1
		LEFT JOIN meals m ON m.meal_id = $1
		INNER JOIN user_groups ug ON u.user_id = ug.user_id AND ug.group_id = m.group_id
		WHERE mp.meal_id = $1 OR mc.meal_id = $1
		AND (mp.deleted_at IS NULL OR mc.deleted_at IS NULL);
`
	rows, err := db.Query(query, mealId)
	var mealParticipants []MealPreferences
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mealParticipants, nil
		}
		return mealParticipants, err
	}
	defer rows.Close()

	for rows.Next() {
		var mealParticipant MealPreferences
		err := rows.Scan(
			&mealParticipant.UserId,
			&mealParticipant.MealId,
			&mealParticipant.UserGroupId,
			&mealParticipant.PreferenceId,
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

func GetGroupMembersNotParticipatingInMeal(mealId string, groupId string, db *sql.DB) ([]MealPreferences, error) {
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

` //TODO: this needs cleaning up with the meal_cooks
	rows, err := db.Query(query, mealId, groupId)
	var mealParticipants []MealPreferences
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mealParticipants, nil
		}
		return mealParticipants, err
	}
	defer rows.Close()

	for rows.Next() {
		var mealParticipant MealPreferences
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

func GetAllDeletedMealParticipationIds(mealId string, lastRequested *string, db *sql.DB) ([]string, error) {
	query := `
		SELECT
			mp.preference_id
		FROM meal_preferences mp
		LEFT JOIN meal_cooks mc ON mc.meal_id = $1
		WHERE mp.meal_id = $1
		AND (
		    mp.deleted_at IS NOT NULL
	 		AND ($2::timestamp IS NULL OR mp.deleted_at >= $2::timestamp) 
		) 
		AND  (
		    mc.deleted_at IS NOT NULL
		    AND ($2::timestamp IS NULL OR mc.deleted_at >= $2::timestamp)
		)
`

	rows, err := db.Query(query, mealId, lastRequested)
	var deletedIds []string
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return deletedIds, nil
		}
		return deletedIds, err
	}
	defer rows.Close()
	for rows.Next() {
		var preferenceId string
		if err := rows.Scan(&preferenceId); err != nil {
			return nil, err
		}
		deletedIds = append(deletedIds, preferenceId)
	}
	return deletedIds, nil
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
        INSERT INTO meal_preferences (meal_id, user_id, preference, updated_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (meal_id, user_id)
        DO UPDATE 
        SET preference = EXCLUDED.preference,
            updated_at = EXCLUDED.updated_at;`

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

func GetAllMealsInGroupInTimeframe(groupId string, userId string, startDate string, endDate string, db *sql.DB) ([]group.MealCard, error) {
	query := `
        SELECT 
            m.meal_id,
            m.group_id,
            m.title,
            m.closed,
            m.fulfilled,
            m.date_time,
            m.meal_type,
            m.notes,
            COUNT(CASE WHEN mp.preference = 'opt-in' OR mp.preference = 'eat later' THEN 1 END) AS participant_count,
            COALESCE(user_pref.preference, 'undecided') AS user_preference,
            CASE WHEN mc.user_id IS NOT NULL THEN true ELSE false END AS is_cook
        FROM meals m
        LEFT JOIN meal_preferences mp ON mp.meal_id = m.meal_id AND mp.deleted_at IS NULL
        LEFT JOIN meal_preferences user_pref ON user_pref.meal_id = m.meal_id AND user_pref.user_id = $2 AND user_pref.deleted_at IS NULL
        LEFT JOIN meal_cooks mc ON mc.meal_id = m.meal_id AND mc.user_id = $2 AND mc.deleted_at IS NULL
        WHERE m.group_id = $1
    	AND m.deleted_at IS NULL
        AND (m.date_time BETWEEN $3 AND $4)
		
        GROUP BY m.meal_id, user_pref.preference, mc.user_id
        ORDER BY m.date_time desc 
`
	rows, err := db.Query(query, groupId, userId, startDate, endDate)
	var mealCards []group.MealCard
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mealCards, nil
		}
		return mealCards, err
	}

	defer rows.Close()

	for rows.Next() {
		var mealCard group.MealCard
		err := rows.Scan(
			&mealCard.MealId,
			&mealCard.GroupId,
			&mealCard.Title,
			&mealCard.Closed,
			&mealCard.Fulfilled,
			&mealCard.DateTime,
			&mealCard.MealType,
			&mealCard.Notes,
			&mealCard.ParticipantCount,
			&mealCard.UserPreference,
			&mealCard.IsCook,
		)
		if err != nil {
			return mealCards, err
		}
		mealCards = append(mealCards, mealCard)
	}

	return mealCards, nil

}

func GetDeletedMealsInTimeframe(groupId string, startDate string, endDate string, db *sql.DB) ([]string, error) {
	query := `
	SELECT 
		m.meal_id
	FROM meals m
	WHERE m.group_id = $1
	AND m.deleted_at IS NOT NULL
	AND (m.date_time BETWEEN $2 AND $3)
`

	rows, err := db.Query(query, groupId, startDate, endDate)
	var deletedIds []string
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return deletedIds, nil
		}
		return deletedIds, err
	}

	defer rows.Close()

	for rows.Next() {
		var mealId string
		if err := rows.Scan(&mealId); err != nil {
			return nil, err
		}
		deletedIds = append(deletedIds, mealId)
	}

	return deletedIds, nil
}
