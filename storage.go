package main

import (
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(account *Account) (*Account, error)
	GetAccount() ([]*Account, error)
	GetAccountByID(id uuid.UUID) (*Account, error)
	UpdateAccount(firstName string, lastName string, cardNumber uint64, id uuid.UUID) (*Account, error)
	DeleteAccount(id uuid.UUID) error
}

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

func (s *PostgresStorage) InitTables() error {
	if err := s.CreateAccountTable(); err != nil {
		return err
	}
	if err := s.CreatePaymentTable(); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) CreateAccountTable() error {
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

func (s *PostgresStorage) CreatePaymentTable() error {
	query := `CREATE TABLE IF NOT EXISTS payment 
	(
		id UUID,
		business_id UUID PRIMARY KEY,
		foreign key (business_id) references account (id),
		order_id serial,
		operation VARCHAR(50),
		amount serial,
		status VARCHAR(50),
		description VARCHAR(50),
		card_number serial,
		created_at TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStorage) CreateAccount(account *Account) (*Account, error) {
	query := `INSERT INTO account (id, first_name, last_name, card_number, balance, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING *`
	acc := &Account{}
	if err := s.db.QueryRow(
		query, 
		account.ID, 
		account.FirstName, 
		account.LastName, 
		account.CardNumber, 
		account.Balance, 
		account.CreatedAt,
	).Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.CardNumber, &acc.Balance, &acc.CreatedAt); err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *PostgresStorage) GetAccount() ([]*Account, error) {
	query := `SELECT * FROM account`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	acc := &Account{}
	for rows.Next() {
		if err := rows.Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.CardNumber, &acc.Balance, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (s *PostgresStorage) GetAccountByID(id uuid.UUID) (*Account, error) {
	query := `SELECT * FROM account WHERE id = $1`
	acc := &Account{}
	if err := s.db.QueryRow(
		query, id,
	).Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.CardNumber, &acc.Balance, &acc.CreatedAt); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) UpdateAccount(firstName string, lastName string, cardNumber uint64, id uuid.UUID) (*Account, error) {
	query := `UPDATE account
				SET first_name = COALESCE(NULLIF($1, ''), first_name),
					last_name = COALESCE(NULLIF($2, ''), last_name),
					card_number = COALESCE(NULLIF($3, 0), card_number)
				WHERE id = $4
				RETURNING *`
	acc := &Account{}
	if err := s.db.QueryRow(
		query,
		firstName,
		lastName,
		cardNumber,
		id,
	).Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.CardNumber, &acc.Balance, &acc.CreatedAt); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) DeleteAccount(id uuid.UUID) error {
	query := `DELETE FROM account WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *PostgresStorage) CreatePayment() (*Payment, error) {
	return nil, nil
}