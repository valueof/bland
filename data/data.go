package data

import "database/sql"

var db *sql.DB

func ConnectToDB(fp string) (err error) {
	db, err = sql.Open("sqlite3", fp)
	if err != nil {
		return
	}

	return db.Ping()
}
