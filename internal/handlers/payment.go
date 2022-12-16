package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Edbeer/paymentapi/internal/models"
	"github.com/google/uuid"
)

// Created payment: Acceptance of payment
func (s *JSONApiServer) createdPayment(w http.ResponseWriter, r *http.Request) error {
	// read body request
	reqPay := &models.PaymentRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPay); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account
	id, err := getMerchantID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	merchantAccount, err := s.storage.GetAccountByID(r.Context(), id)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// personal account
	personalAccountId := reqPay.AccountId
	personalAccount, err := s.storage.GetAccountByID(r.Context(), personalAccountId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// check payment request
	if reqPay.CardNumber != personalAccount.CardNumber || 
	reqPay.CardExpiryMonth != personalAccount.CardExpiryMonth ||
	reqPay.CardExpiryYear != personalAccount.CardExpiryYear ||
	reqPay.CardSecurityCode	!= personalAccount.CardSecurityCode {
		payment := models.CreateAuthPayment(reqPay, personalAccount, merchantAccount, "wrong payment request")
		savedPayment, err := s.storage.SavePayment(r.Context(), payment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
		merchantAccount, err = s.storage.UpdateStatement(r.Context(), id, savedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusOK, models.PaymentResponse{
			ID:     payment.ID,
			Status: payment.Status,
		})
	}
	// consume user balance
	// balance < req amount
	if personalAccount.Balance < reqPay.Amount {
		payment := models.CreateAuthPayment(reqPay, personalAccount, merchantAccount, "Insufficient funds")
		savedPayment, err := s.storage.SavePayment(r.Context(), payment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
		merchantAccount, err = s.storage.UpdateStatement(r.Context(), id, payment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusBadGateway, models.PaymentResponse{
			ID:     payment.ID,
			Status: payment.Status,
		})
	}
	// balance > req amount
	// personal acc new balance
	personalAccount.Balance = personalAccount.Balance - reqPay.Amount
	personalAccount.BlockedMoney = personalAccount.BlockedMoney + reqPay.Amount
	personalAccount, err = s.storage.SaveBalance(r.Context(), personalAccount, personalAccount.Balance, personalAccount.BlockedMoney)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account new balance
	merchantAccount.BlockedMoney = merchantAccount.BlockedMoney + reqPay.Amount
	merchantAccount, err = s.storage.SaveBalance(r.Context(), merchantAccount, merchantAccount.Balance, merchantAccount.BlockedMoney)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// create new payment
	payment := models.CreateAuthPayment(reqPay, personalAccount, merchantAccount, "Approved")
	savedPayment, err := s.storage.SavePayment(r.Context(), payment)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account append statement
	merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
	merchantAccount, err = s.storage.UpdateStatement(r.Context(), id, savedPayment.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// personal account append statement
	personalAccount.Statement = append(personalAccount.Statement, savedPayment.ID.String())
	personalAccount, err = s.storage.UpdateStatement(r.Context(), personalAccountId, savedPayment.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	return WriteJSON(w, http.StatusOK, models.PaymentResponse{
		ID:     payment.ID,
		Status: payment.Status,
	})
}

// Capture payment: Successful payment
func (s *JSONApiServer) capturePayment(w http.ResponseWriter, r *http.Request) error {
	reqPaid := &models.PaidRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPaid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	paymentId, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// get merchant
	merchantId, err := getMerchantID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	merchant, err := s.storage.GetAccountByID(r.Context(), merchantId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// check previous payment
	reqPaid.Operation = "Capture"
	reqPaid.PaymentId = paymentId
	referncedPayment, err := s.storage.GetPaymentByID(r.Context(), paymentId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	if referncedPayment.Operation == "Authorization" && referncedPayment.Status  == "Approved" {
		// Invalid amount
		if referncedPayment.Amount < reqPaid.Amount {
			completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Invalid amount")
			invalidPayment, err := s.storage.SavePayment(r.Context(), completedPayment)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
			merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, invalidPayment.ID)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			return WriteJSON(w, http.StatusOK, models.PaymentResponse{
				ID:     invalidPayment.ID,
				Status: invalidPayment.Status,
			})
		}
		// Successful payment
		referncedPayment.Amount = referncedPayment.Amount - reqPaid.Amount
		referncedPayment, err = s.storage.SavePayment(r.Context(), referncedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Successful payment")
		completedPayment, err = s.storage.SavePayment(r.Context(), completedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new personal account balance and append new statement
		personalAccount, err := s.storage.GetAccountByCard(r.Context(), referncedPayment.CardNumber)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.BlockedMoney = personalAccount.BlockedMoney - reqPaid.Amount
		personalAccount, err = s.storage.SaveBalance(r.Context(), personalAccount, 0, personalAccount.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Statement = append(personalAccount.Statement, completedPayment.ID.String())
		personalAccount, err = s.storage.UpdateStatement(r.Context(), personalAccount.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new merchant balance and append new statement
		merchant.Balance = merchant.Balance + reqPaid.Amount
		merchant.BlockedMoney = merchant.BlockedMoney - reqPaid.Amount
		merchant, err = s.storage.SaveBalance(r.Context(), merchant, merchant.Balance, merchant.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusOK, models.PaymentResponse{
			ID:     completedPayment.ID,
			Status: completedPayment.Status,
		})
	}
	return WriteJSON(w, http.StatusOK, models.PaymentResponse{
		ID:     reqPaid.PaymentId,
		Status: "Invalid transaction",
	})
}

// Refund: Refunded payment, if there is a refund
func (s *JSONApiServer) refundPayment(w http.ResponseWriter, r *http.Request) error {
	// payment id
	paymentId, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	reqPaid := &models.PaidRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPaid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// get merchant
	merchantId, err := getMerchantID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	merchant, err := s.storage.GetAccountByID(r.Context(), merchantId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// check referenced payment
	reqPaid.Operation = "Refund"
	reqPaid.PaymentId = paymentId
	referncedPayment, err := s.storage.GetPaymentByID(r.Context(), paymentId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	if referncedPayment.Operation == "Capture" && referncedPayment.Status == "Successful payment" {
		// Invalid amount
		if referncedPayment.Amount < reqPaid.Amount {
			completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Invalid amount")
			invalidPayment, err := s.storage.SavePayment(r.Context(), completedPayment)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
			merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, invalidPayment.ID)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			return WriteJSON(w, http.StatusOK, models.PaymentResponse{
				ID:     invalidPayment.ID,
				Status: invalidPayment.Status,
			})
		}
		// Successful refund
		referncedPayment.Amount = referncedPayment.Amount - reqPaid.Amount
		referncedPayment, err = s.storage.SavePayment(r.Context(), referncedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Successful refund")
		completedPayment, err = s.storage.SavePayment(r.Context(), completedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new personal account balance and append new statement
		personalAccount, err := s.storage.GetAccountByCard(r.Context(), referncedPayment.CardNumber)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Balance = personalAccount.Balance + reqPaid.Amount
		personalAccount, err = s.storage.SaveBalance(r.Context(), personalAccount, personalAccount.Balance, 0)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Statement = append(personalAccount.Statement, completedPayment.ID.String())
		personalAccount, err = s.storage.UpdateStatement(r.Context(), personalAccount.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new merchant balance and append new statement
		merchant.Balance = merchant.Balance - reqPaid.Amount
		merchant, err = s.storage.SaveBalance(r.Context(), merchant, merchant.Balance, 0)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusOK, models.PaymentResponse{
			ID:     completedPayment.ID,
			Status: completedPayment.Status,
		})
	}
	return WriteJSON(w, http.StatusOK, models.PaymentResponse{
		ID:     reqPaid.PaymentId,
		Status: "Invalid transaction",
	})
}

// cancel payment: cancel authorization payment
func (s *JSONApiServer) cancelPayment(w http.ResponseWriter, r *http.Request) error {
	// payment id
	paymentId, err := GetUUID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	reqPaid := &models.PaidRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPaid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// get merchant
	merchantId, err := getMerchantID(r)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	merchant, err := s.storage.GetAccountByID(r.Context(), merchantId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// check referenced payment
	reqPaid.Operation = "Cancel"
	reqPaid.PaymentId = paymentId
	referncedPayment, err := s.storage.GetPaymentByID(r.Context(), paymentId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	if referncedPayment.Operation == "Authorization" && referncedPayment.Status == "Approved" {
		// Invalid amount
		if referncedPayment.Amount < reqPaid.Amount {
			completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Invalid amount")
			invalidPayment, err := s.storage.SavePayment(r.Context(), completedPayment)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
			merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, invalidPayment.ID)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			return WriteJSON(w, http.StatusOK, models.PaymentResponse{
				ID:     invalidPayment.ID,
				Status: invalidPayment.Status,
			})
		}
		// Successful refund
		referncedPayment.Amount = referncedPayment.Amount - reqPaid.Amount
		referncedPayment, err = s.storage.SavePayment(r.Context(), referncedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Successful cancel")
		completedPayment, err = s.storage.SavePayment(r.Context(), completedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new personal account balance and append new statement
		personalAccount, err := s.storage.GetAccountByCard(r.Context(), referncedPayment.CardNumber)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Balance = personalAccount.Balance + reqPaid.Amount
		personalAccount.BlockedMoney = personalAccount.BlockedMoney - reqPaid.Amount
		personalAccount, err = s.storage.SaveBalance(r.Context(), personalAccount, personalAccount.Balance, personalAccount.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Statement = append(personalAccount.Statement, completedPayment.ID.String())
		personalAccount, err = s.storage.UpdateStatement(r.Context(), personalAccount.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new merchant balance and append new statement
		merchant.BlockedMoney = merchant.BlockedMoney - reqPaid.Amount
		merchant, err = s.storage.SaveBalance(r.Context(), merchant, 0, merchant.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		merchant, err = s.storage.UpdateStatement(r.Context(), merchant.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		return WriteJSON(w, http.StatusOK, models.PaymentResponse{
			ID:     completedPayment.ID,
			Status: completedPayment.Status,
		})
	}
	return WriteJSON(w, http.StatusOK, models.PaymentResponse{
		ID:     reqPaid.PaymentId,
		Status: "Invalid transaction",
	})
}

// get merchant id
func getMerchantID(r *http.Request) (uuid.UUID, error) {
	id := r.Header.Get("From")
	merchantID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}
	return merchantID, nil
}