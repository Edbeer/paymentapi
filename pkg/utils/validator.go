package utils

import (
	"errors"

	"github.com/Edbeer/paymentapi/types"
)

func ValidateCreateRequest(req *types.RequestCreate) error {
	if len(req.CardNumber) != 16 || len(req.CardExpiryMonth) != 2 || len(req.CardExpiryYear) != 2 || len(req.CardSecurityCode) != 3 {
		return errors.New("invalid parameters")
	}
	return nil
}

func ValidatePaymentRequest(req *types.PaymentRequest) error {
	if len(req.CardNumber) != 16 || len(req.CardExpiryMonth) != 2 || len(req.CardExpiryYear) != 2 || len(req.CardSecurityCode) != 3 {
		return errors.New("invalid parameters")
	}
	return nil
}

func ValidateUpdateRequest(req *types.RequestUpdate) error {
	if len(req.FirstName) >= 30 || len(req.LastName) >= 30 || len(req.CardNumber) != 16 {
		return errors.New("invalid parameters")
	}
	return nil
}

func ValidateDepositRequest(req *types.RequestDeposit) error {
	if len(req.CardNumber) != 16 {
		return errors.New("invalid parameters")
	}
	return nil
}

