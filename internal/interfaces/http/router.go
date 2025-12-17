package http

import (
	"encoding/json"
	"net/http"

	"github.com/maxwellgithinji/payment-reconciliation-api/application/queries"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/customer"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/payment"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/policy"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/reconciliation"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/user"
	"github.com/maxwellgithinji/payment-reconciliation-api/interfaces/dto"
)

// HTTP Response Helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string, err error) {
	errorResponse := dto.ErrorResponse{
		Message: message,
	}
	if err != nil {
		errorResponse.Error = err.Error()
	}
	respondJSON(w, status, errorResponse)
}

// Domain to DTO Mappers

func mapUserToDTO(u *user.User) dto.UserDTO {
	return dto.UserDTO{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		FullName:  u.FullName(),
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.AuditInfo.CreatedAt,
	}
}

func mapCustomerToDTO(c *customer.Customer) dto.CustomerDTO {
	return dto.CustomerDTO{
		ID:          c.ID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		FullName:    c.FullName(),
		Email:       c.Email,
		PhoneNumber: c.PhoneNumber,
		DateOfBirth: c.DateOfBirth,
		Age:         c.Age(),
		Address: dto.AddressDTO{
			Street:     c.Address.Street,
			City:       c.Address.City,
			State:      c.Address.State,
			PostalCode: c.Address.PostalCode,
			Country:    c.Address.Country,
		},
		Status:    string(c.Status),
		CreatedAt: c.AuditInfo.CreatedAt,
	}
}

func mapPolicyToDTO(p *policy.Policy) dto.PolicyDTO {
	return dto.PolicyDTO{
		ID:           p.ID,
		PolicyNumber: p.PolicyNumber,
		CustomerID:   p.CustomerID,
		ProductName:  p.ProductName,
		PremiumAmount: dto.MoneyDTO{
			Amount:   p.PremiumAmount.Amount,
			Currency: p.PremiumAmount.Currency,
		},
		PaymentFrequency: p.PaymentFrequency,
		StartDate:        p.StartDate,
		EndDate:          p.EndDate,
		Status:           p.Status,
		NextPaymentDue:   p.NextPaymentDue,
		DaysUntilPayment: p.DaysUntilPayment(),
		IsOverdue:        p.IsPaymentOverdue(),
		CreatedAt:        p.AuditInfo.CreatedAt,
	}
}

func mapPoliciesToDTOs(policies []*policy.Policy) []dto.PolicyDTO {
	dtos := make([]dto.PolicyDTO, len(policies))
	for i, p := range policies {
		dtos[i] = mapPolicyToDTO(p)
	}
	return dtos
}

func mapPaymentToDTO(p *payment.Payment) dto.PaymentDTO {
	return dto.PaymentDTO{
		ID:         p.ID,
		PolicyID:   p.PolicyID,
		CustomerID: p.CustomerID,
		Amount: dto.MoneyDTO{
			Amount:   p.Amount.Amount,
			Currency: p.Amount.Currency,
		},
		PaymentMethod:    p.PaymentMethod,
		Status:           p.Status,
		TransactionID:    p.TransactionID,
		PaymentReference: p.PaymentReference,
		PaymentDate:      p.PaymentDate,
		IsReconciled:     p.IsReconciled,
		CreatedAt:        p.AuditInfo.CreatedAt,
	}
}

func mapPaymentsToDTOs(payments []*payment.Payment) []dto.PaymentDTO {
	dtos := make([]dto.PaymentDTO, len(payments))
	for i, p := range payments {
		dtos[i] = mapPaymentToDTO(p)
	}
	return dtos
}

func mapReconciliationToDTO(r *reconciliation.Reconciliation) dto.ReconciliationDTO {
	return dto.ReconciliationDTO{
		ID:        r.ID,
		PaymentID: r.PaymentID,
		PolicyID:  r.PolicyID,
		ExpectedAmount: dto.MoneyDTO{
			Amount:   r.ExpectedAmount.Amount,
			Currency: r.ExpectedAmount.Currency,
		},
		ReceivedAmount: dto.MoneyDTO{
			Amount:   r.ReceivedAmount.Amount,
			Currency: r.ReceivedAmount.Currency,
		},
		Variance: dto.MoneyDTO{
			Amount:   r.Variance.Amount,
			Currency: r.Variance.Currency,
		},
		Status:         string(r.Status),
		MatchResult:    string(r.MatchResult),
		AutoReconciled: r.AutoReconciled,
		RequiresReview: r.RequiresManualReview(),
		Notes:          r.Notes,
		ReconciledAt:   r.ReconciledAt,
	}
}

func mapReconciliationDetailsToDTOs(details []*queries.ReconciliationDetail) []dto.ReconciliationDetailDTO {
	dtos := make([]dto.ReconciliationDetailDTO, len(details))
	for i, d := range details {
		detailDTO := dto.ReconciliationDetailDTO{
			Reconciliation: mapReconciliationToDTO(d.Reconciliation),
		}

		if d.Payment != nil {
			detailDTO.Payment = mapPaymentToDTO(d.Payment)
		}

		if d.Policy != nil {
			detailDTO.Policy = mapPolicyToDTO(d.Policy)
		}

		if d.Customer != nil {
			detailDTO.Customer = mapCustomerToDTO(d.Customer)
		}

		dtos[i] = detailDTO
	}
	return dtos
}

func mapMoneyToDTO(m domain.Money) dto.MoneyDTO {
	return dto.MoneyDTO{
		Amount:   m.Amount,
		Currency: m.Currency,
	}
}

func mapAddressToDomain(a dto.AddressDTO) customer.Address {
	return customer.Address{
		Street:     a.Street,
		City:       a.City,
		State:      a.State,
		PostalCode: a.PostalCode,
		Country:    a.Country,
	}
}
