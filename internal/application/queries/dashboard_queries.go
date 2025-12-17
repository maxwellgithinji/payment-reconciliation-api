package queries

import (
	"time"

	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/domain/customer"
	"github.com/insurance-portal/poc/internal/domain/payment"
	"github.com/insurance-portal/poc/internal/domain/policy"
	"github.com/insurance-portal/poc/internal/domain/reconciliation"
)

// PaymentTrackingDashboardQuery represents a query for the payment tracking dashboard
type PaymentTrackingDashboardQuery struct{}

// PaymentTrackingDashboard represents the dashboard data
type PaymentTrackingDashboard struct {
	TotalPolicies       int
	ActivePolicies      int
	LapsedPolicies      int
	PaidCount           int
	DueCount            int
	OverdueCount        int
	TotalPremiumDue     int64
	TotalPremiumPaid    int64
	RecentPayments      []*payment.Payment
	OverduePolicies     []*policy.Policy
	UnreconciledCount   int
}

// PaymentTrackingDashboardHandler handles dashboard queries
type PaymentTrackingDashboardHandler struct {
	policyRepo         policy.Repository
	paymentRepo        payment.Repository
	reconciliationRepo reconciliation.Repository
}

// NewPaymentTrackingDashboardHandler creates a new handler
func NewPaymentTrackingDashboardHandler(
	policyRepo policy.Repository,
	paymentRepo payment.Repository,
	reconciliationRepo reconciliation.Repository,
) *PaymentTrackingDashboardHandler {
	return &PaymentTrackingDashboardHandler{
		policyRepo:         policyRepo,
		paymentRepo:        paymentRepo,
		reconciliationRepo: reconciliationRepo,
	}
}

// Handle executes the dashboard query
func (h *PaymentTrackingDashboardHandler) Handle(query PaymentTrackingDashboardQuery) (*PaymentTrackingDashboard, error) {
	dashboard := &PaymentTrackingDashboard{}

	// Get all policies
	allPolicies, err := h.policyRepo.FindAll()
	if err != nil {
		return nil, err
	}

	dashboard.TotalPolicies = len(allPolicies)

	// Calculate policy statistics
	now := time.Now()
	var activePolicies, lapsedPolicies, paidCount, dueCount, overdueCount int
	var totalDue, totalPaid int64
	var overduePolicies []*policy.Policy

	for _, pol := range allPolicies {
		// Check expiration
		pol.CheckExpiration()

		switch pol.Status {
		case policy.PolicyStatusActive:
			activePolicies++
			
			// Check payment status
			if pol.IsPaymentOverdue() {
				overdueCount++
				overduePolicies = append(overduePolicies, pol)
				totalDue += pol.PremiumAmount.Amount
			} else if pol.NextPaymentDue.After(now) {
				dueCount++
			}
			
		case policy.PolicyStatusLapsed:
			lapsedPolicies++
		}
	}

	dashboard.ActivePolicies = activePolicies
	dashboard.LapsedPolicies = lapsedPolicies
	dashboard.DueCount = dueCount
	dashboard.OverdueCount = overdueCount
	dashboard.TotalPremiumDue = totalDue
	dashboard.OverduePolicies = overduePolicies

	// Get completed payments count
	completedPayments, err := h.paymentRepo.FindByStatus(payment.PaymentStatusCompleted)
	if err == nil {
		dashboard.PaidCount = len(completedPayments)
		for _, pmt := range completedPayments {
			totalPaid += pmt.Amount.Amount
		}
		dashboard.TotalPremiumPaid = totalPaid
	}

	// Get unreconciled payments count
	unreconciledPayments, err := h.paymentRepo.FindUnreconciled()
	if err == nil {
		dashboard.UnreconciledCount = len(unreconciledPayments)
	}

	// Get recent payments (last 10)
	// Note: This would ideally be a separate query method with limit
	dashboard.RecentPayments = completedPayments
	if len(completedPayments) > 10 {
		dashboard.RecentPayments = completedPayments[:10]
	}

	return dashboard, nil
}

// ReconciliationReportQuery represents a query for reconciliation report
type ReconciliationReportQuery struct {
	StartDate *time.Time
	EndDate   *time.Time
	Status    *reconciliation.ReconciliationStatus
}

// ReconciliationReport represents the reconciliation report data
type ReconciliationReport struct {
	TotalReconciliations    int
	AutoReconciled          int
	ManuallyReconciled      int
	RequiringReview         int
	TotalMatched            int64
	TotalUnmatched          int64
	Reconciliations         []*ReconciliationDetail
}

// ReconciliationDetail provides detailed reconciliation information
type ReconciliationDetail struct {
	Reconciliation *reconciliation.Reconciliation
	Payment        *payment.Payment
	Policy         *policy.Policy
	Customer       *customer.Customer
}

// ReconciliationReportHandler handles reconciliation report queries
type ReconciliationReportHandler struct {
	reconciliationRepo reconciliation.Repository
	paymentRepo        payment.Repository
	policyRepo         policy.Repository
	customerRepo       customer.Repository
}

// NewReconciliationReportHandler creates a new handler
func NewReconciliationReportHandler(
	reconciliationRepo reconciliation.Repository,
	paymentRepo payment.Repository,
	policyRepo policy.Repository,
	customerRepo customer.Repository,
) *ReconciliationReportHandler {
	return &ReconciliationReportHandler{
		reconciliationRepo: reconciliationRepo,
		paymentRepo:        paymentRepo,
		policyRepo:         policyRepo,
		customerRepo:       customerRepo,
	}
}

// Handle executes the reconciliation report query
func (h *ReconciliationReportHandler) Handle(query ReconciliationReportQuery) (*ReconciliationReport, error) {
	report := &ReconciliationReport{}

	// Get reconciliations based on filters
	var reconciliations []*reconciliation.Reconciliation
	var err error

	if query.Status != nil {
		reconciliations, err = h.reconciliationRepo.FindByStatus(*query.Status)
	} else if query.StartDate != nil && query.EndDate != nil {
		reconciliations, err = h.reconciliationRepo.FindByDateRange(*query.StartDate, *query.EndDate)
	} else {
		// Get requiring review by default
		reconciliations, err = h.reconciliationRepo.FindRequiringReview()
	}

	if err != nil {
		return nil, err
	}

	report.TotalReconciliations = len(reconciliations)

	// Build detailed report
	for _, recon := range reconciliations {
		detail := &ReconciliationDetail{
			Reconciliation: recon,
		}

		// Get payment
		pmt, err := h.paymentRepo.FindByID(recon.PaymentID)
		if err == nil {
			detail.Payment = pmt
		}

		// Get policy
		pol, err := h.policyRepo.FindByID(recon.PolicyID)
		if err == nil {
			detail.Policy = pol

			// Get customer
			cust, err := h.customerRepo.FindByID(pol.CustomerID)
			if err == nil {
				detail.Customer = cust
			}
		}

		report.Reconciliations = append(report.Reconciliations, detail)

		// Update statistics
		if recon.AutoReconciled {
			report.AutoReconciled++
		} else if recon.ManuallyReviewed {
			report.ManuallyReconciled++
		}

		if recon.RequiresManualReview() {
			report.RequiringReview++
		}

		if recon.Status == reconciliation.ReconciliationStatusMatched {
			report.TotalMatched += recon.ReceivedAmount.Amount
		} else {
			report.TotalUnmatched += recon.ReceivedAmount.Amount
		}
	}

	return report, nil
}

// PolicyDetailQuery represents a query for policy details
type PolicyDetailQuery struct {
	PolicyID uuid.UUID
}

// PolicyDetail represents detailed policy information
type PolicyDetail struct {
	Policy   *policy.Policy
	Customer *customer.Customer
	Payments []*payment.Payment
}

// PolicyDetailHandler handles policy detail queries
type PolicyDetailHandler struct {
	policyRepo   policy.Repository
	customerRepo customer.Repository
	paymentRepo  payment.Repository
}

// NewPolicyDetailHandler creates a new handler
func NewPolicyDetailHandler(
	policyRepo policy.Repository,
	customerRepo customer.Repository,
	paymentRepo payment.Repository,
) *PolicyDetailHandler {
	return &PolicyDetailHandler{
		policyRepo:   policyRepo,
		customerRepo: customerRepo,
		paymentRepo:  paymentRepo,
	}
}

// Handle executes the policy detail query
func (h *PolicyDetailHandler) Handle(query PolicyDetailQuery) (*PolicyDetail, error) {
	detail := &PolicyDetail{}

	// Get policy
	pol, err := h.policyRepo.FindByID(query.PolicyID)
	if err != nil {
		return nil, err
	}
	detail.Policy = pol

	// Get customer
	cust, err := h.customerRepo.FindByID(pol.CustomerID)
	if err != nil {
		return nil, err
	}
	detail.Customer = cust

	// Get payments
	payments, err := h.paymentRepo.FindByPolicyID(query.PolicyID)
	if err != nil {
		return nil, err
	}
	detail.Payments = payments

	return detail, nil
}