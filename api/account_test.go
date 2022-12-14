package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Edbeer/paymentapi/types"
	"github.com/Edbeer/paymentapi/pkg/utils"
	mockstore "github.com/Edbeer/paymentapi/api/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_CreateAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", db, mockStorage)
	req := &types.RequestCreate{
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       "4444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
	}
	err = utils.ValidateCreateRequest(req)
	require.NoError(t, err)
	buffer, err := utils.AnyToBytesBuffer(req)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/account", buffer)
	recorder := httptest.NewRecorder()
	reqAcc := types.NewAccount(req)
	mockStorage.EXPECT().CreateAccount(request.Context(), gomock.Eq(req)).Return(&types.Account{
		ID:               reqAcc.ID,
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       "4444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
		Balance:          0,
		BlockedMoney:     0,
		Statement:        reqAcc.Statement,
		CreatedAt:        reqAcc.CreatedAt,
	}, nil).AnyTimes()

	err = server.createAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
	require.NotNil(t, reqAcc.ID)
}

func Test_GetAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", db, mockStorage)

	request := httptest.NewRequest(http.MethodGet, "/account", nil)
	recorder := httptest.NewRecorder()

	accounts := []*types.Account{
		{
			ID:               uuid.New(),
			FirstName:        "Pasha",
			LastName:         "volkov",
			CardNumber:       "444444444444444",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.New(),
			FirstName:        "Pasha1",
			LastName:         "volkov1",
			CardNumber:       "444444444444442",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		},
		{
			ID:               uuid.New(),
			FirstName:        "Pasha12",
			LastName:         "volkov12",
			CardNumber:       "444444444444443",
			CardExpiryMonth:  "12",
			CardExpiryYear:   "24",
			CardSecurityCode: "924",
			Balance:          0,
			BlockedMoney:     0,
			Statement:        make([]string, 1),
			CreatedAt:        time.Now(),
		},
	}

	mockStorage.EXPECT().GetAccount(request.Context()).Return(accounts, nil).AnyTimes()

	err = server.getAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_GetAccountByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", db, mockStorage)
	request := httptest.NewRequest(http.MethodGet, "/accoun??/{id}", nil)
	recorder := httptest.NewRecorder()
	uid := uuid.New()

	account := &types.Account{
		ID:               uid,
		FirstName:        "Pasha",
		LastName:         "volkov",
		CardNumber:       "444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
		Balance:          0,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}

	mockStorage.EXPECT().GetAccountByID(request.Context(), uid).Return(account, nil).AnyTimes()

	err = server.getAccountByID(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_UpdateAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", db, mockStorage)
	reqUp := &types.RequestUpdate{
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       "444444444444444",
		CardExpiryMonth:  "",
		CardExpiryYear:   "",
		CardSecurityCode: "",
	}
	err = utils.ValidateUpdateRequest(reqUp)
	buffer, err := utils.AnyToBytesBuffer(reqUp)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)
	request := httptest.NewRequest(http.MethodPut, "/account/{id}", buffer)
	recorder := httptest.NewRecorder()
	uid := uuid.New()

	account := &types.Account{
		ID:               uid,
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       "444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
		Balance:          0,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}

	mockStorage.EXPECT().UpdateAccount(request.Context(), reqUp, uid).Return(account, nil).AnyTimes()
	require.Equal(t, reqUp.FirstName, account.FirstName)
	require.Equal(t, reqUp.LastName, account.LastName)
	require.Equal(t, reqUp.CardNumber, account.CardNumber)
	
	err = server.updateAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_DeleteAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", db, mockStorage)

	request := httptest.NewRequest(http.MethodDelete, "/account/{id}", nil)
	recorder := httptest.NewRecorder()

	uid := uuid.New()

	mockStorage.EXPECT().DeleteAccount(request.Context(), uid).Return(nil).AnyTimes()

	err = server.deleteAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_DepositAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", db, mockStorage)
	reqDep := &types.RequestDeposit{
		CardNumber: "4444444444424323",
		Balance:    44,
	}
	err = utils.ValidateDepositRequest(reqDep)
	require.NoError(t, err)
	buffer, err := utils.AnyToBytesBuffer(reqDep)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)
	request := httptest.NewRequest(http.MethodPost, "/account/deposit", buffer)
	recorder := httptest.NewRecorder()
	uid := uuid.New()
	account := &types.Account{
		ID:               uid,
		FirstName:        "Pavel",
		LastName:         "volkov",
		CardNumber:       "4444444444424323",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
		Balance:          0,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}
	account2 := &types.Account{
		ID:               uid,
		FirstName:        "Pavel",
		LastName:         "volkov",
		CardNumber:       "4444444444424323",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
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

func Test_GetStatemetn(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)

	server := NewJSONApiServer("", db, mockStorage)
	request := httptest.NewRequest(http.MethodGet, "/accoun??/statement/{id}", nil)
	recorder := httptest.NewRecorder()
	uid := uuid.New()

	account := &types.Account{
		ID:               uid,
		FirstName:        "Pasha",
		LastName:         "volkov",
		CardNumber:       "444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
		Balance:          0,
		BlockedMoney:     0,
		Statement:        make([]string, 1),
		CreatedAt:        time.Now(),
	}

	mockStorage.EXPECT().GetAccountStatement(request.Context(), uid).Return(account.Statement, nil).AnyTimes()

	err = server.getAccountByID(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}