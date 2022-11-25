package main

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(ctx context.Context, account *Account) (*Account, error)
	GetAccount(ctx context.Context,) ([]*Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error)
	UpdateAccount(ctx context.Context, reqUp *RequestUpdate, id uuid.UUID) (*Account, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	DepositAccount(ctx context.Context, id uuid.UUID, amount uint64) (*Account, error)
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

func (s *PostgresStorage) InitTables(ctx context.Context) error {
	if err := s.CreateAccountTable(ctx); err != nil {
		return err
	}
	if err := s.CreatePaymentTable(ctx); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) CreateAccountTable(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS account 
	(
		id UUID PRIMARY KEY,
		first_name VARCHAR(50),
		last_name VARCHAR(50),
		card_number serial,
		balance serial,
		created_at TIMESTAMP
	)`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *PostgresStorage) CreatePaymentTable(ctx context.Context) error {
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

	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *PostgresStorage) CreateAccount(ctx context.Context, account *Account) (*Account, error) {
	query := `INSERT INTO account (id, first_name, last_name, card_number, balance, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING *`
	acc := &Account{}
	if err := s.db.QueryRowContext(
		ctx,
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

func (s *PostgresStorage) GetAccount(ctx context.Context) ([]*Account, error) {
	query := `SELECT * FROM account`

	rows, err := s.db.QueryContext(ctx, query)
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

func (s *PostgresStorage) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	query := `SELECT * FROM account WHERE id = $1`
	acc := &Account{}
	if err := s.db.QueryRowContext(
		ctx, query, id,
	).Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.CardNumber, &acc.Balance, &acc.CreatedAt); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) UpdateAccount(ctx context.Context, reqUp *RequestUpdate, id uuid.UUID) (*Account, error) {
	query := `UPDATE account
				SET first_name = COALESCE(NULLIF($1, ''), first_name),
					last_name = COALESCE(NULLIF($2, ''), last_name),
					card_number = COALESCE(NULLIF($3, 0), card_number)
				WHERE id = $4
				RETURNING *`
	acc := &Account{}
	if err := s.db.QueryRowContext(
		ctx,
		query,
		reqUp.FirstName,
		reqUp.LastName,
		reqUp.CardNumber,
		id,
	).Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.CardNumber, &acc.Balance, &acc.CreatedAt); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM account WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *PostgresStorage) DepositAccount(ctx context.Context, id uuid.UUID, amount uint64) (*Account, error) {
	query := `UPDATE account
				SET balance = COALESCE(NULLIF($1, 0), balance)
				WHERE id = $2
				RETURNING *`
	acc := &Account{}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx,
		query,
		amount,
		id,
	).Scan(&acc.ID, &acc.FirstName, &acc.LastName, &acc.CardNumber, &acc.Balance, &acc.CreatedAt); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) CreatePayment() (*Payment, error) {
	return nil, nil
}