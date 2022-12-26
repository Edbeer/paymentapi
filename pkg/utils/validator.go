package utils

import (
	"errors"

	"github.com/Edbeer/paymentapi/internal/models"
)

func ValidateCreateRequest(req *models.RequestCreate) error {
	if len(req.CardNumber) != 16 || len(req.CardExpiryMonth) != 2 || len(req.CardExpiryYear) != 2 || len(req.CardSecurityCode) != 3 {
		return errors.New("invalid parameters")
	}
	return nil
}

