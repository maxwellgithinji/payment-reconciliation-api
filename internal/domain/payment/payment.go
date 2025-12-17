package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain"
)

// PaymentStatus represents the current state of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// PaymentMethod represents how the payment was made
type PaymentMethod string

const (
	PaymentMethodMobileMoney  PaymentMethod = "mobile_money"
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// Payment represents a payment transaction
type Payment struct {
	ID               uuid.UUID
	PolicyID         uuid.UUID
	CustomerID       uuid.UUID
	Amount           domain.Money
	PaymentMethod    PaymentMethod
	Status           PaymentStatus
	TransactionID    string // External payment provider transaction ID
	PaymentReference string // Internal reference number
	PaymentDate      time.Time
	ProcessedAt      *time.Time
	FailureReason    string
	IsReconciled     bool
	ReconciliationID *uuid.UUID
	AuditInfo        domain.AuditInfo
}

// NewPayment creates a new payment transaction
func NewPayment(
	policyID, customerID uuid.UUID,
	amount domain.Money,
	paymentMethod PaymentMethod,
	transactionID string,
	createdBy string,
) (*Payment, error) {
	if err := validatePaymentInput(policyID, customerID, amount, paymentMethod, transactionID); err != nil {
		return nil, err
	}

	return &Payment{
		ID:               uuid.New(),
		PolicyID:         policyID,
		CustomerID:       customerID,
		Amount:           amount,
		PaymentMethod:    paymentMethod,
		Status:           PaymentStatusPending,
		TransactionID:    transactionID,
		PaymentReference: generatePaymentReference(),
		PaymentDate:      time.Now(),
		IsReconciled:     false,
		AuditInfo:        domain.NewAuditInfo(createdBy),
	}, nil
}

// Complete marks the payment as completed
func (p *Payment) Complete(updatedBy string) error {
	if p.Status != PaymentStatusPending {
		return errors.New("only pending payments can be completed")
	}

	now := time.Now()
	p.Status = PaymentStatusCompleted
	p.ProcessedAt = &now
	p.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// Fail marks the payment as failed
func (p *Payment) Fail(reason string, updatedBy string) error {
	if p.Status != PaymentStatusPending {
		return errors.New("only pending payments can be failed")
	}

	now := time.Now()
	p.Status = PaymentStatusFailed
	p.ProcessedAt = &now
	p.FailureReason = reason
	p.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// Refund processes a refund for the payment
func (p *Payment) Refund(updatedBy string) error {
	if p.Status != PaymentStatusCompleted {
		return errors.New("only completed payments can be refunded")
	}

	p.Status = PaymentStatusRefunded
	p.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// MarkAsReconciled marks the payment as reconciled
func (p *Payment) MarkAsReconciled(reconciliationID uuid.UUID, updatedBy string) error {
	if p.Status != PaymentStatusCompleted {
		return errors.New("only completed payments can be reconciled")
	}
	if p.IsReconciled {
		return errors.New("payment already reconciled")
	}

	p.IsReconciled = true
	p.ReconciliationID = &reconciliationID
	p.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// IsCompleted checks if payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted
}

// CanBeReconciled checks if payment can be reconciled
func (p *Payment) CanBeReconciled() bool {
	return p.Status == PaymentStatusCompleted && !p.IsReconciled
}

func validatePaymentInput(
	policyID, customerID uuid.UUID,
	amount domain.Money,
	paymentMethod PaymentMethod,
	transactionID string,
) error {
	if policyID == uuid.Nil {
		return errors.New("policy ID is required")
	}
	if customerID == uuid.Nil {
		return errors.New("customer ID is required")
	}
	if amount.IsZero() || amount.Amount <= 0 {
		return errors.New("payment amount must be positive")
	}
	if !isValidPaymentMethod(paymentMethod) {
		return errors.New("invalid payment method")
	}
	if transactionID == "" {
		return errors.New("transaction ID is required")
	}

	return nil
}

func isValidPaymentMethod(method PaymentMethod) bool {
	switch method {
	case PaymentMethodMobileMoney, PaymentMethodCard, PaymentMethodBankTransfer:
		return true
	default:
		return false
	}
}

// generatePaymentReference generates a unique payment reference
func generatePaymentReference() string {
	// Format: PAY-YYYYMMDD-HHMMSS-RANDOM
	now := time.Now()
	randomPart := uuid.New().String()[:6]
	return "PAY-" + now.Format("20060102-150405") + "-" + randomPart
}

// Repository defines the interface for payment persistence
type Repository interface {
	Save(payment *Payment) error
	FindByID(id uuid.UUID) (*Payment, error)
	FindByTransactionID(transactionID string) (*Payment, error)
	FindByPolicyID(policyID uuid.UUID) ([]*Payment, error)
	FindByCustomerID(customerID uuid.UUID) ([]*Payment, error)
	FindUnreconciled() ([]*Payment, error)
	FindByStatus(status PaymentStatus) ([]*Payment, error)
	Update(payment *Payment) error
}

// Gateway defines the interface for payment provider integration
type Gateway interface {
	InitiatePayment(amount domain.Money, customerEmail, policyNumber string) (string, error)
	VerifyPayment(transactionID string) (bool, error)
	ProcessRefund(transactionID string, amount domain.Money) error
	GetPaymentStatus(transactionID string) (PaymentStatus, error)
}
