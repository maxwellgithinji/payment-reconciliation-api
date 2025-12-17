package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/application/commands"
	"github.com/insurance-portal/poc/internal/application/queries"
	"github.com/insurance-portal/poc/internal/domain/customer"
	"github.com/insurance-portal/poc/internal/domain/policy"
	"github.com/insurance-portal/poc/internal/interfaces/dto"
)

// PolicyHandler handles policy-related HTTP requests
type PolicyHandler struct {
	createPolicyHandler *commands.CreatePolicyHandler
	cancelPolicyHandler *commands.CancelPolicyHandler
	policyDetailHandler *queries.PolicyDetailHandler
	policyRepo          policy.Repository
	customerRepo        customer.Repository
}

// NewPolicyHandler creates a new PolicyHandler
func NewPolicyHandler(
	createPolicyHandler *commands.CreatePolicyHandler,
	cancelPolicyHandler *commands.CancelPolicyHandler,
	policyDetailHandler *queries.PolicyDetailHandler,
	policyRepo policy.Repository,
	customerRepo customer.Repository,
) *PolicyHandler {
	return &PolicyHandler{
		createPolicyHandler: createPolicyHandler,
		cancelPolicyHandler: cancelPolicyHandler,
		policyDetailHandler: policyDetailHandler,
		policyRepo:          policyRepo,
		customerRepo:        customerRepo,
	}
}

// CreatePolicy handles POST /api/policies
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, _ := r.Context().Value("user_id").(string)

	cmd := commands.CreatePolicyCommand{
		CustomerID:       req.CustomerID,
		ProductName:      req.ProductName,
		PremiumAmount:    req.PremiumAmount,
		Currency:         req.Currency,
		PaymentFrequency: req.PaymentFrequency,
		StartDate:        req.StartDate,
		DurationMonths:   req.DurationMonths,
		CreatedBy:        userID,
	}

	pol, err := h.createPolicyHandler.Handle(cmd)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to create policy", err)
		return
	}

	respondJSON(w, http.StatusCreated, mapPolicyToDTO(pol))
}

// GetPolicy handles GET /api/policies/{id}
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	query := queries.PolicyDetailQuery{PolicyID: policyID}
	detail, err := h.policyDetailHandler.Handle(query)
	if err != nil {
		respondError(w, http.StatusNotFound, "Policy not found", err)
		return
	}

	response := dto.PolicyDetailDTO{
		Policy:   mapPolicyToDTO(detail.Policy),
		Customer: mapCustomerToDTO(detail.Customer),
		Payments: mapPaymentsToDTOs(detail.Payments),
	}

	respondJSON(w, http.StatusOK, response)
}

// ListPolicies handles GET /api/policies
func (h *PolicyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.policyRepo.FindAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch policies", err)
		return
	}

	policyDTOs := make([]dto.PolicyDTO, len(policies))
	for i, pol := range policies {
		policyDTOs[i] = mapPolicyToDTO(pol)
	}

	respondJSON(w, http.StatusOK, policyDTOs)
}

// ListCustomerPolicies handles GET /api/customers/{id}/policies
func (h *PolicyHandler) ListCustomerPolicies(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	policies, err := h.policyRepo.FindByCustomerID(customerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch policies", err)
		return
	}

	policyDTOs := make([]dto.PolicyDTO, len(policies))
	for i, pol := range policies {
		policyDTOs[i] = mapPolicyToDTO(pol)
	}

	respondJSON(w, http.StatusOK, policyDTOs)
}

// CancelPolicy handles POST /api/policies/{id}/cancel
func (h *PolicyHandler) CancelPolicy(w http.ResponseWriter, r *http.Request) {
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid policy ID", err)
		return
	}

	userID, _ := r.Context().Value("user_id").(string)

	cmd := commands.CancelPolicyCommand{
		PolicyID:   policyID,
		CanceledBy: userID,
	}

	if err := h.cancelPolicyHandler.Handle(cmd); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to cancel policy", err)
		return
	}

	respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Policy cancelled successfully",
	})
}

// ListOverduePolicies handles GET /api/policies/overdue
func (h *PolicyHandler) ListOverduePolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.policyRepo.FindOverdue()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch overdue policies", err)
		return
	}

	policyDTOs := make([]dto.PolicyDTO, len(policies))
	for i, pol := range policies {
		policyDTOs[i] = mapPolicyToDTO(pol)
	}

	respondJSON(w, http.StatusOK, policyDTOs)
}
