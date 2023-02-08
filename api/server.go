//go:generate mockgen -source server.go -destination mock/storage_mock.go -package mock
package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Edbeer/paymentapi/config"
	"github.com/Edbeer/paymentapi/types"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Postgres storage interface
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

// Redis storage interface
type RedisStorage interface {
	CreateSession(ctx context.Context, session *types.Session, expire int) (string, error)
	GetUserID(ctx context.Context, refreshToken string) (uuid.UUID, error)
	DeleteSession(ctx context.Context, refreshToken string) error
}

// Server
type JSONApiServer struct {
	config       *config.Config
	storage      Storage
	redisStorage RedisStorage
	Server       *http.Server
	db           *sql.DB
	redis        *redis.Client
}

// Constructor
func NewJSONApiServer(config *config.Config, db *sql.DB, redis *redis.Client, storage Storage, redisStorage RedisStorage) *JSONApiServer {
	return &JSONApiServer{
		db:           db,
		redis:        redis,
		storage:      storage,
		redisStorage: redisStorage,
		Server: &http.Server{
			Addr:         config.Server.Port,
			ReadTimeout:  time.Duration(config.Server.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(config.Server.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(config.Server.IdleTimeout) * time.Second,
		},
	}
}

func (s *JSONApiServer) Run() {
	router := mux.NewRouter()
	// POST
	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/account", HTTPHandler(s.createAccount))
	postRouter.HandleFunc("/account/sign-in", HTTPHandler(s.signIn))
	postRouter.HandleFunc("/account/sign-out", HTTPHandler(s.signOut))
	postRouter.HandleFunc("/account/deposit", HTTPHandler(s.depositAccount))
	postRouter.HandleFunc("/account/refresh", HTTPHandler(s.refreshTokens))
	// payment
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
