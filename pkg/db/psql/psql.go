package psql

import "database/sql"


func NewPostgresDB() (*sql.DB, error) {
	connString := "host=paydb user=postgres password=postgres dbname=paydb sslmode=disable"
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}