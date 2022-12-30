package api

import (
	"encoding/json"
	"net/http"

	"github.com/Edbeer/paymentapi/models"
	"github.com/Edbeer/paymentapi/pkg/utils"
	"github.com/google/uuid"
)

// Create payment: Acceptance of payment
func (s *JSONApiServer) createPayment(w http.ResponseWriter, r *http.Request) error {
	// read body request
	reqPay := &models.PaymentRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPay); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	defer r.Body.Close()
	// validate request
	if err := utils.ValidatePaymentRequest(reqPay); err != nil {
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
		// Begin Transaction
		tx, err := s.db.BeginTx(r.Context(), nil)
		defer tx.Rollback()
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
		}
		payment := models.CreateAuthPayment(reqPay, personalAccount, merchantAccount, "wrong payment request")
		savedPayment, err := s.storage.SavePayment(r.Context(), tx, payment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
		merchantAccount, err = s.storage.UpdateStatement(r.Context(), tx, id, savedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// Commit transaction
		if err := tx.Commit(); err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
		} 
		return WriteJSON(w, http.StatusOK, models.PaymentResponse{
			ID:     payment.ID,
			Status: payment.Status,
		})
	}
	// consume user balance
	// balance < req amount
	if personalAccount.Balance < reqPay.Amount {
		// Begin Transaction
		tx, err := s.db.BeginTx(r.Context(), nil)
		defer tx.Rollback()
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
		}
		payment := models.CreateAuthPayment(reqPay, personalAccount, merchantAccount, "Insufficient funds")
		savedPayment, err := s.storage.SavePayment(r.Context(), tx, payment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
		merchantAccount, err = s.storage.UpdateStatement(r.Context(), tx, id, payment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// Commit transaction
		if err := tx.Commit(); err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
		} 
		return WriteJSON(w, http.StatusBadGateway, models.PaymentResponse{
			ID:     payment.ID,
			Status: payment.Status,
		})
	}
	// balance > req amount
	// Begin transaction
	tx, err := s.db.BeginTx(r.Context(), nil)
	defer tx.Rollback()
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
	}
	// personal acc new balance
	personalAccount.Balance = personalAccount.Balance - reqPay.Amount
	personalAccount.BlockedMoney = personalAccount.BlockedMoney + reqPay.Amount
	personalAccount, err = s.storage.SaveBalance(r.Context(), tx, personalAccount, personalAccount.Balance, personalAccount.BlockedMoney)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account new balance
	merchantAccount.BlockedMoney = merchantAccount.BlockedMoney + reqPay.Amount
	merchantAccount, err = s.storage.SaveBalance(r.Context(), tx, merchantAccount, merchantAccount.Balance, merchantAccount.BlockedMoney)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// create new payment
	payment := models.CreateAuthPayment(reqPay, personalAccount, merchantAccount, "Approved")
	savedPayment, err := s.storage.SavePayment(r.Context(), tx, payment)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// merchant account append statement
	merchantAccount.Statement = append(merchantAccount.Statement, savedPayment.ID.String())
	merchantAccount, err = s.storage.UpdateStatement(r.Context(), tx, id, savedPayment.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// personal account append statement
	personalAccount.Statement = append(personalAccount.Statement, savedPayment.ID.String())
	personalAccount, err = s.storage.UpdateStatement(r.Context(), tx, personalAccountId, savedPayment.ID)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
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
	defer r.Body.Close()
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
			// Begin transaction
			tx, err := s.db.BeginTx(r.Context(), nil)
			defer tx.Rollback()
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
			}
			completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Invalid amount")
			invalidPayment, err := s.storage.SavePayment(r.Context(), tx, completedPayment)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
			merchant, err = s.storage.UpdateStatement(r.Context(), tx, merchant.ID, invalidPayment.ID)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			// Commit transaction
			if err := tx.Commit(); err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
			} 
			return WriteJSON(w, http.StatusOK, models.PaymentResponse{
				ID:     invalidPayment.ID,
				Status: invalidPayment.Status,
			})
		}
		// Successful payment
		// Begin Transaction
		tx, err := s.db.BeginTx(r.Context(), nil)
		defer tx.Rollback()
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
		}
		referncedPayment.Amount = referncedPayment.Amount - reqPaid.Amount
		referncedPayment, err = s.storage.SavePayment(r.Context(), tx, referncedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Successful payment")
		completedPayment, err = s.storage.SavePayment(r.Context(), tx, completedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new personal account balance and append new statement
		personalAccount, err := s.storage.GetAccountByCard(r.Context(), referncedPayment.CardNumber)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.BlockedMoney = personalAccount.BlockedMoney - reqPaid.Amount
		personalAccount, err = s.storage.SaveBalance(r.Context(), tx, personalAccount, 0, personalAccount.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Statement = append(personalAccount.Statement, completedPayment.ID.String())
		personalAccount, err = s.storage.UpdateStatement(r.Context(), tx, personalAccount.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new merchant balance and append new statement
		merchant.Balance = merchant.Balance + reqPaid.Amount
		merchant.BlockedMoney = merchant.BlockedMoney - reqPaid.Amount
		merchant, err = s.storage.SaveBalance(r.Context(), tx, merchant, merchant.Balance, merchant.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		merchant, err = s.storage.UpdateStatement(r.Context(), tx, merchant.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// Commit transaction
		if err := tx.Commit(); err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
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
	reqPaid := &models.PaidRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPaid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	defer r.Body.Close()
	// payment id
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
			// Begin Transaction
			tx, err := s.db.BeginTx(r.Context(), nil)
			defer tx.Rollback()
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
			}
			completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Invalid amount")
			invalidPayment, err := s.storage.SavePayment(r.Context(), tx, completedPayment)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
			merchant, err = s.storage.UpdateStatement(r.Context(), tx, merchant.ID, invalidPayment.ID)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			// Commit transaction
			if err := tx.Commit(); err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
			} 
			return WriteJSON(w, http.StatusOK, models.PaymentResponse{
				ID:     invalidPayment.ID,
				Status: invalidPayment.Status,
			})
		}
		// Successful refund
		// Begin transaction
		tx, err := s.db.BeginTx(r.Context(), nil)
		defer tx.Rollback()
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
		}
		referncedPayment.Amount = referncedPayment.Amount - reqPaid.Amount
		referncedPayment, err = s.storage.SavePayment(r.Context(), tx, referncedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Successful refund")
		completedPayment, err = s.storage.SavePayment(r.Context(), tx, completedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new personal account balance and append new statement
		personalAccount, err := s.storage.GetAccountByCard(r.Context(), referncedPayment.CardNumber)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Balance = personalAccount.Balance + reqPaid.Amount
		personalAccount, err = s.storage.SaveBalance(r.Context(), tx, personalAccount, personalAccount.Balance, 0)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Statement = append(personalAccount.Statement, completedPayment.ID.String())
		personalAccount, err = s.storage.UpdateStatement(r.Context(), tx, personalAccount.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new merchant balance and append new statement
		merchant.Balance = merchant.Balance - reqPaid.Amount
		merchant, err = s.storage.SaveBalance(r.Context(), tx, merchant, merchant.Balance, 0)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		merchant, err = s.storage.UpdateStatement(r.Context(), tx, merchant.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// Commit transaction
		if err := tx.Commit(); err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
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
	// paid request
	reqPaid := &models.PaidRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPaid); err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	defer r.Body.Close()
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
			// Begin transaction
			tx, err := s.db.BeginTx(r.Context(), nil)
			defer tx.Rollback()
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
			}
			completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Invalid amount")
			invalidPayment, err := s.storage.SavePayment(r.Context(), tx, completedPayment)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			merchant.Statement = append(merchant.Statement, invalidPayment.ID.String())
			merchant, err = s.storage.UpdateStatement(r.Context(), tx, merchant.ID, invalidPayment.ID)
			if err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			}
			// Commit transaction
			if err := tx.Commit(); err != nil {
				return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
			} 
			return WriteJSON(w, http.StatusOK, models.PaymentResponse{
				ID:     invalidPayment.ID,
				Status: invalidPayment.Status,
			})
		}
		// Successful refund
		// Begin transaction
		tx, err := s.db.BeginTx(r.Context(), nil)
		defer tx.Rollback()
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
		}
		referncedPayment.Amount = referncedPayment.Amount - reqPaid.Amount
		referncedPayment, err = s.storage.SavePayment(r.Context(), tx, referncedPayment)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		completedPayment := models.CreateCompletePayment(reqPaid, referncedPayment, "Successful cancel")
		completedPayment, err = s.storage.SavePayment(r.Context(), tx, completedPayment)
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
		personalAccount, err = s.storage.SaveBalance(r.Context(), tx, personalAccount, personalAccount.Balance, personalAccount.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		personalAccount.Statement = append(personalAccount.Statement, completedPayment.ID.String())
		personalAccount, err = s.storage.UpdateStatement(r.Context(), tx, personalAccount.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// update new merchant balance and append new statement
		merchant.BlockedMoney = merchant.BlockedMoney - reqPaid.Amount
		merchant, err = s.storage.SaveBalance(r.Context(), tx, merchant, 0, merchant.BlockedMoney)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		merchant.Statement = append(merchant.Statement, completedPayment.ID.String())
		merchant, err = s.storage.UpdateStatement(r.Context(), tx, merchant.ID, completedPayment.ID)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
		// Commit transaction
		if err := tx.Commit(); err != nil {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "wrong transaction"})
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