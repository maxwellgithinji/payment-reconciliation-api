package commands

import (
	"errors"

	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/domain"
	"github.com/insurance-portal/poc/internal/domain/payment"
	"github.com/insurance-portal/poc/internal/domain/policy"
	"github.com/insurance-portal/poc/internal/domain/reconciliation"
)

// ProcessPaymentCommand represents a command to process a payment
type ProcessPaymentCommand struct {
	PolicyID      uuid.UUID
	CustomerID    uuid.UUID
	Amount        int64
	Currency      string
	PaymentMethod payment.PaymentMethod
	TransactionID string
	CreatedBy     string
}

// ProcessPaymentHandler handles payment processing and automatic reconciliation
type ProcessPaymentHandler struct {
	paymentRepo        payment.Repository
	policyRepo         policy.Repository
	reconciliationRepo reconciliation.Repository
	reconService       *reconciliation.ReconciliationService
}

// NewProcessPaymentHandler creates a new ProcessPaymentHandler
func NewProcessPaymentHandler(
	paymentRepo payment.Repository,
	policyRepo policy.Repository,
	reconciliationRepo reconciliation.Repository,
	reconService *reconciliation.ReconciliationService,
) *ProcessPaymentHandler {
	return &ProcessPaymentHandler{
		paymentRepo:        paymentRepo,
		policyRepo:         policyRepo,
		reconciliationRepo: reconciliationRepo,
		reconService:       reconService,
	}
}

// Handle executes the process payment command with automatic reconciliation
func (h *ProcessPaymentHandler) Handle(cmd ProcessPaymentCommand) (*payment.Payment, *reconciliation.Reconciliation, error) {
	// Verify policy exists
	pol, err := h.policyRepo.FindByID(cmd.PolicyID)
	if err != nil {
		return nil, nil, err
	}

	// Check for duplicate transaction
	existingPayment, _ := h.paymentRepo.FindByTransactionID(cmd.TransactionID)
	if existingPayment != nil {
		return nil, nil, errors.New("payment with this transaction ID already exists")
	}

	// Create payment record
	paymentAmount := domain.NewMoney(cmd.Amount, cmd.Currency)
	pmt, err := payment.NewPayment(
		cmd.PolicyID,
		cmd.CustomerID,
		paymentAmount,
		cmd.PaymentMethod,
		cmd.TransactionID,
		cmd.CreatedBy,
	)
	if err != nil {
		return nil, nil, err
	}

	// Mark payment as completed (since funds are already in bank account)
	if err := pmt.Complete(cmd.CreatedBy); err != nil {
		return nil, nil, err
	}

	// Save payment
	if err := h.paymentRepo.Save(pmt); err != nil {
		return nil, nil, err
	}

	// Attempt automatic reconciliation
	recon, err := h.reconService.ReconcilePayment(pmt, pol, cmd.CreatedBy)
	if err != nil {
		// Log error but don't fail the payment processing
		// Reconciliation can be retried or done manually
		return pmt, nil, err
	}

	// Save reconciliation record
	if err := h.reconciliationRepo.Save(recon); err != nil {
		return pmt, nil, err
	}

	// If reconciliation was successful, mark payment as reconciled
	if recon.IsFullyReconciled() {
		if err := pmt.MarkAsReconciled(recon.ID, cmd.CreatedBy); err != nil {
			return pmt, recon, err
		}
		if err := h.paymentRepo.Update(pmt); err != nil {
			return pmt, recon, err
		}

		// Update policy payment record
		if err := pol.RecordPayment(pmt.PaymentDate, cmd.CreatedBy); err != nil {
			return pmt, recon, err
		}
		if err := h.policyRepo.Update(pol); err != nil {
			return pmt, recon, err
		}
	}

	return pmt, recon, nil
}

// ManualReconciliationCommand represents a command for manual reconciliation
type ManualReconciliationCommand struct {
	ReconciliationID uuid.UUID
	Status           reconciliation.ReconciliationStatus
	Notes            string
	ReviewedBy       string
}

// ManualReconciliationHandler handles manual reconciliation review
type ManualReconciliationHandler struct {
	reconciliationRepo reconciliation.Repository
	paymentRepo        payment.Repository
	policyRepo         policy.Repository
}

// NewManualReconciliationHandler creates a new ManualReconciliationHandler
func NewManualReconciliationHandler(
	reconciliationRepo reconciliation.Repository,
	paymentRepo payment.Repository,
	policyRepo policy.Repository,
) *ManualReconciliationHandler {
	return &ManualReconciliationHandler{
		reconciliationRepo: reconciliationRepo,
		paymentRepo:        paymentRepo,
		policyRepo:         policyRepo,
	}
}

// Handle executes the manual reconciliation command
func (h *ManualReconciliationHandler) Handle(cmd ManualReconciliationCommand) error {
	// Find reconciliation record
	recon, err := h.reconciliationRepo.FindByID(cmd.ReconciliationID)
	if err != nil {
		return err
	}

	// Perform manual review
	if err := recon.ManualReview(cmd.ReviewedBy, cmd.Status, cmd.Notes); err != nil {
		return err
	}

	// Update reconciliation
	if err := h.reconciliationRepo.Update(recon); err != nil {
		return err
	}

	// If approved, update payment and policy
	if cmd.Status == reconciliation.ReconciliationStatusMatched {
		pmt, err := h.paymentRepo.FindByID(recon.PaymentID)
		if err != nil {
			return err
		}

		if err := pmt.MarkAsReconciled(recon.ID, cmd.ReviewedBy); err != nil {
			return err
		}

		if err := h.paymentRepo.Update(pmt); err != nil {
			return err
		}

		pol, err := h.policyRepo.FindByID(recon.PolicyID)
		if err != nil {
			return err
		}

		if err := pol.RecordPayment(pmt.PaymentDate, cmd.ReviewedBy); err != nil {
			return err
		}

		if err := h.policyRepo.Update(pol); err != nil {
			return err
		}
	}

	return nil
}

// RefundPaymentCommand represents a command to refund a payment
type RefundPaymentCommand struct {
	PaymentID uuid.UUID
	RefundedBy string
}

// RefundPaymentHandler handles payment refunds
type RefundPaymentHandler struct {
	paymentRepo payment.Repository
	paymentGateway payment.Gateway
}

// NewRefundPaymentHandler creates a new RefundPaymentHandler
func NewRefundPaymentHandler(paymentRepo payment.Repository, paymentGateway payment.Gateway) *RefundPaymentHandler {
	return &RefundPaymentHandler{
		paymentRepo: paymentRepo,
		paymentGateway: paymentGateway,
	}
}

// Handle executes the refund payment command
func (h *RefundPaymentHandler) Handle(cmd RefundPaymentCommand) error {
	pmt, err := h.paymentRepo.FindByID(cmd.PaymentID)
	if err != nil {
		return err
	}

	// Process refund through payment gateway
	if err := h.paymentGateway.ProcessRefund(pmt.TransactionID, pmt.Amount); err != nil {
		return err
	}

	// Update payment status
	if err := pmt.Refund(cmd.RefundedBy); err != nil {
		return err
	}

	return h.paymentRepo.Update(pmt)
}