package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Edbeer/paymentapi/internal/models"
	mockstore "github.com/Edbeer/paymentapi/internal/storage/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_CreateAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", mockStorage)
	req := &models.RequestCreate{
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       444444444444444,
		CardExpiryMonth:  12,
		CardExpiryYear:   24,
		CardSecurityCode: 924,
	}
	buffer, err := AnyToBytesBuffer(req)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/account", buffer)
	recorder := httptest.NewRecorder()
	reqAcc := models.NewAccount(req)

	mockStorage.EXPECT().CreateAccount(request.Context(), gomock.Eq(req)).Return(&models.Account{
		ID:               reqAcc.ID,
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       444444444444444,
		CardExpiryMonth:  12,
		CardExpiryYear:   24,
		CardSecurityCode: 924,
		Balance:          0,
		BlockedMoney:     0,
		Statement:        reqAcc.Statement,
		CreatedAt:        reqAcc.CreatedAt,
	}, nil).AnyTimes()

	err = server.createAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_GetAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", mockStorage)

	request := httptest.NewRequest(http.MethodGet, "/account", nil)
	recorder := httptest.NewRecorder()

	accounts := []*models.Account{
		{
			ID:               uuid.New(),
			FirstName:        "Pasha",
			LastName:         "volkov",
			CardNumber:       444444444444444,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.New(),
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       444444444444442,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.New(),
			FirstName:        "Pasha12",
			LastName:         "volkov12",
			CardNumber:       444444444444443,
			CardExpiryMonth:  12,
			CardExpiryYear:   24,
			CardSecurityCode: 924,
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		},
	}

	mockStorage.EXPECT().GetAccount(request.Context()).Return(accounts, nil).AnyTimes()

	err := server.getAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_GetAccountByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", mockStorage)
	request := httptest.NewRequest(http.MethodGet, "/accounе/{id}", nil)
	recorder := httptest.NewRecorder()
	uid := uuid.New()

	account := &models.Account{
		ID:               uid,
		FirstName:        "Pasha",
		LastName:         "volkov",
		CardNumber:       444444444444444,
		CardExpiryMonth:  12,
		CardExpiryYear:   24,
		CardSecurityCode: 924,
		Balance:          0,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}

	mockStorage.EXPECT().GetAccountByID(request.Context(), uid).Return(account, nil).AnyTimes()

	err := server.getAccountByID(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_UpdateAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", mockStorage)
	reqUp := &models.RequestUpdate{
		FirstName:  "Pavel",
		CardNumber: 4444444444424323,
	}
	buffer, err := AnyToBytesBuffer(reqUp)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)
	request := httptest.NewRequest(http.MethodPut, "/account/{id}", buffer)
	recorder := httptest.NewRecorder()
	uid := uuid.New()

	account := &models.Account{
		ID:               uid,
		FirstName:        "Pavel",
		LastName:         "volkov",
		CardNumber:       4444444444424323,
		CardExpiryMonth:  12,
		CardExpiryYear:   24,
		CardSecurityCode: 924,
		Balance:          0,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}

	mockStorage.EXPECT().UpdateAccount(request.Context(), reqUp, uid).Return(account, nil).AnyTimes()
	err = server.updateAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_DeleteAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", mockStorage)

	request := httptest.NewRequest(http.MethodDelete, "/account/{id}", nil)
	recorder := httptest.NewRecorder()

	uid := uuid.New()

	mockStorage.EXPECT().DeleteAccount(request.Context(), uid).Return(nil).AnyTimes()

	err := server.deleteAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_DepositAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", mockStorage)
	reqDep := &models.RequestDeposit{
		CardNumber: 4444444444424323,
		Balance: 44,
	}
	buffer, err := AnyToBytesBuffer(reqDep)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)
	request := httptest.NewRequest(http.MethodPost, "/account/deposit", buffer)
	recorder := httptest.NewRecorder()
	uid := uuid.New()
	account := &models.Account{
		ID:               uid,
		FirstName:        "Pavel",
		LastName:         "volkov",
		CardNumber:       4444444444424323,
		CardExpiryMonth:  12,
		CardExpiryYear:   24,
		CardSecurityCode: 924,
		Balance:          0,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}
	account2 := &models.Account{
		ID:               uid,
		FirstName:        "Pavel",
		LastName:         "volkov",
		CardNumber:       4444444444424323,
		CardExpiryMonth:  12,
		CardExpiryYear:   24,
		CardSecurityCode: 924,
		Balance:          reqDep.Balance,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}
	mockStorage.EXPECT().GetAccountByCard(request.Context(), reqDep.CardNumber).Return(account, nil).AnyTimes()
	mockStorage.EXPECT().DepositAccount(request.Context(), reqDep).Return(account2, nil).AnyTimes()

	err = server.depositAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

// Convert bytes to buffer helper
func AnyToBytesBuffer(i interface{}) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(i)
	if err != nil {
		return buf, err
	}
	return buf, nil
}
