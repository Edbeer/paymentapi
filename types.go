package main

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Payment
type Payment struct {
	ID              uuid.UUID `json:"id"`
	BusinessId      uuid.UUID `json:"business_id"`
	OrderId         string    `json:"order_id"`
	Operation       string    `json:"operation"`
	Amount          uint64    `json:"amount"`
	Status          string    `json:"status"`
	Currency        string    `json:"currency"`
	CardNumber      uint64    `json:"card_number"`
	CardExpiryMonth uint64    `json:"card_expiry_month"`
	CardExpiryYear  uint64    `json:"card_expiry_year"`
	CreatedAt       time.Time `json:"creation_at"`
}

func CreateAuthPayment(paymentCreate *PaymentRequest, personalAccount *Account, merchantAccount *Account, status string) *Payment {
	return &Payment{
		ID:              uuid.New(),
		BusinessId:      merchantAccount.ID,
		OrderId:         paymentCreate.OrderId,
		Operation:       "Authorization",
		Amount:          paymentCreate.Amount,
		Status:          status,
		Currency:        "RUB",
		CardNumber:      personalAccount.CardNumber,
		CardExpiryMonth: personalAccount.CardExpiryMonth,
		CardExpiryYear:  personalAccount.CardExpiryYear,
		CreatedAt:       time.Now(),
	}
}

// TODO remove account id
type PaymentRequest struct {
	AccountId        uuid.UUID `json:"id"`
	OrderId          string    `json:"order_id"`
	Amount           uint64    `json:"amount"`
	Currency         string    `json:"currency"`
	CardNumber       uint64    `json:"card_number"`
	CardExpiryMonth  uint64    `json:"card_expiry_month"`
	CardExpiryYear   uint64    `json:"card_expiry_year"`
	CardSecurityCode uint64    `json:"card_security_code"`
}

// TODO auth code
type PaymentResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
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
	ID               uuid.UUID `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	CardNumber       uint64    `json:"card_number"`
	CardExpiryMonth  uint64    `json:"card_expiry_month"`
	CardExpiryYear   uint64    `json:"card_expiry_year"`
	CardSecurityCode uint64    `json:"card_security_code"`
	Balance          uint64    `json:"balance"`
	BlockedMoney     uint64    `json:"blocked_money"`
	Statement        []string  `json:"statement"`
	CreatedAt        time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		ID:               uuid.New(),
		FirstName:        firstName,
		LastName:         lastName,
		CardNumber:       uint64(rand.Intn(999999999-233333333) + 233333333),
		CardExpiryMonth:  uint64(rand.Intn(12-1) + 1),
		CardExpiryYear:   uint64(rand.Intn(26-2) + 2),
		CardSecurityCode: uint64(rand.Intn(999-100) + 100),
		Balance:          0,
		BlockedMoney:     0,
		Statement:        []string{},
		CreatedAt:        time.Now(),
	}
}
