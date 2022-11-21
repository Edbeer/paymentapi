package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type JSONApiServer struct {
	listenAddr string
	storage Storage
}

func NewJSONApiServer(listenAddr string, storage Storage) *JSONApiServer {
	return &JSONApiServer{
		listenAddr: listenAddr,
		storage: storage,
	}
}

func (s *JSONApiServer) Run() {
	router := mux.NewRouter()
	// POST
	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/account", HTTPHanddler(s.createAccount))
	// GET
	getRouter := router.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/account", HTTPHanddler(s.getAccount))
	getRouter.HandleFunc("/account/{id}", HTTPHanddler(s.getAccountByID))
	// UPDATE
	putRouter := router.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/account/{id}", HTTPHanddler(s.updateAccount))
	// DELETE
	deleteRouter := router.Methods(http.MethodDelete).Subrouter()
	deleteRouter.HandleFunc("/account/{id}", HTTPHanddler(s.deleteAccount))

	http.ListenAndServe(s.listenAddr, router)
}

func (s *JSONApiServer) createAccount(w http.ResponseWriter, r *http.Request) error {
	req := &RequestCreate{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	reqAcc := NewAccount(req.FirstName, req.LastName)
	account, err := s.storage.CreateAccount(reqAcc)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *JSONApiServer) getAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.storage.GetAccount()
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *JSONApiServer) getAccountByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	uuid, err := uuid.Parse(id)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	account, err := s.storage.GetAccountByID(uuid)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *JSONApiServer) updateAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	uuid, err := uuid.Parse(id)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	reqUpd := &RequestUpdate{}
	if err := json.NewDecoder(r.Body).Decode(reqUpd); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	account, err := s.storage.UpdateAccount(reqUpd.FirstName, reqUpd.LastName, reqUpd.CardNumber, uuid)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *JSONApiServer) deleteAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	uuid, err := uuid.Parse(id)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	if err := s.storage.DeleteAccount(uuid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, "account was deleted")
}

type ApiFunc func(w http.ResponseWriter, r *http.Request) error

func HTTPHanddler(f ApiFunc) http.HandlerFunc {
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