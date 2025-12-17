package commands

import (
	"errors"

	"github.com/google/uuid"
	"github.com/maxwellgithinji/payment-reconciliation-api/domain/user"
)

// CreateUserCommand represents a command to create a new user
type CreateUserCommand struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Role      user.Role
	CreatedBy string
}

// CreateUserHandler handles user creation
type CreateUserHandler struct {
	userRepo    user.Repository
	userService user.Service
}

// NewCreateUserHandler creates a new CreateUserHandler
func NewCreateUserHandler(userRepo user.Repository, userService user.Service) *CreateUserHandler {
	return &CreateUserHandler{
		userRepo:    userRepo,
		userService: userService,
	}
}

// Handle executes the create user command
func (h *CreateUserHandler) Handle(cmd CreateUserCommand) (*user.User, error) {
	// Check if user already exists
	existingUser, _ := h.userRepo.FindByEmail(cmd.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create new user
	newUser, err := user.NewUser(
		cmd.Email,
		cmd.FirstName,
		cmd.LastName,
		cmd.Role,
		cmd.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := h.userService.GeneratePasswordHash(cmd.Password)
	if err != nil {
		return nil, err
	}
	newUser.SetPasswordHash(passwordHash)

	// Save user
	if err := h.userRepo.Save(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// UpdateUserCommand represents a command to update user details
type UpdateUserCommand struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
	UpdatedBy string
}

// UpdateUserHandler handles user updates
type UpdateUserHandler struct {
	userRepo user.Repository
}

// NewUpdateUserHandler creates a new UpdateUserHandler
func NewUpdateUserHandler(userRepo user.Repository) *UpdateUserHandler {
	return &UpdateUserHandler{
		userRepo: userRepo,
	}
}

// Handle executes the update user command
func (h *UpdateUserHandler) Handle(cmd UpdateUserCommand) error {
	// Find user
	existingUser, err := h.userRepo.FindByID(cmd.UserID)
	if err != nil {
		return err
	}

	// Update profile
	if err := existingUser.UpdateProfile(cmd.FirstName, cmd.LastName, cmd.UpdatedBy); err != nil {
		return err
	}

	// Save changes
	return h.userRepo.Update(existingUser)
}

// DeactivateUserCommand represents a command to deactivate a user
type DeactivateUserCommand struct {
	UserID        uuid.UUID
	DeactivatedBy string
}

// DeactivateUserHandler handles user deactivation
type DeactivateUserHandler struct {
	userRepo user.Repository
}

// NewDeactivateUserHandler creates a new DeactivateUserHandler
func NewDeactivateUserHandler(userRepo user.Repository) *DeactivateUserHandler {
	return &DeactivateUserHandler{
		userRepo: userRepo,
	}
}

// Handle executes the deactivate user command
func (h *DeactivateUserHandler) Handle(cmd DeactivateUserCommand) error {
	existingUser, err := h.userRepo.FindByID(cmd.UserID)
	if err != nil {
		return err
	}

	existingUser.Deactivate(cmd.DeactivatedBy)
	return h.userRepo.Update(existingUser)
}
