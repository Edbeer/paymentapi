package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type JSONApiServer struct {
	listenAddr string
}

func NewJSONApiServer(listenAddr string) *JSONApiServer {
	return &JSONApiServer{
		listenAddr: listenAddr,
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
	// UPDATE
	putRouter := router.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/account", HTTPHanddler(s.updateAccount))
	// DELETE
	deleteRouter := router.Methods(http.MethodDelete).Subrouter()
	deleteRouter.HandleFunc("/account", HTTPHanddler(s.deleteAccount))

	http.ListenAndServe(s.listenAddr, router)
}

func (s *JSONApiServer) createAccount(w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("Hello"))
	return nil
}

func (s *JSONApiServer) getAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *JSONApiServer) updateAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *JSONApiServer) deleteAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
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