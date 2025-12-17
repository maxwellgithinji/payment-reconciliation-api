package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/domain/payment"
	"github.com/insurance-portal/poc/internal/domain/policy"
	"github.com/insurance-portal/poc/internal/domain/user"
)

// Authentication DTOs

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserDTO   `json:"user"`
}

// User DTOs

type CreateUserRequest struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      user.Role `json:"role"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UserDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	FullName  string    `json:"full_name"`
	Role      user.Role `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Customer DTOs

type CreateCustomerRequest struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Address     AddressDTO `json:"address"`
}

type UpdateCustomerRequest struct {
	Email       string     `json:"email,omitempty"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	Address     *AddressDTO `json:"address,omitempty"`
}

type AddressDTO struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type CustomerDTO struct {
	ID          uuid.UUID  `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	FullName    string     `json:"full_name"`
	Email       string     `json:"email"`
	PhoneNumber string     `json:"phone_number"`
	DateOfBirth time.Time  `json:"date_of_birth"`
	Age         int        `json:"age"`
	Address     AddressDTO `json:"address"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Policy DTOs

type CreatePolicyRequest struct {
	CustomerID       uuid.UUID              `json:"customer_id"`
	ProductName      string                 `json:"product_name"`
	PremiumAmount    int64                  `json:"premium_amount"`
	Currency         string                 `json:"currency"`
	PaymentFrequency policy.PaymentFrequency `json:"payment_frequency"`
	StartDate        time.Time              `json:"start_date"`
	DurationMonths   int                    `json:"duration_months"`
}

type PolicyDTO struct {
	ID               uuid.UUID               `json:"id"`
	PolicyNumber     string                  `json:"policy_number"`
	CustomerID       uuid.UUID               `json:"customer_id"`
	ProductName      string                  `json:"product_name"`
	PremiumAmount    MoneyDTO                `json:"premium_amount"`
	PaymentFrequency policy.PaymentFrequency `json:"payment_frequency"`
	StartDate        time.Time               `json:"start_date"`
	EndDate          time.Time               `json:"end_date"`
	Status           policy.PolicyStatus     `json:"status"`
	NextPaymentDue   time.Time               `json:"next_payment_due"`
	DaysUntilPayment int                     `json:"days_until_payment"`
	IsOverdue        bool                    `json:"is_overdue"`
	CreatedAt        time.Time               `json:"created_at"`
}

type PolicyDetailDTO struct {
	Policy   PolicyDTO     `json:"policy"`
	Customer CustomerDTO   `json:"customer"`
	Payments []PaymentDTO  `json:"payments"`
}

// Payment DTOs

type ProcessPaymentRequest struct {
	PolicyID      uuid.UUID            `json:"policy_id"`
	CustomerID    uuid.UUID            `json:"customer_id"`
	Amount        int64                `json:"amount"`
	Currency      string               `json:"currency"`
	PaymentMethod payment.PaymentMethod `json:"payment_method"`
	TransactionID string               `json:"transaction_id"`
}

type PaymentDTO struct {
	ID               uuid.UUID             `json:"id"`
	PolicyID         uuid.UUID             `json:"policy_id"`
	CustomerID       uuid.UUID             `json:"customer_id"`
	Amount           MoneyDTO              `json:"amount"`
	PaymentMethod    payment.PaymentMethod `json:"payment_method"`
	Status           payment.PaymentStatus `json:"status"`
	TransactionID    string                `json:"transaction_id"`
	PaymentReference string                `json:"payment_reference"`
	PaymentDate      time.Time             `json:"payment_date"`
	IsReconciled     bool                  `json:"is_reconciled"`
	CreatedAt        time.Time             `json:"created_at"`
}

type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// Dashboard DTOs

type DashboardResponse struct {
	TotalPolicies     int           `json:"total_policies"`
	ActivePolicies    int           `json:"active_policies"`
	LapsedPolicies    int           `json:"lapsed_policies"`
	PaidCount         int           `json:"paid_count"`
	DueCount          int           `json:"due_count"`
	OverdueCount      int           `json:"overdue_count"`
	TotalPremiumDue   MoneyDTO      `json:"total_premium_due"`
	TotalPremiumPaid  MoneyDTO      `json:"total_premium_paid"`
	RecentPayments    []PaymentDTO  `json:"recent_payments"`
	OverduePolicies   []PolicyDTO   `json:"overdue_policies"`
	UnreconciledCount int           `json:"unreconciled_count"`
}

// Reconciliation DTOs

type ManualReconciliationRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type ReconciliationDTO struct {
	ID              uuid.UUID `json:"id"`
	PaymentID       uuid.UUID `json:"payment_id"`
	PolicyID        uuid.UUID `json:"policy_id"`
	ExpectedAmount  MoneyDTO  `json:"expected_amount"`
	ReceivedAmount  MoneyDTO  `json:"received_amount"`
	Variance        MoneyDTO  `json:"variance"`
	Status          string    `json:"status"`
	MatchResult     string    `json:"match_result"`
	AutoReconciled  bool      `json:"auto_reconciled"`
	RequiresReview  bool      `json:"requires_review"`
	Notes           string    `json:"notes"`
	ReconciledAt    time.Time `json:"reconciled_at"`
}

type ReconciliationReportResponse struct {
	TotalReconciliations int                        `json:"total_reconciliations"`
	AutoReconciled       int                        `json:"auto_reconciled"`
	ManuallyReconciled   int                        `json:"manually_reconciled"`
	RequiringReview      int                        `json:"requiring_review"`
	TotalMatched         MoneyDTO                   `json:"total_matched"`
	TotalUnmatched       MoneyDTO                   `json:"total_unmatched"`
	Reconciliations      []ReconciliationDetailDTO  `json:"reconciliations"`
}

type ReconciliationDetailDTO struct {
	Reconciliation ReconciliationDTO `json:"reconciliation"`
	Payment        PaymentDTO        `json:"payment"`
	Policy         PolicyDTO         `json:"policy"`
	Customer       CustomerDTO       `json:"customer"`
}

// Common Response DTOs

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}