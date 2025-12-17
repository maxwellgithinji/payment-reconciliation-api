package policy

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/domain"
)

// PaymentFrequency defines how often premium payments are due
type PaymentFrequency string

const (
	FrequencyMonthly   PaymentFrequency = "monthly"
	FrequencyQuarterly PaymentFrequency = "quarterly"
	FrequencyAnnually  PaymentFrequency = "annually"
)

// PolicyStatus represents the current state of a policy
type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusLapsed   PolicyStatus = "lapsed"
	PolicyStatusExpired  PolicyStatus = "expired"
	PolicyStatusCanceled PolicyStatus = "canceled"
)

// Policy represents an insurance policy
type Policy struct {
	ID               uuid.UUID
	PolicyNumber     string // Unique policy identifier for customers
	CustomerID       uuid.UUID
	ProductName      string
	PremiumAmount    domain.Money
	PaymentFrequency PaymentFrequency
	StartDate        time.Time
	EndDate          time.Time
	Status           PolicyStatus
	NextPaymentDue   time.Time
	AuditInfo        domain.AuditInfo
}

// NewPolicy creates a new insurance policy
func NewPolicy(
	customerID uuid.UUID,
	productName string,
	premiumAmount domain.Money,
	frequency PaymentFrequency,
	startDate, endDate time.Time,
	createdBy string,
) (*Policy, error) {
	if err := validatePolicyInput(customerID, productName, premiumAmount, frequency, startDate, endDate); err != nil {
		return nil, err
	}

	policy := &Policy{
		ID:               uuid.New(),
		PolicyNumber:     generatePolicyNumber(),
		CustomerID:       customerID,
		ProductName:      productName,
		PremiumAmount:    premiumAmount,
		PaymentFrequency: frequency,
		StartDate:        startDate,
		EndDate:          endDate,
		Status:           PolicyStatusActive,
		AuditInfo:        domain.NewAuditInfo(createdBy),
	}

	policy.calculateNextPaymentDue()
	return policy, nil
}

// RecordPayment records a successful payment and updates next payment due date
func (p *Policy) RecordPayment(paidAt time.Time, updatedBy string) error {
	if p.Status != PolicyStatusActive {
		return errors.New("cannot record payment for non-active policy")
	}

	// Update next payment due date
	p.calculateNextPaymentDue()
	p.AuditInfo.UpdatedNow(updatedBy)
	
	return nil
}

// MarkAsLapsed marks the policy as lapsed due to non-payment
func (p *Policy) MarkAsLapsed(updatedBy string) error {
	if p.Status != PolicyStatusActive {
		return errors.New("only active policies can be marked as lapsed")
	}

	p.Status = PolicyStatusLapsed
	p.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// Reinstate reinstates a lapsed policy
func (p *Policy) Reinstate(updatedBy string) error {
	if p.Status != PolicyStatusLapsed {
		return errors.New("only lapsed policies can be reinstated")
	}

	p.Status = PolicyStatusActive
	p.calculateNextPaymentDue()
	p.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// Cancel cancels the policy
func (p *Policy) Cancel(updatedBy string) error {
	if p.Status == PolicyStatusCanceled || p.Status == PolicyStatusExpired {
		return errors.New("policy already terminated")
	}

	p.Status = PolicyStatusCanceled
	p.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// CheckExpiration checks if the policy has expired and updates status
func (p *Policy) CheckExpiration() {
	if time.Now().After(p.EndDate) && p.Status == PolicyStatusActive {
		p.Status = PolicyStatusExpired
	}
}

// IsPaymentOverdue checks if a payment is overdue
func (p *Policy) IsPaymentOverdue() bool {
	return time.Now().After(p.NextPaymentDue) && p.Status == PolicyStatusActive
}

// DaysUntilPayment returns the number of days until next payment is due
func (p *Policy) DaysUntilPayment() int {
	duration := time.Until(p.NextPaymentDue)
	return int(duration.Hours() / 24)
}

// calculateNextPaymentDue calculates the next payment due date
func (p *Policy) calculateNextPaymentDue() {
	now := time.Now()
	
	// If policy hasn't started yet, next payment is start date
	if now.Before(p.StartDate) {
		p.NextPaymentDue = p.StartDate
		return
	}

	// Calculate next due date based on frequency
	var nextDue time.Time
	switch p.PaymentFrequency {
	case FrequencyMonthly:
		nextDue = now.AddDate(0, 1, 0)
	case FrequencyQuarterly:
		nextDue = now.AddDate(0, 3, 0)
	case FrequencyAnnually:
		nextDue = now.AddDate(1, 0, 0)
	}

	// Ensure next due date doesn't exceed policy end date
	if nextDue.After(p.EndDate) {
		p.NextPaymentDue = p.EndDate
	} else {
		p.NextPaymentDue = nextDue
	}
}

func validatePolicyInput(
	customerID uuid.UUID,
	productName string,
	premiumAmount domain.Money,
	frequency PaymentFrequency,
	startDate, endDate time.Time,
) error {
	if customerID == uuid.Nil {
		return errors.New("customer ID is required")
	}
	if productName == "" {
		return errors.New("product name is required")
	}
	if premiumAmount.IsZero() || premiumAmount.Amount < 0 {
		return errors.New("premium amount must be positive")
	}
	if !isValidFrequency(frequency) {
		return errors.New("invalid payment frequency")
	}
	if endDate.Before(startDate) {
		return errors.New("end date must be after start date")
	}

	return nil
}

func isValidFrequency(frequency PaymentFrequency) bool {
	switch frequency {
	case FrequencyMonthly, FrequencyQuarterly, FrequencyAnnually:
		return true
	default:
		return false
	}
}

// generatePolicyNumber generates a unique policy number
func generatePolicyNumber() string {
	// Format: POL-YYYYMMDD-RANDOM
	now := time.Now()
	randomPart := uuid.New().String()[:8]
	return "POL-" + now.Format("20060102") + "-" + randomPart
}

// Repository defines the interface for policy persistence
type Repository interface {
	Save(policy *Policy) error
	FindByID(id uuid.UUID) (*Policy, error)
	FindByPolicyNumber(policyNumber string) (*Policy, error)
	FindByCustomerID(customerID uuid.UUID) ([]*Policy, error)
	FindAll() ([]*Policy, error)
	FindOverdue() ([]*Policy, error)
	Update(policy *Policy) error
	Delete(id uuid.UUID) error
}