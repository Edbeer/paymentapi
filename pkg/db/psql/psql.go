package psql

import (
	"database/sql"
	"fmt"

	"github.com/Edbeer/paymentapi/config"
)


func NewPostgresDB(config *config.Config) (*sql.DB, error) {
	connString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable", 
		config.Postgres.PostgresqlHost,  config.Postgres.PostgresqlUser, 
		config.Postgres.PostgresqlPassword, config.Postgres.PostgresqlDbname,
	)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}