package main

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Payment
type Payment struct {
	ID          uuid.UUID `json:"id"`
	BusinessId  uuid.UUID `json:"business_id"`
	OrderId     string    `json:"order_id"`
	Operation   string    `json:"operation"`
	Amount      uint64    `json:"amount"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Currency    string    `json:"currency"`
	CardNumber  uint64    `json:"card_number"`
	CreatedAt   time.Time `json:"creation_at"`
}

func NewPayment(paymentCreate *PaymentRequest, personalAccount *Account, businessAccount *Account) *Payment {
	return &Payment{
		ID:         uuid.New(),
		BusinessId: businessAccount.ID,
		OrderId:    paymentCreate.OrderId,
		Operation:  "",
		Amount:     paymentCreate.Amount,
		Currency:   "RUB",
		CardNumber: personalAccount.CardNumber,
		CreatedAt:  time.Now(),
	}
}

type PaymentRequest struct {
	AccountId        uuid.UUID `json:"id"`
	OrderId          string    `json:"order_id"`
	Amount           uint64    `json:"amount"`
	Currency         string    `json:"currency"`
	CardNumber       uint64    `json:"card_number"`
	CardSecurityCode int       `json:"card_security_code"`
}

type RequestDeposit struct {
	ID         uuid.UUID `json:"id"`
	CardNumber uint64    `json:"card_number"`
	Balance    uint64    `json:"balance"`
}

// Request for update account
type RequestUpdate struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	CardNumber uint64 `json:"card_number"`
}

// Request for create account
type RequestCreate struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Account
type Account struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	CardNumber uint64    `json:"card_number"`
	Balance    uint64    `json:"balance"`
	Statement  []string  `json:"statement"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		ID:         uuid.New(),
		FirstName:  firstName,
		LastName:   lastName,
		CardNumber: uint64(rand.Intn(10000000)),
		Balance:    0,
		Statement: []string{},
		CreatedAt:  time.Now(),
	}
}
