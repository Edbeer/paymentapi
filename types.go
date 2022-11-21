package main

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type RequestUpdate struct {
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	CardNumber uint64    `json:"card_number"`
}

type RequestCreate struct {
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
}

type Account struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	CardNumber uint64    `json:"card_number"`
	Balance    uint64    `json:"balance"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		ID:         uuid.New(),
		FirstName:  firstName,
		LastName:   lastName,
		CardNumber: uint64(rand.Intn(10000000)),
		Balance:    0,
		CreatedAt:  time.Now(),
	}
}
