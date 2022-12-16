package models

import (
	"time"

	"github.com/google/uuid"
)

// TODO FIX order_id
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

// creating a payment
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

// creating a complete payment
func CreateCompletePayment(paidPayment *PaidRequest, referncedPayment *Payment, status string) *Payment {
	return &Payment{
		ID:              uuid.New(),
		BusinessId:      referncedPayment.BusinessId,
		OrderId:         paidPayment.OrderId,
		Operation:       paidPayment.Operation,
		Amount:          paidPayment.Amount,
		Status:          status,
		Currency:        "RUB",
		CardNumber:      referncedPayment.CardNumber,
		CardExpiryMonth: referncedPayment.CardExpiryMonth,
		CardExpiryYear:  referncedPayment.CardExpiryYear,
		CreatedAt:       time.Now(),
	}
}

type PaidRequest struct {
	OrderId   string    `json:"order_id"`
	PaymentId uuid.UUID `json:"payment_id"`
	Operation string    `json:"operation"`
	Amount    uint64    `json:"amount"`
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