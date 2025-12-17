package reconciliation

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/domain"
	"github.com/insurance-portal/poc/internal/domain/payment"
	"github.com/insurance-portal/poc/internal/domain/policy"
)

// ReconciliationStatus represents the status of a reconciliation attempt
type ReconciliationStatus string

const (
	ReconciliationStatusMatched    ReconciliationStatus = "matched"
	ReconciliationStatusUnmatched  ReconciliationStatus = "unmatched"
	ReconciliationStatusPartial    ReconciliationStatus = "partial"
	ReconciliationStatusDisputed   ReconciliationStatus = "disputed"
)

// MatchResult represents the outcome of a reconciliation match
type MatchResult string

const (
	MatchResultPerfect      MatchResult = "perfect"       // Exact match
	MatchResultAmountMismatch MatchResult = "amount_mismatch" // Wrong amount
	MatchResultPolicyMismatch MatchResult = "policy_mismatch" // Wrong policy
	MatchResultNoMatch        MatchResult = "no_match"        // No match found
)

// Reconciliation represents a payment reconciliation record
type Reconciliation struct {
	ID              uuid.UUID
	PaymentID       uuid.UUID
	PolicyID        uuid.UUID
	ExpectedAmount  domain.Money
	ReceivedAmount  domain.Money
	Status          ReconciliationStatus
	MatchResult     MatchResult
	ReconciledAt    time.Time
	Notes           string
	Variance        domain.Money // Difference between expected and received
	AutoReconciled  bool         // True if automatically reconciled
	ManuallyReviewed bool
	ReviewedBy      *string
	ReviewedAt      *time.Time
	AuditInfo       domain.AuditInfo
}

// ReconciliationRule defines rules for automatic reconciliation
type ReconciliationRule struct {
	AllowedVariancePercentage float64 // Percentage variance allowed for auto-reconciliation
	RequireExactMatch         bool    // If true, no variance is allowed
	GracePeriodDays           int     // Days before marking as unmatched
}

// DefaultReconciliationRule returns the default reconciliation rules
func DefaultReconciliationRule() ReconciliationRule {
	return ReconciliationRule{
		AllowedVariancePercentage: 0.0, // No variance by default
		RequireExactMatch:         true,
		GracePeriodDays:           3,
	}
}

// NewReconciliation creates a new reconciliation record
func NewReconciliation(
	paymentID, policyID uuid.UUID,
	expectedAmount, receivedAmount domain.Money,
	createdBy string,
) (*Reconciliation, error) {
	if paymentID == uuid.Nil {
		return nil, errors.New("payment ID is required")
	}
	if policyID == uuid.Nil {
		return nil, errors.New("policy ID is required")
	}
	if expectedAmount.Currency != receivedAmount.Currency {
		return nil, errors.New("currency mismatch between expected and received amounts")
	}

	variance := domain.Money{
		Amount:   receivedAmount.Amount - expectedAmount.Amount,
		Currency: expectedAmount.Currency,
	}

	return &Reconciliation{
		ID:             uuid.New(),
		PaymentID:      paymentID,
		PolicyID:       policyID,
		ExpectedAmount: expectedAmount,
		ReceivedAmount: receivedAmount,
		Variance:       variance,
		ReconciledAt:   time.Now(),
		AutoReconciled: false,
		AuditInfo:      domain.NewAuditInfo(createdBy),
	}, nil
}

// AttemptAutoReconciliation attempts to automatically reconcile the payment
func (r *Reconciliation) AttemptAutoReconciliation(rule ReconciliationRule) error {
	// Check for exact match
	if r.ExpectedAmount.Equals(r.ReceivedAmount) {
		r.Status = ReconciliationStatusMatched
		r.MatchResult = MatchResultPerfect
		r.AutoReconciled = true
		return nil
	}

	// If exact match required, mark as unmatched
	if rule.RequireExactMatch {
		r.Status = ReconciliationStatusUnmatched
		r.MatchResult = MatchResultAmountMismatch
		r.Notes = "Amount mismatch - exact match required"
		return errors.New("exact match required but amounts differ")
	}

	// Check if within allowed variance
	variancePercentage := calculateVariancePercentage(r.ExpectedAmount.Amount, r.ReceivedAmount.Amount)
	
	if variancePercentage <= rule.AllowedVariancePercentage {
		r.Status = ReconciliationStatusMatched
		r.MatchResult = MatchResultPerfect
		r.AutoReconciled = true
		r.Notes = "Matched within allowed variance"
		return nil
	}

	// Mark as partial match requiring review
	r.Status = ReconciliationStatusPartial
	r.MatchResult = MatchResultAmountMismatch
	r.Notes = "Amount variance exceeds threshold - manual review required"
	return errors.New("variance exceeds allowed threshold")
}

// ManualReview marks the reconciliation as manually reviewed
func (r *Reconciliation) ManualReview(reviewedBy string, status ReconciliationStatus, notes string) error {
	if r.ManuallyReviewed {
		return errors.New("reconciliation already manually reviewed")
	}

	now := time.Now()
	r.ManuallyReviewed = true
	r.ReviewedBy = &reviewedBy
	r.ReviewedAt = &now
	r.Status = status
	r.Notes = notes
	r.AuditInfo.UpdatedNow(reviewedBy)
	
	return nil
}

// MarkAsDisputed marks the reconciliation as disputed
func (r *Reconciliation) MarkAsDisputed(reason string, updatedBy string) error {
	r.Status = ReconciliationStatusDisputed
	r.Notes = reason
	r.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// IsFullyReconciled checks if the payment is fully reconciled
func (r *Reconciliation) IsFullyReconciled() bool {
	return r.Status == ReconciliationStatusMatched
}

// RequiresManualReview checks if manual review is needed
func (r *Reconciliation) RequiresManualReview() bool {
	return (r.Status == ReconciliationStatusUnmatched || 
	        r.Status == ReconciliationStatusPartial ||
	        r.Status == ReconciliationStatusDisputed) && 
	       !r.ManuallyReviewed
}

// GetVarianceAmount returns the absolute variance amount
func (r *Reconciliation) GetVarianceAmount() int64 {
	if r.Variance.Amount < 0 {
		return -r.Variance.Amount
	}
	return r.Variance.Amount
}

func calculateVariancePercentage(expected, received int64) float64 {
	if expected == 0 {
		return 100.0
	}
	
	variance := float64(received - expected)
	percentage := (variance / float64(expected)) * 100
	
	if percentage < 0 {
		return -percentage
	}
	return percentage
}

// ReconciliationService defines the domain service for reconciliation logic
type ReconciliationService struct {
	rule ReconciliationRule
}

// NewReconciliationService creates a new reconciliation service
func NewReconciliationService(rule ReconciliationRule) *ReconciliationService {
	return &ReconciliationService{
		rule: rule,
	}
}

// ReconcilePayment attempts to reconcile a payment against a policy
func (s *ReconciliationService) ReconcilePayment(
	pmt *payment.Payment,
	pol *policy.Policy,
	performedBy string,
) (*Reconciliation, error) {
	// Validate payment and policy match
	if pmt.PolicyID != pol.ID {
		return nil, errors.New("payment and policy IDs do not match")
	}

	// Create reconciliation record
	recon, err := NewReconciliation(
		pmt.ID,
		pol.ID,
		pol.PremiumAmount,
		pmt.Amount,
		performedBy,
	)
	if err != nil {
		return nil, err
	}

	// Attempt automatic reconciliation
	if err := recon.AttemptAutoReconciliation(s.rule); err != nil {
		// Auto-reconciliation failed, but record is still created
		// for manual review
		return recon, nil
	}

	return recon, nil
}

// Repository defines the interface for reconciliation persistence
type Repository interface {
	Save(reconciliation *Reconciliation) error
	FindByID(id uuid.UUID) (*Reconciliation, error)
	FindByPaymentID(paymentID uuid.UUID) (*Reconciliation, error)
	FindByPolicyID(policyID uuid.UUID) ([]*Reconciliation, error)
	FindUnmatched() ([]*Reconciliation, error)
	FindRequiringReview() ([]*Reconciliation, error)
	FindByStatus(status ReconciliationStatus) ([]*Reconciliation, error)
	FindByDateRange(startDate, endDate time.Time) ([]*Reconciliation, error)
	Update(reconciliation *Reconciliation) error
}