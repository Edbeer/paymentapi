package storage

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Edbeer/paymentapi/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func NewPostgresMock() (*PostgresStorage, error) {
	db, err := sql.Open("sqlmock", "")
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		db: db,
	}, err
}

func Test_CreateAccount(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("Create", func(t *testing.T) {
		req := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(req)

		colums := []string{
			"id",
			"first_name",
			"last_name",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"card_security_code",
			"balance", "blocked_money", "statement",
			"created_at",
		}
		rows := sqlmock.NewRows(colums).AddRow(
			account.ID,
			"Pasha1",
			"volkov1",
			444444444444444,
			12,
			24,
			924,
			0,
			0,
			pq.Array(account.Statement),
			account.CreatedAt,
		)
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO account (id, first_name, 
			last_name, card_number, card_expiry_month, 
			card_expiry_year, card_security_code, 
			balance, blocked_money, statement, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				RETURNING *`)).WithArgs(
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
			account.CreatedAt).WillReturnRows(rows)
		createdUser, err := psql.CreateAccount(context.Background(), account)
		require.NoError(t, err)
		require.NotNil(t, createdUser)
		require.Equal(t, createdUser, account)
	})
}

func Test_GetAccount(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("GetAccounts", func(t *testing.T) {
		req1 := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account1 := models.NewAccount(req1)
		colums := []string{
			"id",
			"first_name",
			"last_name",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"card_security_code",
			"balance", "blocked_money", "statement",
			"created_at",
		}
		rows1 := sqlmock.NewRows(colums).AddRow(
			account1.ID,
			"Pasha1",
			"volkov1",
			444444444444444,
			12,
			24,
			924,
			0,
			0,
			pq.Array(account1.Statement),
			account1.CreatedAt,
		)
		req2 := &models.RequestCreate{
			FirstName:        "Pasha",
			LastName:         "volkov",
			CardNumber:       444444444444432,
			CardExpiryMonth:  10,
			CardExpiryYear:   25,
			CardSecurityCode: 934,
		}
		account2 := models.NewAccount(req2)

		rows2 := sqlmock.NewRows(colums).AddRow(
			account2.ID,
			"Pasha1",
			"volkov1",
			444444444444444,
			12,
			24,
			924,
			0,
			0,
			pq.Array(account2.Statement),
			account2.CreatedAt,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM account`)).WillReturnRows(rows1, rows2)
		userList, err := psql.GetAccount(context.Background())
		require.NoError(t, err)
		require.NotNil(t, userList)
	})
}

func Test_UpdateAccount(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("Update", func(t *testing.T) {
		req := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(req)

		colums := []string{
			"id",
			"first_name",
			"last_name",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"card_security_code",
			"balance", "blocked_money", "statement",
			"created_at",
		}

		rows2 := sqlmock.NewRows(colums).AddRow(
			account.ID,
			"Pasha12",
			"volkov12",
			444444444444444,
			12,
			24,
			924,
			0,
			0,
			pq.Array(account.Statement),
			account.CreatedAt,
		)


		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE account
		SET first_name = COALESCE(NULLIF($1, ''), first_name),
			last_name = COALESCE(NULLIF($2, ''), last_name),
			card_number = COALESCE(NULLIF($3, 0), card_number)
		WHERE id = $4
		RETURNING *`)).WithArgs(
			account.FirstName,
			account.LastName,
			account.CardNumber,
			account.ID).WillReturnRows(rows2)

		updatedUser, err := psql.UpdateAccount(context.Background(), account, account.ID)
		require.NoError(t, err)
		require.NotEqual(t, updatedUser, account)
	})
}

func Test_DeleteAccount(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("Delete", func(t *testing.T) {
		uid := uuid.New()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM account WHERE id = $1`)).WithArgs(uid).WillReturnResult(sqlmock.NewResult(1, 1))

		err := psql.DeleteAccount(context.Background(), uid)
		require.NoError(t, err)
	})
}

func Test_GetAccountByID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("GetAccountByID", func(t *testing.T) {
		req := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(req)
		colums := []string{
			"id",
			"first_name",
			"last_name",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"card_security_code",
			"balance", "blocked_money", "statement",
			"created_at",
		}
		rows := sqlmock.NewRows(colums).AddRow(
			account.ID,
			"Pasha1",
			"volkov1",
			444444444444444,
			12,
			24,
			924,
			0,
			0,
			pq.Array(account.Statement),
			account.CreatedAt,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM account WHERE id = $1`)).WithArgs(account.ID).WillReturnRows(rows)
		acc, err := psql.GetAccountByID(context.Background(), account.ID)
		require.NoError(t, err)
		require.NotNil(t, acc)
	})
}

func Test_GetAccountByCard(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("GetAccountByCard", func(t *testing.T) {
		req := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(req)
		colums := []string{
			"id",
			"first_name",
			"last_name",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"card_security_code",
			"balance", "blocked_money", "statement",
			"created_at",
		}
		rows := sqlmock.NewRows(colums).AddRow(
			account.ID,
			"Pasha1",
			"volkov1",
			444444444444444,
			12,
			24,
			924,
			0,
			0,
			pq.Array(account.Statement),
			account.CreatedAt,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM account WHERE card_number = $1`)).WithArgs(account.CardNumber).WillReturnRows(rows)
		acc, err := psql.GetAccountByCard(context.Background(), account.CardNumber)
		require.NoError(t, err)
		require.NotNil(t, acc)
		require.Equal(t, acc.ID, account.ID)
	})
}

func Test_DepositAccount(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("Daposit", func(t *testing.T) {
		req := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(req)

		reqDep := &models.RequestDeposit{
			ID:         account.ID,
			CardNumber: 444444444444444,
			Balance:    50,
		}

		colums := []string{
			"id",
			"first_name",
			"last_name",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"card_security_code",
			"balance", "blocked_money", "statement",
			"created_at",
		}
		rows := sqlmock.NewRows(colums).AddRow(
			account.ID,
			"Pasha1",
			"volkov1",
			444444444444444,
			12,
			24,
			924,
			50,
			0,
			pq.Array(account.Statement),
			account.CreatedAt,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE account
		SET balance = COALESCE(NULLIF($1, 0), balance)
		WHERE id = $2 and card_number = $3
		RETURNING *`)).WithArgs(uint64(50), account.ID, account.CardNumber).WillReturnRows(rows)
		acc, err := psql.DepositAccount(context.Background(), reqDep)
		require.NoError(t, err)
		require.Equal(t, acc.CardNumber, account.CardNumber)
		require.Equal(t, acc.Balance, uint64(50))
	})
}

func Test_SavePayment(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("SavePayment", func(t *testing.T) {
		reqAcc := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(reqAcc)

		reqMer := &models.RequestCreate{
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       444444444444434,
			CardExpiryMonth:  10,
			CardExpiryYear:   22,
			CardSecurityCode: 934,
		}
		merchant := models.NewAccount(reqMer)

		payReq := &models.PaymentRequest{
			AccountId: account.ID,
			OrderId: "1",
			Amount: 50,
			Currency: "RUB",
			CardNumber: 444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		payment := models.CreateAuthPayment(payReq, account, merchant, "")

		colums := []string{
			"id",
			"business_id",
			"order_id",
			"operation",
			"amount",
			"status",
			"currency",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"created_at",
		}
		rows := sqlmock.NewRows(colums).AddRow(
			payment.ID,
			merchant.ID,
			"1",
			payment.Operation,
			50,
			"",
			"RUB",
			444444444444444,
			12,
			24,
			payment.CreatedAt,
		)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO payment (id, business_id, 
			order_id, operation, amount, status, 
			currency, card_number, card_expiry_month,
			 card_expiry_year, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				RETURNING *`)).WithArgs(payment.ID,
					payment.BusinessId,
					payment.OrderId,
					payment.Operation,
					payment.Amount,
					payment.Status,
					payment.Currency,
					payment.CardNumber,
					payment.CardExpiryMonth,
					payment.CardExpiryYear,
					payment.CreatedAt,).WillReturnRows(rows)
		mock.ExpectCommit()

		pay, err := psql.SavePayment(context.Background(), payment)
		require.NoError(t, err)
		require.NotNil(t, pay)
	})
}

func Test_SaveBalance(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("SaveBalance", func(t *testing.T) {
		req := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(req)

		colums := []string{
			"id",
			"first_name",
			"last_name",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"card_security_code",
			"balance", "blocked_money", "statement",
			"created_at",
		}
		rows := sqlmock.NewRows(colums).AddRow(
			account.ID,
			"Pasha1",
			"volkov1",
			444444444444444,
			12,
			24,
			924,
			50,
			50,
			pq.Array(account.Statement),
			account.CreatedAt,
		)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE account
		SET balance = COALESCE(NULLIF($1, 0), balance),
			blocked_money = COALESCE(NULLIF($2, 0), blocked_money)
		WHERE id = $3
		RETURNING *`)).WithArgs(50, 50, account.ID).WillReturnRows(rows)
		mock.ExpectCommit()
		acc, err := psql.SaveBalance(context.Background(), account, 50, 50)
		require.NoError(t, err)
		require.NotNil(t, acc)
	})
}

func Test_GetPaymentByID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	psql := NewPostgresStorage(db)

	t.Run("GetPaymentByID", func(t *testing.T) {
		reqAcc := &models.RequestCreate{
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		account := models.NewAccount(reqAcc)

		reqMer := &models.RequestCreate{
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       444444444444434,
			CardExpiryMonth:  10,
			CardExpiryYear:   22,
			CardSecurityCode: 934,
		}
		merchant := models.NewAccount(reqMer)

		payReq := &models.PaymentRequest{
			AccountId: account.ID,
			OrderId: "1",
			Amount: 50,
			Currency: "RUB",
			CardNumber: 444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
		}
		payment := models.CreateAuthPayment(payReq, account, merchant, "")

		colums := []string{
			"id",
			"business_id",
			"order_id",
			"operation",
			"amount",
			"status",
			"currency",
			"card_number",
			"card_expiry_month",
			"card_expiry_year",
			"created_at",
		}
		rows := sqlmock.NewRows(colums).AddRow(
			payment.ID,
			merchant.ID,
			"1",
			payment.Operation,
			50,
			"",
			"RUB",
			444444444444444,
			12,
			24,
			payment.CreatedAt,
		)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM payment WHERE id = $1`)).WithArgs(payment.ID).WillReturnRows(rows)
		mock.ExpectCommit()
		pay, err := psql.GetPaymentByID(context.Background(), payment.ID)
		require.NoError(t, err)
		require.NotNil(t, pay)
	})
}