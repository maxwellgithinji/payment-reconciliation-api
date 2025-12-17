package commands

import (
	"time"

	"github.com/google/uuid"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/customer"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/policy"
)

// CreatePolicyCommand represents a command to create a new policy
type CreatePolicyCommand struct {
	CustomerID       uuid.UUID
	ProductName      string
	PremiumAmount    int64
	Currency         string
	PaymentFrequency policy.PaymentFrequency
	StartDate        time.Time
	DurationMonths   int
	CreatedBy        string
}

// CreatePolicyHandler handles policy creation
type CreatePolicyHandler struct {
	policyRepo   policy.Repository
	customerRepo customer.Repository
}

// NewCreatePolicyHandler creates a new CreatePolicyHandler
func NewCreatePolicyHandler(policyRepo policy.Repository, customerRepo customer.Repository) *CreatePolicyHandler {
	return &CreatePolicyHandler{
		policyRepo:   policyRepo,
		customerRepo: customerRepo,
	}
}

// Handle executes the create policy command
func (h *CreatePolicyHandler) Handle(cmd CreatePolicyCommand) (*policy.Policy, error) {
	// Verify customer exists and is active
	cust, err := h.customerRepo.FindByID(cmd.CustomerID)
	if err != nil {
		return nil, err
	}
	if !cust.IsActive() {
		return nil, domain.ErrInvalidState
	}

	// Calculate end date based on duration
	endDate := cmd.StartDate.AddDate(0, cmd.DurationMonths, 0)

	// Create policy
	premiumAmount := domain.NewMoney(cmd.PremiumAmount, cmd.Currency)
	newPolicy, err := policy.NewPolicy(
		cmd.CustomerID,
		cmd.ProductName,
		premiumAmount,
		cmd.PaymentFrequency,
		cmd.StartDate,
		endDate,
		cmd.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Save policy
	if err := h.policyRepo.Save(newPolicy); err != nil {
		return nil, err
	}

	return newPolicy, nil
}

// CancelPolicyCommand represents a command to cancel a policy
type CancelPolicyCommand struct {
	PolicyID   uuid.UUID
	CanceledBy string
}

// CancelPolicyHandler handles policy cancellation
type CancelPolicyHandler struct {
	policyRepo policy.Repository
}

// NewCancelPolicyHandler creates a new CancelPolicyHandler
func NewCancelPolicyHandler(policyRepo policy.Repository) *CancelPolicyHandler {
	return &CancelPolicyHandler{
		policyRepo: policyRepo,
	}
}

// Handle executes the cancel policy command
func (h *CancelPolicyHandler) Handle(cmd CancelPolicyCommand) error {
	pol, err := h.policyRepo.FindByID(cmd.PolicyID)
	if err != nil {
		return err
	}

	if err := pol.Cancel(cmd.CanceledBy); err != nil {
		return err
	}

	return h.policyRepo.Update(pol)
}

// RecordPolicyPaymentCommand represents a command to record a payment
type RecordPolicyPaymentCommand struct {
	PolicyID  uuid.UUID
	PaidAt    time.Time
	UpdatedBy string
}

// RecordPolicyPaymentHandler handles recording policy payments
type RecordPolicyPaymentHandler struct {
	policyRepo policy.Repository
}

// NewRecordPolicyPaymentHandler creates a new RecordPolicyPaymentHandler
func NewRecordPolicyPaymentHandler(policyRepo policy.Repository) *RecordPolicyPaymentHandler {
	return &RecordPolicyPaymentHandler{
		policyRepo: policyRepo,
	}
}

// Handle executes the record payment command
func (h *RecordPolicyPaymentHandler) Handle(cmd RecordPolicyPaymentCommand) error {
	pol, err := h.policyRepo.FindByID(cmd.PolicyID)
	if err != nil {
		return err
	}

	if err := pol.RecordPayment(cmd.PaidAt, cmd.UpdatedBy); err != nil {
		return err
	}

	return h.policyRepo.Update(pol)
}
