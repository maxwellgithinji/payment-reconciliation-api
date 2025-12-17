package domain

import (
	"errors"
	"time"
)

// Common domain errors
var (
	ErrNotFound            = errors.New("resource not found")
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrAlreadyExists       = errors.New("resource already exists")
	ErrInvalidState        = errors.New("invalid state transition")
	ErrConcurrencyConflict = errors.New("concurrency conflict")
)

// Money represents a monetary value in the smallest currency unit (e.g., cents)
type Money struct {
	Amount   int64  // Amount in smallest unit (e.g., cents)
	Currency string // ISO 4217 currency code (e.g., "KES", "USD")
}

// NewMoney creates a new Money value object
func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

// IsZero checks if the amount is zero
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// Equals checks if two Money values are equal
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// Add adds two Money values (must be same currency)
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, errors.New("cannot add different currencies")
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// AuditInfo tracks creation and modification metadata
type AuditInfo struct {
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time
	UpdatedBy string
}

// NewAuditInfo creates a new AuditInfo
func NewAuditInfo(userID string) AuditInfo {
	now := time.Now()
	return AuditInfo{
		CreatedAt: now,
		CreatedBy: userID,
		UpdatedAt: now,
		UpdatedBy: userID,
	}
}

// UpdatedBy records an update
func (a *AuditInfo) UpdatedNow(userID string) {
	a.UpdatedAt = time.Now()
	a.UpdatedBy = userID
}
