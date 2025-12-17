package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maxwellgithinji/payment-reconciliation-api/application/commands"
	"github.com/maxwellgithinji/payment-reconciliation-api/application/queries"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/reconciliation"
	"github.com/maxwellgithinji/payment-reconciliation-api/interfaces/dto"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	processPaymentHandler *commands.ProcessPaymentHandler
	dashboardHandler      *queries.PaymentTrackingDashboardHandler
	reconReportHandler    *queries.ReconciliationReportHandler
	manualReconHandler    *commands.ManualReconciliationHandler
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(
	processPaymentHandler *commands.ProcessPaymentHandler,
	dashboardHandler *queries.PaymentTrackingDashboardHandler,
	reconReportHandler *queries.ReconciliationReportHandler,
	manualReconHandler *commands.ManualReconciliationHandler,
) *PaymentHandler {
	return &PaymentHandler{
		processPaymentHandler: processPaymentHandler,
		dashboardHandler:      dashboardHandler,
		reconReportHandler:    reconReportHandler,
		manualReconHandler:    manualReconHandler,
	}
}

// ProcessPayment handles POST /api/payments
// This endpoint receives payment notifications from the payment gateway
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req dto.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userID := "system" // Payment processing is automated
	if id, ok := r.Context().Value("user_id").(string); ok {
		userID = id
	}

	cmd := commands.ProcessPaymentCommand{
		PolicyID:      req.PolicyID,
		CustomerID:    req.CustomerID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		TransactionID: req.TransactionID,
		CreatedBy:     userID,
	}

	pmt, recon, err := h.processPaymentHandler.Handle(cmd)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to process payment", err)
		return
	}

	response := map[string]interface{}{
		"payment": mapPaymentToDTO(pmt),
	}

	if recon != nil {
		response["reconciliation"] = mapReconciliationToDTO(recon)
		if recon.IsFullyReconciled() {
			response["message"] = "Payment processed and automatically reconciled"
		} else {
			response["message"] = "Payment processed but requires manual reconciliation"
		}
	} else {
		response["message"] = "Payment processed"
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetDashboard handles GET /api/dashboard
func (h *PaymentHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	query := queries.PaymentTrackingDashboardQuery{}
	dashboard, err := h.dashboardHandler.Handle(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch dashboard data", err)
		return
	}

	response := dto.DashboardResponse{
		TotalPolicies:     dashboard.TotalPolicies,
		ActivePolicies:    dashboard.ActivePolicies,
		LapsedPolicies:    dashboard.LapsedPolicies,
		PaidCount:         dashboard.PaidCount,
		DueCount:          dashboard.DueCount,
		OverdueCount:      dashboard.OverdueCount,
		TotalPremiumDue:   dto.MoneyDTO{Amount: dashboard.TotalPremiumDue, Currency: "KES"},
		TotalPremiumPaid:  dto.MoneyDTO{Amount: dashboard.TotalPremiumPaid, Currency: "KES"},
		RecentPayments:    mapPaymentsToDTOs(dashboard.RecentPayments),
		OverduePolicies:   mapPoliciesToDTOs(dashboard.OverduePolicies),
		UnreconciledCount: dashboard.UnreconciledCount,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetReconciliationReport handles GET /api/reconciliation/report
func (h *PaymentHandler) GetReconciliationReport(w http.ResponseWriter, r *http.Request) {
	query := queries.ReconciliationReportQuery{}

	// Parse optional query parameters
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := reconciliation.ReconciliationStatus(statusStr)
		query.Status = &status
	}

	report, err := h.reconReportHandler.Handle(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate report", err)
		return
	}

	response := dto.ReconciliationReportResponse{
		TotalReconciliations: report.TotalReconciliations,
		AutoReconciled:       report.AutoReconciled,
		ManuallyReconciled:   report.ManuallyReconciled,
		RequiringReview:      report.RequiringReview,
		TotalMatched:         dto.MoneyDTO{Amount: report.TotalMatched, Currency: "KES"},
		TotalUnmatched:       dto.MoneyDTO{Amount: report.TotalUnmatched, Currency: "KES"},
		Reconciliations:      mapReconciliationDetailsToDTOs(report.Reconciliations),
	}

	respondJSON(w, http.StatusOK, response)
}

// ManualReconciliation handles POST /api/reconciliation/{id}/review
func (h *PaymentHandler) ManualReconciliation(w http.ResponseWriter, r *http.Request) {
	reconID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid reconciliation ID", err)
		return
	}

	var req dto.ManualReconciliationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	cmd := commands.ManualReconciliationCommand{
		ReconciliationID: reconID,
		Status:           reconciliation.ReconciliationStatus(req.Status),
		Notes:            req.Notes,
		ReviewedBy:       userID,
	}

	if err := h.manualReconHandler.Handle(cmd); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to review reconciliation", err)
		return
	}

	respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Reconciliation reviewed successfully",
	})
}
