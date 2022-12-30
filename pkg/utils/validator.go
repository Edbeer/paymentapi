package utils

import (
	"errors"

	"github.com/Edbeer/paymentapi/models"
)

func ValidateCreateRequest(req *models.RequestCreate) error {
	if len(req.CardNumber) != 16 || len(req.CardExpiryMonth) != 2 || len(req.CardExpiryYear) != 2 || len(req.CardSecurityCode) != 3 {
		return errors.New("invalid parameters")
	}
	return nil
}

func ValidatePaymentRequest(req *models.PaymentRequest) error {
	if len(req.CardNumber) != 16 || len(req.CardExpiryMonth) != 2 || len(req.CardExpiryYear) != 2 || len(req.CardSecurityCode) != 3 {
		return errors.New("invalid parameters")
	}
	return nil
}

func ValidateUpdateRequest(req *models.RequestUpdate) error {
	if len(req.FirstName) >= 30 || len(req.LastName) >= 30 || len(req.CardNumber) != 16 {
		return errors.New("invalid parameters")
	}
	return nil
}

func ValidateDepositRequest(req *models.RequestDeposit) error {
	if len(req.CardNumber) != 16 {
		return errors.New("invalid parameters")
	}
	return nil
}

