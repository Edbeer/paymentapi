package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Edbeer/paymentapi/pkg/utils"
	"github.com/Edbeer/paymentapi/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

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
	fmt.Println(tokenString)
	return WriteJSON(w, http.StatusOK, account)
}

// get all accounts
func (s *JSONApiServer) getAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.storage.GetAccount(r.Context())
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

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
