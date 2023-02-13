package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mockstore "github.com/Edbeer/paymentapi/api/mock"
	"github.com/Edbeer/paymentapi/config"
	"github.com/Edbeer/paymentapi/pkg/utils"
	"github.com/Edbeer/paymentapi/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
)

func Test_CreatePayment(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()


	mockStorage := mockstore.NewMockStorage(ctrl)
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)

	uid := uuid.New()
	reqPay := &types.PaymentRequest{
		AccountId:        uid,
		OrderId:          "1",
		Amount:           50,
		Currency:         "RUB",
		CardNumber:       "4444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
	}
	err = utils.ValidatePaymentRequest(reqPay)
	require.NoError(t, err)
	buffer, err := utils.AnyToBytesBuffer(reqPay)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/payment/auth", buffer)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Payment.createPayment")
	defer span.Finish()
	
	recorder := httptest.NewRecorder()

	t.Run("CreatePayment", func(t *testing.T) {
		// account
		account := &types.Account{
			ID:               uid,
			FirstName:        "Pavel",
			LastName:         "Voklov",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          50,
			BlockedMoney:     0,
			Statement:        []string{},
			CreatedAt:        time.Now(),
		}

		// merchant
		mid := uuid.New()
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          0,
			BlockedMoney:     0,
			Statement:        []string{},
			CreatedAt:        time.Now(),
		}
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, uid).Return(account, nil).AnyTimes()
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		account.Balance = account.Balance - reqPay.Amount
		account.BlockedMoney = account.BlockedMoney + reqPay.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, account, account.Balance, account.BlockedMoney).Return(account, nil).AnyTimes()
		merchant.BlockedMoney = merchant.BlockedMoney + reqPay.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, merchant, merchant.Balance, merchant.BlockedMoney).Return(merchant, nil).AnyTimes()

		
		payment := types.CreateAuthPayment(reqPay, account, merchant, "Approved")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, payment).Return(payment, nil).AnyTimes()

		merchant.Statement = append(merchant.Statement, payment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, payment.ID).Return(merchant, nil).AnyTimes()

		account.Statement = append(account.Statement, payment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, uid, payment.ID).Return(account, nil).AnyTimes()

		err = server.createPayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, payment.ID.String())
		require.Equal(t, merchant.Balance, uint64(0))
		require.Equal(t, merchant.BlockedMoney, uint64(50))
		require.Equal(t, account.Balance, uint64(0))
		require.Equal(t, account.BlockedMoney, uint64(50))
	})

	t.Run("Wrong payment request", func(t *testing.T) {
		// account
		account := &types.Account{
			ID:               uid,
			FirstName:        "Pavel",
			LastName:         "Voklov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          50,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		}

		// merchant
		mid := uuid.New()
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444234",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		}
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, uid).Return(account, nil).AnyTimes()
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		
		tx, _ := db.BeginTx(ctxWithTrace, nil)
		payment := types.CreateAuthPayment(reqPay, account, merchant, "wrong payment request")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, payment).Return(payment, nil).AnyTimes()
		
		merchant.Statement = append(merchant.Statement, payment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, payment.ID).Return(merchant, nil).AnyTimes()

		err = server.createPayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.NotEqual(t, account.CardNumber, reqPay.CardNumber)
	})

	t.Run("Insufficient funds", func(t *testing.T) {
		// account
		account := &types.Account{
			ID:               uid,
			FirstName:        "Pavel",
			LastName:         "Voklov",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          30,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		}

		// merchant
		mid := uuid.New()
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444234",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		}

		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, uid).Return(account, nil).AnyTimes()
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		require.Less(t, account.Balance, reqPay.Amount)

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		payment := types.CreateAuthPayment(reqPay, account, merchant, "wrong payment request")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, payment).Return(payment, nil).AnyTimes()

		merchant.Statement = append(merchant.Statement, payment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, payment.ID).Return(merchant, nil).AnyTimes()

		err = server.createPayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, payment.ID.String())
		require.Equal(t, merchant.BlockedMoney, uint64(0))
	})
}

func Test_CapturePayment(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)
	pid := uuid.New()
	reqPaid := &types.PaidRequest{
		OrderId:   "1",
		PaymentId: pid,
		Operation: "Capture",
		Amount:    50,
	}

	buffer, err := utils.AnyToBytesBuffer(reqPaid)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/payment/capture/{id}", buffer)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Payment.capturePayment")
	defer span.Finish()

	recorder := httptest.NewRecorder()

	t.Run("CapturePayment", func(t *testing.T) {
		mid := uuid.New()
		refPayment := &types.Payment{
			ID:              pid,
			BusinessId:      mid,
			OrderId:         "1",
			Operation:       "Authorization",
			Amount:          50,
			Status:          "Approved",
			Currency:        "RUB",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CreatedAt:       time.Time{},
		}
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          0,
			BlockedMoney:     50,
			Statement:        []string{pid.String()},
			CreatedAt:        time.Now(),
		}
		uid := uuid.New()
		account := &types.Account{
			ID:               uid,
			FirstName:        "Pavel",
			LastName:         "Voklov",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          0,
			BlockedMoney:     50,
			Statement:        []string{pid.String()},
			CreatedAt:        time.Now(),
		}

		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		mockStorage.EXPECT().GetPaymentByID(ctxWithTrace, pid).Return(refPayment, nil).AnyTimes()

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		refPayment.Amount = refPayment.Amount - reqPaid.Amount
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, refPayment).Return(refPayment, nil).AnyTimes()

		completedPayment := types.CreateCompletePayment(reqPaid, refPayment, "Successful payment")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, completedPayment).Return(completedPayment, nil).AnyTimes()

		mockStorage.EXPECT().GetAccountByCard(ctxWithTrace, refPayment.CardNumber).Return(account, nil).AnyTimes()
		account.BlockedMoney = account.BlockedMoney - reqPaid.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, account, account.Balance, account.BlockedMoney).Return(account, nil).AnyTimes()
		account.Statement = append(account.Statement, completedPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, uid, completedPayment.ID).Return(account, nil).AnyTimes()

		merchant.Balance = merchant.Balance + reqPaid.Amount
		merchant.BlockedMoney = merchant.BlockedMoney - reqPaid.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, merchant, merchant.Balance, merchant.BlockedMoney).Return(merchant, nil).AnyTimes()
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, completedPayment.ID).Return(merchant, nil).AnyTimes()

		err := server.capturePayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, completedPayment.ID.String())
		require.Equal(t, merchant.BlockedMoney, uint64(0))
		require.Equal(t, merchant.Balance, uint64(50))
		require.Equal(t, account.Balance, uint64(0))
	})

	t.Run("Invalid amount", func(t *testing.T) {
		mid := uuid.New()
		refPayment := &types.Payment{
			ID:              pid,
			BusinessId:      mid,
			OrderId:         "1",
			Operation:       "Authorization",
			Amount:          50,
			Status:          "Approved",
			Currency:        "RUB",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CreatedAt:       time.Time{},
		}
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          0,
			BlockedMoney:     50,
			Statement:        []string{pid.String()},
			CreatedAt:        time.Now(),
		}

		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		mockStorage.EXPECT().GetPaymentByID(ctxWithTrace, pid).Return(refPayment, nil).AnyTimes()

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		invalidPayment := types.CreateCompletePayment(reqPaid, refPayment, "Invalid amount")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, invalidPayment).Return(invalidPayment, nil).AnyTimes()

		merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, invalidPayment.ID).Return(merchant, nil).AnyTimes()

		err := server.capturePayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, invalidPayment.ID.String())
		require.Equal(t, merchant.BlockedMoney, uint64(50))
	})
}

func Test_RefundPayment(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)
	pid := uuid.New()
	reqPaid := &types.PaidRequest{
		OrderId:   "1",
		PaymentId: pid,
		Operation: "Refund",
		Amount:    50,
	}

	buffer, err := utils.AnyToBytesBuffer(reqPaid)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/payment/refund/{id}", buffer)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Payment.refundPayment")
	defer span.Finish()
	recorder := httptest.NewRecorder()

	t.Run("Refund", func(t *testing.T) {
		mid := uuid.New()
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          50,
			BlockedMoney:     0,
			Statement:        []string{},
			CreatedAt:        time.Now(),
		}

		uid := uuid.New()
		account := &types.Account{
			ID:               uid,
			FirstName:        "Pavel",
			LastName:         "Voklov",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          0,
			BlockedMoney:     0,
			Statement:        []string{},
			CreatedAt:        time.Now(),
		}

		refPayment := &types.Payment{
			ID:              uuid.New(),
			BusinessId:      mid,
			OrderId:         "1",
			Operation:       "Capture",
			Amount:          50,
			Status:          "Approved",
			Currency:        "RUB",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CreatedAt:       time.Time{},
		}
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		mockStorage.EXPECT().GetPaymentByID(ctxWithTrace, pid).Return(refPayment, nil).AnyTimes()

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		refPayment.Amount = refPayment.Amount - reqPaid.Amount
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, refPayment).Return(refPayment, nil).AnyTimes()

		completedPayment := types.CreateCompletePayment(reqPaid, refPayment, "Successful refund")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, completedPayment).Return(completedPayment, nil).AnyTimes()

		mockStorage.EXPECT().GetAccountByCard(ctxWithTrace, refPayment.CardNumber).Return(account, nil).AnyTimes()
		account.Balance = account.Balance + reqPaid.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, account, account.Balance, account.BlockedMoney).Return(account, nil).AnyTimes()
		account.Statement = append(account.Statement, completedPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, uid, completedPayment.ID).Return(account, nil).AnyTimes()

		merchant.Balance = merchant.Balance - reqPaid.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, merchant, merchant.Balance, merchant.BlockedMoney).Return(merchant, nil).AnyTimes()
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, completedPayment.ID).Return(merchant, nil).AnyTimes()

		err := server.refundPayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, completedPayment.ID.String())
		require.Equal(t, merchant.Balance, uint64(0))
		require.Equal(t, account.Balance, uint64(50))
	})

	t.Run("Invalid amount", func(t *testing.T) {
		mid := uuid.New()
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          50,
			BlockedMoney:     0,
			Statement:        []string{},
			CreatedAt:        time.Now(),
		}

		refPayment := &types.Payment{
			ID:              uuid.New(),
			BusinessId:      mid,
			OrderId:         "1",
			Operation:       "Authorization",
			Amount:          50,
			Status:          "Approved",
			Currency:        "RUB",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CreatedAt:       time.Time{},
		}
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		mockStorage.EXPECT().GetPaymentByID(ctxWithTrace, pid).Return(refPayment, nil).AnyTimes()

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		invalidPayment := types.CreateCompletePayment(reqPaid, refPayment, "Invalid amount")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, invalidPayment).Return(invalidPayment, nil).AnyTimes()

		merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, invalidPayment.ID).Return(merchant, nil).AnyTimes()

		err := server.refundPayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, invalidPayment.ID.String())
		require.Equal(t, merchant.Balance, uint64(50))
	})
}

func Test_CancelPaymen(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)
	pid := uuid.New()
	reqPaid := &types.PaidRequest{
		OrderId:   "1",
		PaymentId: pid,
		Operation: "Cancel",
		Amount:    50,
	}

	buffer, err := utils.AnyToBytesBuffer(reqPaid)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/payment/cancel/{id}", buffer)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Payment.cancelPayment")
	defer span.Finish()
	recorder := httptest.NewRecorder()

	t.Run("Cancel", func(t *testing.T) {
		mid := uuid.New()
		refPayment := &types.Payment{
			ID:              pid,
			BusinessId:      mid,
			OrderId:         "1",
			Operation:       "Authorization",
			Amount:          50,
			Status:          "Approved",
			Currency:        "RUB",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CreatedAt:       time.Time{},
		}
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          0,
			BlockedMoney:     50,
			Statement:        []string{pid.String()},
			CreatedAt:        time.Now(),
		}
		uid := uuid.New()
		account := &types.Account{
			ID:               uid,
			FirstName:        "Pavel",
			LastName:         "Voklov",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          0,
			BlockedMoney:     50,
			Statement:        []string{},
			CreatedAt:        time.Now(),
		}

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		mockStorage.EXPECT().GetPaymentByID(ctxWithTrace, pid).Return(refPayment, nil).AnyTimes()

		refPayment.Amount = refPayment.Amount - reqPaid.Amount
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, refPayment).Return(refPayment, nil).AnyTimes()

		completedPayment := types.CreateCompletePayment(reqPaid, refPayment, "Successful cancel")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, completedPayment).Return(completedPayment, nil).AnyTimes()
		
		mockStorage.EXPECT().GetAccountByCard(ctxWithTrace, refPayment.CardNumber).Return(account, nil).AnyTimes()
		account.Balance = account.Balance + reqPaid.Amount
		account.BlockedMoney = account.BlockedMoney - reqPaid.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, account, account.Balance, account.BlockedMoney).Return(account, nil).AnyTimes()
		account.Statement = append(account.Statement, completedPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, uid, completedPayment.ID).Return(account, nil).AnyTimes()

		merchant.BlockedMoney = merchant.BlockedMoney - reqPaid.Amount
		mockStorage.EXPECT().SaveBalance(ctxWithTrace, tx, merchant, merchant.Balance, merchant.BlockedMoney).Return(account, nil).AnyTimes()
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, completedPayment.ID).Return(merchant, nil).AnyTimes()

		err := server.cancelPayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, completedPayment.ID.String())
		require.Equal(t, merchant.BlockedMoney, uint64(0))
		require.Equal(t, account.Balance, uint64(50))
	})

	t.Run("Invalid amount", func(t *testing.T) {
		mid := uuid.New()
		merchant := &types.Account{
			ID:               mid,
			FirstName:        "Pasha",
			LastName:         "Volkov",
			CardNumber:       "444444444444434",
			CardExpiryMonth:  "10",
			CardExpiryYear:   "24",
			CardSecurityCode: "934",
			Balance:          0,
			BlockedMoney:     50,
			Statement:        []string{},
			CreatedAt:        time.Now(),
		}

		refPayment := &types.Payment{
			ID:              uuid.New(),
			BusinessId:      mid,
			OrderId:         "1",
			Operation:       "Authorization",
			Amount:          50,
			Status:          "Approved",
			Currency:        "RUB",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CreatedAt:       time.Time{},
		}
		mockStorage.EXPECT().GetAccountByID(ctxWithTrace, mid).Return(merchant, nil).AnyTimes()
		mockStorage.EXPECT().GetPaymentByID(ctxWithTrace, pid).Return(refPayment, nil).AnyTimes()

		tx, _ := db.BeginTx(ctxWithTrace, nil)
		invalidPayment := types.CreateCompletePayment(reqPaid, refPayment, "Invalid amount")
		mockStorage.EXPECT().SavePayment(ctxWithTrace, tx, invalidPayment).Return(invalidPayment, nil).AnyTimes()

		merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
		mockStorage.EXPECT().UpdateStatement(ctxWithTrace, tx, mid, invalidPayment.ID).Return(merchant, nil).AnyTimes()

		err := server.refundPayment(recorder, request)
		require.NoError(t, err)
		require.Nil(t, err)
		require.Contains(t, merchant.Statement, invalidPayment.ID.String())
		require.Equal(t, merchant.BlockedMoney, uint64(50))
	})
}
