package models

import (
	"time"

	"github.com/google/uuid"
)

// Account
type Account struct {
	ID               uuid.UUID `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	CardNumber       int64     `json:"card_number"`
	CardExpiryMonth  uint16    `json:"card_expiry_month"`
	CardExpiryYear   uint16    `json:"card_expiry_year"`
	CardSecurityCode uint32    `json:"card_security_code"`
	Balance          uint64    `json:"balance"`
	BlockedMoney     uint64    `json:"blocked_money"`
	Statement        []string  `json:"statement"`
	CreatedAt        time.Time `json:"created_at"`
}

func NewAccount(req *RequestCreate) *Account {
	return &Account{
		ID:               uuid.New(),
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		CardNumber:       req.CardNumber,
		CardExpiryMonth:  req.CardExpiryMonth,
		CardExpiryYear:   req.CardExpiryYear,
		CardSecurityCode: req.CardSecurityCode,
		Balance:          0,
		BlockedMoney:     0,
		Statement:        []string{},
		CreatedAt:        time.Now(),
	}
}

type RequestDeposit struct {
	CardNumber int64       `json:"card_number"`
	Balance    uint64    `json:"balance"`
}

// Request for update account
type RequestUpdate struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	CardNumber int    `json:"card_number"`
}

// Request for create account
type RequestCreate struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	CardNumber       int64  `json:"card_number"`
	CardExpiryMonth  uint16 `json:"card_expiry_month"`
	CardExpiryYear   uint16 `json:"card_expiry_year"`
	CardSecurityCode uint32 `json:"card_security_code"`
}
