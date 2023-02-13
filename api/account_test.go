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
	"github.com/alicebob/miniredis"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func Test_CreateAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockStorage := mockstore.NewMockStorage(ctrl)
	mockRedis := mockstore.NewMockRedisStorage(ctrl)

	config := &config.Config{}
	server := NewJSONApiServer(config, db, client, mockStorage, mockRedis, nil)
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
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.createAccount")
	defer span.Finish()

	recorder := httptest.NewRecorder()
	reqAcc := types.NewAccount(req)
	mockStorage.EXPECT().CreateAccount(ctxWithTrace, gomock.Eq(req)).Return(&types.Account{
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

	sess := &types.Session{
		UserID: reqAcc.ID,
	}
	token := "refresh-token"
	mockRedis.EXPECT().CreateSession(ctxWithTrace, gomock.Eq(sess), 86400).Return(token, nil)

	err = server.createAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
	require.NotNil(t, reqAcc.ID)
}

func Test_SignIn(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockStorage := mockstore.NewMockStorage(ctrl)
	mockRedis := mockstore.NewMockRedisStorage(ctrl)

	config := &config.Config{}
	server := NewJSONApiServer(config, db, client, mockStorage, mockRedis, nil)

	req := &types.LoginRequest{
		ID: uuid.New(),
	}

	buffer, err := utils.AnyToBytesBuffer(req)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/account/sign-in", buffer)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.signIn")
	defer span.Finish()
	recorder := httptest.NewRecorder()

	account := &types.Account{
		ID:               req.ID,
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       "4444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
		Balance:          0,
		BlockedMoney:     0,
		Statement:        []string{},
		CreatedAt:        time.Now(),
	}

	mockStorage.EXPECT().GetAccountByID(ctxWithTrace, req.ID).Return(account, nil).AnyTimes()
	sess := &types.Session{
		UserID: req.ID,
	}
	token := "refresh-token"
	mockRedis.EXPECT().CreateSession(ctxWithTrace, gomock.Eq(sess), 86400).Return(token, nil)

	err = server.signIn(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
	require.NotNil(t, req.ID)
}

func Test_SignOut(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockStorage := mockstore.NewMockStorage(ctrl)
	mockRedis := mockstore.NewMockRedisStorage(ctrl)

	config := &config.Config{}
	server := NewJSONApiServer(config, db, client, mockStorage, mockRedis, nil)

	request := httptest.NewRequest(http.MethodPost, "/account/sign-out", nil)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.signOut")
	defer span.Finish()
	recorder := httptest.NewRecorder()

	token := "refresh-token"
	cookieValue := "cookieValue"

	request.AddCookie(&http.Cookie{Name: token, Value: cookieValue})


	cookie, err := request.Cookie(token)
	require.NoError(t, err)
	require.NotNil(t, cookie)
	require.NotEqual(t, cookie.Value, "")
	require.Equal(t, cookie.Value, cookieValue)

	mockRedis.EXPECT().DeleteSession(ctxWithTrace, gomock.Eq(cookie.Value)).Return(nil)

	err = server.signOut(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_RefreshTokens(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockStorage := mockstore.NewMockStorage(ctrl)
	mockRedis := mockstore.NewMockRedisStorage(ctrl)
	config := &config.Config{}
	
	server := NewJSONApiServer(config, db, client, mockStorage, mockRedis, nil)

	req := &types.RefreshRequest{
		RefreshToken: "cookieValue",
	}

	buffer, err := utils.AnyToBytesBuffer(req)
	require.NoError(t, err)
	require.NotNil(t, buffer)
	require.Nil(t, err)

	request := httptest.NewRequest(http.MethodPost, "/account/refresh", buffer)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.refreshTokens")
	defer span.Finish()

	recorder := httptest.NewRecorder()
	uid := uuid.New()
	mockRedis.EXPECT().GetUserID(ctxWithTrace, req.RefreshToken).Return(uid, nil).AnyTimes()

	account := &types.Account{
		ID:               uid,
		FirstName:        "Pasha1",
		LastName:         "volkov1",
		CardNumber:       "4444444444444444",
		CardExpiryMonth:  "12",
		CardExpiryYear:   "24",
		CardSecurityCode: "924",
		Balance:          0,
		BlockedMoney:     0,
		Statement:        []string{},
		CreatedAt:        time.Now(),
	}

	mockStorage.EXPECT().GetAccountByID(ctxWithTrace, uid).Return(account, nil).AnyTimes()

	token := "refresh-token"
	sess := &types.Session{
		UserID: uid,
	}
	mockRedis.EXPECT().CreateSession(ctxWithTrace, gomock.Eq(sess), 86400).Return(token, nil)

	request.AddCookie(&http.Cookie{Name: token, Value: req.RefreshToken})

	cookie, err := request.Cookie(token)
	require.NoError(t, err)
	require.NotNil(t, cookie)
	require.NotEqual(t, cookie.Value, "")
	require.Equal(t, cookie.Value, req.RefreshToken)

	err = server.refreshTokens(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_GetAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)

	request := httptest.NewRequest(http.MethodGet, "/account", nil)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.getAccount")
	defer span.Finish()

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

	mockStorage.EXPECT().GetAccount(ctxWithTrace).Return(accounts, nil).AnyTimes()

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

	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)
	request := httptest.NewRequest(http.MethodGet, "/accounе/{id}", nil)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.getAccountByID")
	defer span.Finish()

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

	mockStorage.EXPECT().GetAccountByID(ctxWithTrace, uid).Return(account, nil).AnyTimes()

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

	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)
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
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.updateAccount")
	defer span.Finish()

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

	mockStorage.EXPECT().UpdateAccount(ctxWithTrace, reqUp, uid).Return(account, nil).AnyTimes()
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
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)

	request := httptest.NewRequest(http.MethodDelete, "/account/{id}", nil)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.deleteAccount")
	defer span.Finish()

	recorder := httptest.NewRecorder()

	uid := uuid.New()

	mockStorage.EXPECT().DeleteAccount(ctxWithTrace, uid).Return(nil).AnyTimes()

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
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)
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
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.depositAccount")
	defer span.Finish()

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
	mockStorage.EXPECT().GetAccountByCard(ctxWithTrace, reqDep.CardNumber).Return(account, nil).AnyTimes()
	mockStorage.EXPECT().DepositAccount(ctxWithTrace, reqDep).Return(account2, nil).AnyTimes()

	err = server.depositAccount(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}

func Test_GetStatement(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockStorage := mockstore.NewMockStorage(ctrl)
	config := &config.Config{}
	server := NewJSONApiServer(config, db, nil, mockStorage, nil, nil)
	request := httptest.NewRequest(http.MethodGet, "/accounе/statement/{id}", nil)
	span, ctxWithTrace := opentracing.StartSpanFromContext(request.Context(), "Account.getStatement")
	defer span.Finish()

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

	mockStorage.EXPECT().GetAccountStatement(ctxWithTrace, uid).Return(account.Statement, nil).AnyTimes()

	err = server.getAccountByID(recorder, request)
	require.NoError(t, err)
	require.Nil(t, err)
}