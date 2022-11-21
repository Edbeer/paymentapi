package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Storage interface {}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connString := "user=postgres password=postgres dbname=paymentdb sslmode=disable"
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{
		db: db,
	}, err
}

func (s *PostgresStorage) InitTable() error {
	return s.CreateTable()
}

func (s *PostgresStorage) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS account 
	(
		id UUID PRIMARY KEY,
		first_name VARCHAR(50),
		last_name VARCHAR(50),
		card_number serial,
		balance serial,
		created_at TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}