package user

import (
	"errors"

	"github.com/google/uuid"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain"
)

// Role represents user roles in the system
type Role string

const (
	RoleAdmin           Role = "admin"
	RoleAccountsManager Role = "accounts_manager"
)

// User represents a system user
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         Role
	IsActive     bool
	AuditInfo    domain.AuditInfo
}

// NewUser creates a new user
func NewUser(email, firstName, lastName string, role Role, createdBy string) (*User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if firstName == "" {
		return nil, errors.New("first name is required")
	}
	if lastName == "" {
		return nil, errors.New("last name is required")
	}
	if !isValidRole(role) {
		return nil, errors.New("invalid role")
	}

	return &User{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		IsActive:  true,
		AuditInfo: domain.NewAuditInfo(createdBy),
	}, nil
}

// SetPasswordHash sets the hashed password
func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = hash
}

// Deactivate deactivates the user
func (u *User) Deactivate(deactivatedBy string) {
	u.IsActive = false
	u.AuditInfo.UpdatedNow(deactivatedBy)
}

// Activate activates the user
func (u *User) Activate(activatedBy string) {
	u.IsActive = true
	u.AuditInfo.UpdatedNow(activatedBy)
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(firstName, lastName string, updatedBy string) error {
	if firstName == "" {
		return errors.New("first name is required")
	}
	if lastName == "" {
		return errors.New("last name is required")
	}

	u.FirstName = firstName
	u.LastName = lastName
	u.AuditInfo.UpdatedNow(updatedBy)
	return nil
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// CanManagePolicies checks if user can manage policies
func (u *User) CanManagePolicies() bool {
	return u.Role == RoleAdmin || u.Role == RoleAccountsManager
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin
}

func isValidRole(role Role) bool {
	switch role {
	case RoleAdmin, RoleAccountsManager:
		return true
	default:
		return false
	}
}

// Repository defines the interface for user persistence
type Repository interface {
	Save(user *User) error
	FindByID(id uuid.UUID) (*User, error)
	FindByEmail(email string) (*User, error)
	FindAll() ([]*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
}

// Service defines domain operations for users
type Service interface {
	ValidateCredentials(email, password string) (*User, error)
	GeneratePasswordHash(password string) (string, error)
	ValidatePassword(hash, password string) error
}
