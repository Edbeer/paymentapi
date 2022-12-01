package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type JSONApiServer struct {
	storage Storage
	server  *http.Server
}

func NewJSONApiServer(listenAddr string, storage Storage) *JSONApiServer {
	return &JSONApiServer{
		storage: storage,
		server: &http.Server{
			Addr:         listenAddr,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (s *JSONApiServer) Run() {
	router := mux.NewRouter()
	// POST
	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/account", HTTPHandler(s.createAccount))
	postRouter.HandleFunc("/account/deposit", HTTPHandler(s.depositAccount))
	postRouter.HandleFunc("/payment/auth", HTTPHandler(s.createdPayment))
	postRouter.HandleFunc("/payment/capture/{id}", HTTPHandler(s.capturePayment))
	// GET
	getRouter := router.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/account", HTTPHandler(s.getAccount))
	getRouter.HandleFunc("/account/{id}", AuthJWT(HTTPHandler(s.getAccountByID)))
	// UPDATE
	putRouter := router.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/account/{id}", AuthJWT(HTTPHandler(s.updateAccount)))
	// DELETE
	deleteRouter := router.Methods(http.MethodDelete).Subrouter()
	deleteRouter.HandleFunc("/account/{id}", AuthJWT(HTTPHandler(s.deleteAccount)))
	s.server.Handler = router
	s.server.ListenAndServe()
}

func (s *JSONApiServer) createAccount(w http.ResponseWriter, r *http.Request) error {
	req := &RequestCreate{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	defer r.Body.Close()
	reqAcc := NewAccount(req.FirstName, req.LastName)
	account, err := s.storage.CreateAccount(r.Context(), reqAcc)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	tokenString, err := CreateJWT(account)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	w.Header().Add("x-jwt-token", tokenString)
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
	reqUpd := &Account{}
	if err := json.NewDecoder(r.Body).Decode(reqUpd); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	defer r.Body.Close()
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
	reqDep := &RequestDeposit{}
	if err := json.NewDecoder(r.Body).Decode(reqDep); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "account doesn't exist")
	}

	acc, err := s.storage.GetAccountByID(r.Context(), reqDep.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "ccount doesn't exist")
	}
	acc.Balance = acc.Balance + reqDep.Balance
	reqDep.Balance = acc.Balance
	updatedAccount, err := s.storage.DepositAccount(r.Context(), reqDep)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "account doesn't exist")
	}

	return WriteJSON(w, http.StatusOK, updatedAccount)
}

// Created payment: Acceptance of payment
func (s *JSONApiServer) createdPayment(w http.ResponseWriter, r *http.Request) error {
	// read body request
	reqPay := &PaymentRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPay); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account
	id, err := merchantID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	merchantAccount, err := s.storage.GetAccountByID(r.Context(), id)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// personal account
	personalAccountId := reqPay.AccountId
	personalAccount, err := s.storage.GetAccountByID(r.Context(), personalAccountId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// check payment request
	if reqPay.CardNumber != personalAccount.CardNumber || 
	reqPay.CardExpiryMonth != personalAccount.CardExpiryMonth ||
	reqPay.CardExpiryYear != personalAccount.CardExpiryYear ||
	reqPay.CardSecurityCode	!= personalAccount.CardSecurityCode {
		payment := CreateAuthPayment(reqPay, personalAccount, merchantAccount, "wrong payment request")
		savedPayment, err := s.storage.SavePayment(r.Context(), payment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
		merchantAccount, err = s.storage.UpdateStatement(r.Context(), id, savedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusOK, PaymentResponse{
			ID:     payment.ID,
			Status: payment.Status,
		})
	}
	// consume user balance
	// balance < req amount
	if personalAccount.Balance < reqPay.Amount {
		payment := CreateAuthPayment(reqPay, personalAccount, merchantAccount, "Insufficient funds")
		savedPayment, err := s.storage.SavePayment(r.Context(), payment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
		merchantAccount, err = s.storage.UpdateStatement(r.Context(), id, payment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusBadGateway, PaymentResponse{
			ID:     payment.ID,
			Status: payment.Status,
		})
	}
	// balance > req amount
	// personal acc new balance
	personalAccount.Balance = personalAccount.Balance - reqPay.Amount
	personalAccount.BlockedMoney = personalAccount.BlockedMoney + reqPay.Amount
	personalAccount, err = s.storage.SaveBalance(r.Context(), personalAccount, personalAccount.Balance, personalAccount.BlockedMoney)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account new balance
	merchantAccount.BlockedMoney = merchantAccount.BlockedMoney + reqPay.Amount
	merchantAccount, err = s.storage.SaveBalance(r.Context(), merchantAccount, merchantAccount.Balance, merchantAccount.BlockedMoney)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// create new payment
	payment := CreateAuthPayment(reqPay, personalAccount, merchantAccount, "Approved")
	savedPayment, err := s.storage.SavePayment(r.Context(), payment)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account append statement
	merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
	merchantAccount, err = s.storage.UpdateStatement(r.Context(), id, savedPayment.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// personal account append statement
	personalAccount.Statement = append(personalAccount.Statement, savedPayment.ID.String())
	personalAccount, err = s.storage.UpdateStatement(r.Context(), personalAccountId, savedPayment.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, PaymentResponse{
		ID:     payment.ID,
		Status: payment.Status,
	})
}

// Capture payment: Successful payment
func (s *JSONApiServer) capturePayment(w http.ResponseWriter, r *http.Request) error {
	reqPaid := &PaidRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPaid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	paymentId, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// get merchant
	merchantId, err := merchantID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	merchant, err := s.storage.GetAccountByID(r.Context(), merchantId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// check previous payment
	reqPaid.Operation = "Capture"
	reqPaid.PaymentId = paymentId
	referncedPayment, err := s.storage.GetPaymentByID(r.Context(), paymentId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	if referncedPayment.Operation == "Authorization" && referncedPayment.Status  == "Approved" {
		// Invalid amount
		if referncedPayment.Amount < reqPaid.Amount {
			completedPayment := CreateCompletePayment(reqPaid, referncedPayment, "Invalid amount")
			invalidPayment, err := s.storage.SavePayment(r.Context(), completedPayment)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
			merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, invalidPayment.ID)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			return WriteJSON(w, http.StatusOK, PaymentResponse{
				ID:     invalidPayment.ID,
				Status: invalidPayment.Status,
			})
		}
		// Successful payment
		referncedPayment.Amount = referncedPayment.Amount - reqPaid.Amount
		referncedPayment, err = s.storage.SavePayment(r.Context(), referncedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		completedPayment := CreateCompletePayment(reqPaid, referncedPayment, "Successful payment")
		completedPayment, err = s.storage.SavePayment(r.Context(), completedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new personal account balance and append new statement
		personalAccount, err := s.storage.GetAccountByCard(r.Context(), referncedPayment.CardNumber)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.BlockedMoney = personalAccount.BlockedMoney - reqPaid.Amount
		personalAccount, err = s.storage.SaveBalance(r.Context(), personalAccount, 0, personalAccount.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Statement = append(personalAccount.Statement, completedPayment.ID.String())
		personalAccount, err = s.storage.UpdateStatement(r.Context(), personalAccount.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new merchant balance and append new statement
		merchant.Balance = merchant.Balance + reqPaid.Amount
		merchant.BlockedMoney = merchant.BlockedMoney - reqPaid.Amount
		merchant, err = s.storage.SaveBalance(r.Context(), merchant, merchant.Balance, merchant.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusOK, PaymentResponse{
			ID:     completedPayment.ID,
			Status: completedPayment.Status,
		})
	}
	return WriteJSON(w, http.StatusOK, PaymentResponse{
		ID:     reqPaid.PaymentId,
		Status: "Invalid transaction",
	})
}

// Refunded: Refunded payment, if there is a refund
func (s *JSONApiServer) refundedPayment(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func merchantID(r *http.Request) (uuid.UUID, error) {
	id := r.Header.Get("From")
	merchantID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}
	return merchantID, nil
}

type ApiFunc func(w http.ResponseWriter, r *http.Request) error

// Wrapper for handler func
func HTTPHandler(f ApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type ApiError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// auth middleware
func AuthJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("x-jwt-token")
		token, err := ValidateJWT(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}
		uid, err := GetUUID(r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}

		if claims["id"] != uid.String() {
			WriteJSON(w, http.StatusBadRequest, "permission denied")
			return
		}

		next(w, r)
	}
}

// Create JWT
func CreateJWT(account *Account) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":        account.ID.String(),
		"card":      account.CardNumber,
		"expire_at": 15000,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Validate JWT
func ValidateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
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
