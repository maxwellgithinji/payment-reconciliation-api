package customer

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/domain"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`) // E.164 format
)

// Customer represents an insurance customer
type Customer struct {
	ID          uuid.UUID
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
	DateOfBirth time.Time
	Address     Address
	Status      CustomerStatus
	AuditInfo   domain.AuditInfo
}

// Address represents a customer's address
type Address struct {
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
}

// CustomerStatus represents the customer's account status
type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusSuspended CustomerStatus = "suspended"
)

// NewCustomer creates a new customer
func NewCustomer(
	firstName, lastName, email, phoneNumber string,
	dateOfBirth time.Time,
	address Address,
	createdBy string,
) (*Customer, error) {
	if err := validateCustomerInput(firstName, lastName, email, phoneNumber, dateOfBirth); err != nil {
		return nil, err
	}

	return &Customer{
		ID:          uuid.New(),
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		PhoneNumber: phoneNumber,
		DateOfBirth: dateOfBirth,
		Address:     address,
		Status:      CustomerStatusActive,
		AuditInfo:   domain.NewAuditInfo(createdBy),
	}, nil
}

// UpdateContactInfo updates customer contact information
func (c *Customer) UpdateContactInfo(email, phoneNumber string, updatedBy string) error {
	if email != "" && !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	if phoneNumber != "" && !phoneRegex.MatchString(phoneNumber) {
		return errors.New("invalid phone number format")
	}

	if email != "" {
		c.Email = email
	}
	if phoneNumber != "" {
		c.PhoneNumber = phoneNumber
	}

	c.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// UpdateAddress updates customer address
func (c *Customer) UpdateAddress(address Address, updatedBy string) {
	c.Address = address
	c.AuditInfo.UpdatedNow(updatedBy)
}

// Suspend suspends the customer account
func (c *Customer) Suspend(suspendedBy string) error {
	if c.Status == CustomerStatusSuspended {
		return errors.New("customer already suspended")
	}

	c.Status = CustomerStatusSuspended
	c.AuditInfo.UpdatedNow(suspendedBy)
	return nil
}

// Activate activates the customer account
func (c *Customer) Activate(activatedBy string) error {
	if c.Status == CustomerStatusActive {
		return errors.New("customer already active")
	}

	c.Status = CustomerStatusActive
	c.AuditInfo.UpdatedNow(activatedBy)
	return nil
}

// FullName returns the customer's full name
func (c *Customer) FullName() string {
	return c.FirstName + " " + c.LastName
}

// Age calculates the customer's age
func (c *Customer) Age() int {
	now := time.Now()
	age := now.Year() - c.DateOfBirth.Year()
	
	if now.Month() < c.DateOfBirth.Month() ||
		(now.Month() == c.DateOfBirth.Month() && now.Day() < c.DateOfBirth.Day()) {
		age--
	}
	
	return age
}

// IsActive checks if the customer is active
func (c *Customer) IsActive() bool {
	return c.Status == CustomerStatusActive
}

func validateCustomerInput(firstName, lastName, email, phoneNumber string, dateOfBirth time.Time) error {
	if firstName == "" {
		return errors.New("first name is required")
	}
	if lastName == "" {
		return errors.New("last name is required")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	if !phoneRegex.MatchString(phoneNumber) {
		return errors.New("invalid phone number format")
	}
	if dateOfBirth.After(time.Now()) {
		return errors.New("date of birth cannot be in the future")
	}
	if time.Since(dateOfBirth).Hours()/24/365 < 18 {
		return errors.New("customer must be at least 18 years old")
	}

	return nil
}

// Repository defines the interface for customer persistence
type Repository interface {
	Save(customer *Customer) error
	FindByID(id uuid.UUID) (*Customer, error)
	FindByEmail(email string) (*Customer, error)
	FindAll() ([]*Customer, error)
	Update(customer *Customer) error
	Delete(id uuid.UUID) error
}