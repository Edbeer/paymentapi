//go:generate mockgen -source server.go -destination mock/storage_mock.go -package mock
package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Edbeer/paymentapi/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Storage interface {
	CreateAccount(ctx context.Context, reqAcc *types.RequestCreate) (*types.Account, error)
	GetAccount(ctx context.Context) ([]*types.Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*types.Account, error)
	GetAccountByCard(ctx context.Context, card string) (*types.Account, error)
	UpdateAccount(ctx context.Context, reqUp *types.RequestUpdate, id uuid.UUID) (*types.Account, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	DepositAccount(ctx context.Context, reqDep *types.RequestDeposit) (*types.Account, error)
	GetAccountStatement(ctx context.Context, id uuid.UUID) ([]string, error)
	SavePayment(ctx context.Context, tx *sql.Tx, payment *types.Payment) (*types.Payment, error)
	GetPaymentByID(ctx context.Context, id uuid.UUID) (*types.Payment, error)
	SaveBalance(ctx context.Context, tx *sql.Tx, account *types.Account, balance, bmoney uint64) (*types.Account, error)
	UpdateStatement(ctx context.Context, tx *sql.Tx, id, paymentId uuid.UUID) (*types.Account, error)
}

type JSONApiServer struct {
	storage Storage
	Server  *http.Server
	db *sql.DB
}

func NewJSONApiServer(listenAddr string, db *sql.DB, storage Storage) *JSONApiServer {
	return &JSONApiServer{
		db: db,
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
	postRouter.HandleFunc("/payment/auth", HTTPHandler(s.createPayment))
	postRouter.HandleFunc("/payment/capture/{id}", HTTPHandler(s.capturePayment))
	postRouter.HandleFunc("/payment/refund/{id}", HTTPHandler(s.refundPayment))
	postRouter.HandleFunc("/payment/cancel/{id}", HTTPHandler(s.cancelPayment))
	// GET
	getRouter := router.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/account", HTTPHandler(s.getAccount))
	getRouter.HandleFunc("/account/{id}", AuthJWT(HTTPHandler(s.getAccountByID)))
	getRouter.HandleFunc("/account/statement/{id}", AuthJWT(HTTPHandler(s.getStatement)))
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