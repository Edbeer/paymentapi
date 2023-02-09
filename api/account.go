package api

import (
	"encoding/json"
	"net/http"

	"github.com/Edbeer/paymentapi/pkg/utils"
	"github.com/Edbeer/paymentapi/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// createAccount godoc
// @Summary Create new account
// @Description register new account, returns account
// @Tags Account
// @Accept json
// @Produce json
// @Param input body types.RequestCreate true "create account info"
// @Success 200 {object} types.Account
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account [post]
func (s *JSONApiServer) createAccount(w http.ResponseWriter, r *http.Request) error {
	req := &types.RequestCreate{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	defer r.Body.Close()
	// validate request
	if err := utils.ValidateCreateRequest(req); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	account, err := s.storage.CreateAccount(r.Context(), req)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// jwt-token
	tokenString, err := utils.CreateJWT(account)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	w.Header().Add("x-jwt-token", tokenString)
	// refreshToken
	refreshToken, err := s.redisStorage.CreateSession(r.Context(), &types.Session{
		UserID: account.ID,
	}, 86400)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "error"})
	}
	// cookie
	cookie := &http.Cookie{
		Name:       "refresh-token",
		Value:      refreshToken,
		Path:       "/",
		RawExpires: "",
		MaxAge:     86400,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   0,
	}
	http.SetCookie(w, cookie)

	return WriteJSON(w, http.StatusOK, account)
}

// getAccount godoc
// @Summary Get all accounts
// @Description get all accounts, returns accounts
// @Tags Account
// @Produce json
// @Success 200 {object} []types.Account
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account [get]
func (s *JSONApiServer) getAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.storage.GetAccount(r.Context())
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

// getAccountByID godoc
// @Summary Get account by id
// @Description get account by id, returns account
// @Tags Account
// @Produce json
// @Param id path string true "get account by id info"
// @Success 200 {object} types.Account
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/{id} [get]
func (s *JSONApiServer) getAccountByID(w http.ResponseWriter, r *http.Request) error {
	uuid, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	account, err := s.storage.GetAccountByID(r.Context(), uuid)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, account)
}

// updateAccount godoc
// @Summary Update account
// @Description update account, returns updated account
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "update account info"
// @Param input body types.RequestUpdate true "update account info"
// @Success 200 {object} types.Account
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/{id} [put]
func (s *JSONApiServer) updateAccount(w http.ResponseWriter, r *http.Request) error {
	uuid, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	reqUpd := &types.RequestUpdate{}
	if err := json.NewDecoder(r.Body).Decode(reqUpd); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	defer r.Body.Close()
	// validate request
	if err := utils.ValidateUpdateRequest(reqUpd); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	account, err := s.storage.UpdateAccount(r.Context(), reqUpd, uuid)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, account)
}

// deleteAccount godoc
// @Summary Delete account
// @Description delete account, returns status
// @Tags Account
// @Produce json
// @Param id path string true "delete account info"
// @Success 200 {integer} 200
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/{id} [delete]
func (s *JSONApiServer) deleteAccount(w http.ResponseWriter, r *http.Request) error {
	uuid, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	if err := s.storage.DeleteAccount(r.Context(), uuid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, "account was deleted")
}

// depositAccount godoc
// @Summary Deposit money
// @Description deposit money to account, returns account
// @Tags Account
// @Accept json
// @Produce json
// @Param input body types.RequestDeposit true "deposit account info"
// @Success 200 {object} types.Account
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/deposit [post]
func (s *JSONApiServer) depositAccount(w http.ResponseWriter, r *http.Request) error {
	reqDep := &types.RequestDeposit{}
	if err := json.NewDecoder(r.Body).Decode(reqDep); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "account doesn't exist")
	}
	// validate request
	if err := utils.ValidateDepositRequest(reqDep); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	acc, err := s.storage.GetAccountByCard(r.Context(), reqDep.CardNumber)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "account doesn't exist")
	}
	acc.Balance = acc.Balance + reqDep.Balance
	reqDep.Balance = acc.Balance
	updatedAccount, err := s.storage.DepositAccount(r.Context(), reqDep)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "account doesn't exist")
	}

	return WriteJSON(w, http.StatusOK, updatedAccount)
}

// getStatement godoc
// @Summary Get account statement
// @Description get account statement, returns statement
// @Tags Account
// @Produce json
// @Param id path string true "get statement info"
// @Success 200 {object} types.Account.Statement
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/statement/{id} [get]
func (s *JSONApiServer) getStatement(w http.ResponseWriter, r *http.Request) error {
	uuid, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	statement, err := s.storage.GetAccountStatement(r.Context(), uuid)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, statement)
}

// signIn godoc
// @Summary Login
// @Description log in to your account, returns account
// @Tags Account
// @Accept json
// @Produce json
// @Param input body types.LoginRequest true "login account info"
// @Success 200 {object} types.Account
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/sign-in [post]
func (s *JSONApiServer) signIn(w http.ResponseWriter, r *http.Request) error {
	req := &types.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	account, err := s.storage.GetAccountByID(r.Context(), req.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	// jwt-token
	tokenString, err := utils.CreateJWT(account)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	w.Header().Add("x-jwt-token", tokenString)

	// refreshToken
	refreshToken, err := s.redisStorage.CreateSession(r.Context(), &types.Session{
		UserID: account.ID,
	}, 86400)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	// cookie
	cookie := &http.Cookie{
		Name:       "refresh-token",
		Value:      refreshToken,
		Path:       "/",
		RawExpires: "",
		MaxAge:     86400,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   0,
	}
	http.SetCookie(w, cookie)

	return WriteJSON(w, http.StatusOK, account)
}

// signOut godoc
// @Summary Logout
// @Description log out of your account, returns status
// @Tags Account
// @Produce json
// @Success 200 {integer} 200
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/sign-out [post]
func (s *JSONApiServer) signOut(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("refresh-token")
	if err != nil {
		if err == http.ErrNoCookie {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "Cookie doesn't exist"})
		}
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	if err := s.redisStorage.DeleteSession(r.Context(), cookie.Value); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, "LogOut")
}

// refreshTokens godoc
// @Summary Refresh tokens
// @Description refresh access and refresh tokens, returns tokens
// @Tags Account
// @Accept json
// @Produce json
// @Param input body types.RefreshRequest true "refresh tokens account info"
// @Success 200 {object} types.RefreshResponse
// @Failure 400  {object}  api.ApiError
// @Failure 404  {object}  api.ApiError
// @Failure 500  {object}  api.ApiError
// @Router /account/refresh [post]
func (s *JSONApiServer) refreshTokens(w http.ResponseWriter, r *http.Request) error {
	req := &types.RefreshRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	uid, err := s.redisStorage.GetUserID(r.Context(), req.RefreshToken)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	account, err := s.storage.GetAccountByID(r.Context(), uid)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	// jwt-token
	tokenString, err := utils.CreateJWT(account)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	w.Header().Add("x-jwt-token", tokenString)

	// refreshToken
	refreshToken, err := s.redisStorage.CreateSession(r.Context(), &types.Session{
		UserID: account.ID,
	}, 86400)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	// cookie
	cookie := &http.Cookie{
		Name:       "refresh-token",
		Value:      refreshToken,
		Path:       "/",
		RawExpires: "",
		MaxAge:     86400,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   0,
	}
	http.SetCookie(w, cookie)

	return WriteJSON(w, http.StatusOK, &types.RefreshResponse{
		RefreshToken: refreshToken,
		AccessToken:  tokenString,
	})
}

// Get id from url
func GetUUID(r *http.Request) (uuid.UUID, error) {
	id := mux.Vars(r)["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}
