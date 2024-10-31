package tournament

import (
	"database/sql"
)

func GetTournamentByUUIDFromDB(uuid string, db *sql.DB) (UserFromDB, error) {

	query := `SELECT
    			id
    		FROM
    		    tournament
    		WHERE 
        		uuid = $1`
	row := db.QueryRow(query, uuid)

	var id int
	err := row.Scan(&id)
	return id, err
}
