package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Edbeer/paymentapi/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Storage interface {
	CreateAccount(ctx context.Context, account *models.Account) (*models.Account, error)
	GetAccount(ctx context.Context) ([]*models.Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	GetAccountByCard(ctx context.Context, card uint64) (*models.Account, error)
	UpdateAccount(ctx context.Context, reqUp *models.Account, id uuid.UUID) (*models.Account, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	DepositAccount(ctx context.Context, reqDep *models.RequestDeposit) (*models.Account, error)
	SavePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error)
	GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error)
	SaveBalance(ctx context.Context, account *models.Account, balance, bmoney uint64) (*models.Account, error)
	UpdateStatement(ctx context.Context, id, paymentId uuid.UUID) (*models.Account, error)
}

type JSONApiServer struct {
	storage Storage
	Server  *http.Server
}

func NewJSONApiServer(listenAddr string, storage Storage) *JSONApiServer {
	return &JSONApiServer{
		storage: storage,
		Server: &http.Server{
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
	postRouter.HandleFunc("/payment/refund/{id}", HTTPHandler(s.refundPayment))
	postRouter.HandleFunc("/payment/cancel/{id}", HTTPHandler(s.cancelPayment))
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
	s.Server.Handler = router
	s.Server.ListenAndServe()
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