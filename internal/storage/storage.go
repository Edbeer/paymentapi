package storage

import (
	"context"
	"database/sql"

	"github.com/Edbeer/paymentapi/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connString := "host=paymentdb user=postgres password=postgres dbname=paymentdb sslmode=disable"
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

func (s *PostgresStorage) CreateAccount(ctx context.Context, account *models.Account) (*models.Account, error) {
	query := `INSERT INTO account (id, first_name, 
		last_name, card_number, card_expiry_month, 
		card_expiry_year, card_security_code, 
		balance, blocked_money, statement, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING *`
	acc := &models.Account{}
	if err := s.db.QueryRowContext(
		ctx, query,
		account.ID,
		account.FirstName,
		account.LastName,
		account.CardNumber,
		account.CardExpiryMonth,
		account.CardExpiryYear,
		account.CardSecurityCode,
		account.Balance,
		account.BlockedMoney,
		pq.Array(account.Statement),
		account.CreatedAt,
	).Scan(
		&acc.ID, &acc.FirstName,
		&acc.LastName, &acc.CardNumber,
		&acc.CardExpiryMonth, &acc.CardExpiryYear,
		&acc.CardSecurityCode, &acc.Balance,
		&acc.BlockedMoney, pq.Array(&acc.Statement),
		&acc.CreatedAt,
	); err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *PostgresStorage) GetAccount(ctx context.Context) ([]*models.Account, error) {
	query := `SELECT * FROM account`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	accounts := []*models.Account{}
	acc := &models.Account{}
	for rows.Next() {
		if err := rows.Scan(
			&acc.ID, &acc.FirstName,
			&acc.LastName, &acc.CardNumber,
			&acc.CardExpiryMonth, &acc.CardExpiryYear,
			&acc.CardSecurityCode, &acc.Balance,
			&acc.BlockedMoney, pq.Array(&acc.Statement),
			&acc.CreatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (s *PostgresStorage) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	query := `SELECT * FROM account 
			WHERE id = $1`
	acc := &models.Account{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx, query, id,
	).Scan(
		&acc.ID, &acc.FirstName,
		&acc.LastName, &acc.CardNumber,
		&acc.CardExpiryMonth, &acc.CardExpiryYear,
		&acc.CardSecurityCode, &acc.Balance,
		&acc.BlockedMoney, pq.Array(&acc.Statement),
		&acc.CreatedAt,
	); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) GetAccountByCard(ctx context.Context, card uint64) (*models.Account, error) {
	query := `SELECT * FROM account 
			WHERE card_number = $1`
	acc := &models.Account{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx, query, card,
	).Scan(
		&acc.ID, &acc.FirstName,
		&acc.LastName, &acc.CardNumber,
		&acc.CardExpiryMonth, &acc.CardExpiryYear,
		&acc.CardSecurityCode, &acc.Balance,
		&acc.BlockedMoney, pq.Array(&acc.Statement),
		&acc.CreatedAt,
	); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) UpdateAccount(ctx context.Context, reqUp *models.Account, id uuid.UUID) (*models.Account, error) {
	query := `UPDATE account
				SET first_name = COALESCE(NULLIF($1, ''), first_name),
					last_name = COALESCE(NULLIF($2, ''), last_name),
					card_number = COALESCE(NULLIF($3, 0), card_number)
				WHERE id = $4
				RETURNING *`
	acc := &models.Account{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx, query,
		reqUp.FirstName,
		reqUp.LastName,
		reqUp.CardNumber,
		id,
	).Scan(
		&acc.ID, &acc.FirstName,
		&acc.LastName, &acc.CardNumber,
		&acc.CardExpiryMonth, &acc.CardExpiryYear,
		&acc.CardSecurityCode, &acc.Balance,
		&acc.BlockedMoney, pq.Array(&acc.Statement),
		&acc.CreatedAt,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) UpdateStatement(ctx context.Context, id, paymentId uuid.UUID) (*models.Account, error) {
	// TODO change id on card_number
	query := `UPDATE account
				SET statement = array_append(statement, $1)
				WHERE id = $2
				RETURNING *`
	acc := &models.Account{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx,
		query,
		paymentId.String(),
		id,
	).Scan(
		&acc.ID, &acc.FirstName,
		&acc.LastName, &acc.CardNumber,
		&acc.CardExpiryMonth, &acc.CardExpiryYear,
		&acc.CardSecurityCode, &acc.Balance,
		&acc.BlockedMoney, pq.Array(&acc.Statement),
		&acc.CreatedAt,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM account WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *PostgresStorage) DepositAccount(ctx context.Context, reqDep *models.RequestDeposit) (*models.Account, error) {
	query := `UPDATE account
				SET balance = COALESCE(NULLIF($1, 0), balance)
				WHERE id = $2 and card_number = $3
				RETURNING *`
	acc := &models.Account{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx,
		query,
		reqDep.Balance,
		reqDep.ID,
		reqDep.CardNumber,
	).Scan(
		&acc.ID, &acc.FirstName,
		&acc.LastName, &acc.CardNumber,
		&acc.CardExpiryMonth, &acc.CardExpiryYear,
		&acc.CardSecurityCode, &acc.Balance,
		&acc.BlockedMoney, pq.Array(&acc.Statement),
		&acc.CreatedAt,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PostgresStorage) SavePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
	query := `INSERT INTO payment (id, business_id, 
		order_id, operation, amount, status, 
		currency, card_number, card_expiry_month,
		 card_expiry_year, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING *`
	pay := &models.Payment{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx, query,
		payment.ID,
		payment.BusinessId,
		payment.OrderId,
		payment.Operation,
		payment.Amount,
		payment.Status,
		payment.Currency,
		payment.CardNumber,
		payment.CardExpiryMonth,
		payment.CardExpiryYear,
		payment.CreatedAt,
	).Scan(
		&pay.ID, &pay.BusinessId, 
		&pay.OrderId, &pay.Operation, 
		&pay.Amount, &pay.Status, 
		&pay.Currency, &pay.CardNumber,
		&pay.CardExpiryMonth, &pay.CardExpiryYear, 
		&pay.CreatedAt,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return pay, nil
}

func (s *PostgresStorage) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	query := `SELECT * FROM payment WHERE id = $1`
	pay := &models.Payment{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx, query, id,
	).Scan(
		&pay.ID, &pay.BusinessId, 
		&pay.OrderId, &pay.Operation, 
		&pay.Amount, &pay.Status, 
		&pay.Currency, &pay.CardNumber,
		&pay.CardExpiryMonth, &pay.CardExpiryYear, 
		&pay.CreatedAt,
	); err != nil {
		return nil, err
	}
	return pay, nil
}

func (s *PostgresStorage) SaveBalance(ctx context.Context, account *models.Account, balance, bmoney uint64) (*models.Account, error) {
	query := `UPDATE account
				SET balance = COALESCE(NULLIF($1, 0), balance),
					blocked_money = COALESCE(NULLIF($2, 0), blocked_money)
				WHERE id = $3
				RETURNING *`
	acc := &models.Account{}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(
		ctx, query,
		balance,
		bmoney,
		account.ID,
	).Scan(
		&acc.ID, &acc.FirstName,
		&acc.LastName, &acc.CardNumber,
		&acc.CardExpiryMonth, &acc.CardExpiryYear,
		&acc.CardSecurityCode, &acc.Balance,
		&acc.BlockedMoney, pq.Array(&acc.Statement),
		&acc.CreatedAt,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}