package postgres

import (
	"context"
	"database/sql"

	"github.com/Edbeer/paymentapi/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		db: db,
	}
}

func (s *PostgresStorage) CreateAccount(ctx context.Context, reqAcc *types.RequestCreate) (*types.Account, error) {
	query := `INSERT INTO account (first_name, 
		last_name, card_number, card_expiry_month, 
		card_expiry_year, card_security_code, 
		balance, blocked_money, statement, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
			RETURNING *`
	req := types.NewAccount(reqAcc)
	acc := &types.Account{}
	if err := s.db.QueryRowContext(
		ctx, query,
		req.FirstName,
		req.LastName,
		req.CardNumber,
		req.CardExpiryMonth,
		req.CardExpiryYear,
		req.CardSecurityCode,
		req.Balance,
		req.BlockedMoney,
		pq.Array(req.Statement),
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

func (s *PostgresStorage) GetAccount(ctx context.Context) ([]*types.Account, error) {
	query := `SELECT * FROM account`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	accounts := []*types.Account{}
	for rows.Next() {
		acc := &types.Account{}
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

func (s *PostgresStorage) GetAccountByID(ctx context.Context, id uuid.UUID) (*types.Account, error) {
	query := `SELECT * FROM account 
			WHERE id = $1`
	acc := &types.Account{}

	if err := s.db.QueryRowContext(
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

func (s *PostgresStorage) GetAccountByCard(ctx context.Context, card string) (*types.Account, error) {
	query := `SELECT * FROM account WHERE card_number = $1`
	acc := &types.Account{}
	if err := s.db.QueryRowContext(
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

func (s *PostgresStorage) UpdateAccount(ctx context.Context, reqUp *types.RequestUpdate, id uuid.UUID) (*types.Account, error) {
	query := `UPDATE account
	SET first_name = COALESCE(NULLIF($1, ''), first_name),
		last_name = COALESCE(NULLIF($2, ''), last_name),
		card_number = COALESCE(NULLIF($3, ''), card_number),
		card_expiry_month = COALESCE(NULLIF($4, ''), card_expiry_month),
		card_expiry_year = COALESCE(NULLIF($5, ''), card_expiry_year),
		card_security_code = COALESCE(NULLIF($6, ''), card_security_code)
	WHERE id = $7
	RETURNING *`
	acc := &types.Account{}

	if err := s.db.QueryRowContext(
		ctx, query,
		reqUp.FirstName,
		reqUp.LastName,
		reqUp.CardNumber,
		reqUp.CardExpiryMonth,
		reqUp.CardExpiryYear,
		reqUp.CardSecurityCode,
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
	return acc, nil
}

func (s *PostgresStorage) UpdateStatement(ctx context.Context, tx *sql.Tx, id, paymentId uuid.UUID) (*types.Account, error) {
	// TODO change id on card_number
	query := `UPDATE account
				SET statement = array_append(statement, $1)
				WHERE id = $2
				RETURNING *`
	acc := &types.Account{}
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
	return acc, nil
}

func (s *PostgresStorage) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM account WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) DepositAccount(ctx context.Context, reqDep *types.RequestDeposit) (*types.Account, error) {
	query := `UPDATE account
				SET balance = COALESCE(NULLIF($1, 0), balance)
				WHERE card_number = $2
				RETURNING *`
	acc := &types.Account{}

	if err := s.db.QueryRowContext(
		ctx,
		query,
		reqDep.Balance,
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
	return acc, nil
}

func (s *PostgresStorage) GetAccountStatement(ctx context.Context, id uuid.UUID) ([]string, error) {
	query := `SELECT * FROM account WHERE id = $1`
	acc := &types.Account{}

	if err := s.db.QueryRowContext(
		ctx,
		query,
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
	return acc.Statement, nil

}

func (s *PostgresStorage) SavePayment(ctx context.Context, tx *sql.Tx, payment *types.Payment) (*types.Payment, error) {
	query := `INSERT INTO payment (id, business_id, 
		order_id, operation, amount, status, 
		currency, card_number, card_expiry_month,
		 card_expiry_year, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING *`
	pay := &types.Payment{}
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
	return pay, nil
}

func (s *PostgresStorage) GetPaymentByID(ctx context.Context, id uuid.UUID) (*types.Payment, error) {
	query := `SELECT * FROM payment WHERE id = $1`
	pay := &types.Payment{}
	if err := s.db.QueryRowContext(
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

func (s *PostgresStorage) SaveBalance(ctx context.Context, tx *sql.Tx, account *types.Account, balance, bmoney uint64) (*types.Account, error) {
	query := `UPDATE account
				SET balance = COALESCE(NULLIF($1, 0), balance),
					blocked_money = COALESCE(NULLIF($2, 0), blocked_money)
				WHERE id = $3
				RETURNING *`
	acc := &types.Account{}
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
	return acc, nil
}
