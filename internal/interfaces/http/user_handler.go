package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/insurance-portal/poc/internal/application/commands"
	"github.com/insurance-portal/poc/internal/domain/user"
	"github.com/insurance-portal/poc/internal/interfaces/dto"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	createUserHandler     *commands.CreateUserHandler
	updateUserHandler     *commands.UpdateUserHandler
	deactivateUserHandler *commands.DeactivateUserHandler
	userRepo              user.Repository
	userService           user.Service
	jwtSecret             []byte
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	createUserHandler *commands.CreateUserHandler,
	updateUserHandler *commands.UpdateUserHandler,
	deactivateUserHandler *commands.DeactivateUserHandler,
	userRepo user.Repository,
	userService user.Service,
	jwtSecret string,
) *UserHandler {
	return &UserHandler{
		createUserHandler:     createUserHandler,
		updateUserHandler:     updateUserHandler,
		deactivateUserHandler: deactivateUserHandler,
		userRepo:              userRepo,
		userService:           userService,
		jwtSecret:             []byte(jwtSecret),
	}
}

// Login handles POST /api/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate credentials
	usr, err := h.userService.ValidateCredentials(req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	if !usr.IsActive {
		respondError(w, http.StatusForbidden, "User account is deactivated", nil)
		return
	}

	// Generate JWT token
	expiresAt := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": usr.ID.String(),
		"email":   usr.Email,
		"role":    string(usr.Role),
		"exp":     expiresAt.Unix(),
	})

	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	response := dto.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User:      mapUserToDTO(usr),
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	createdBy, _ := r.Context().Value("user_id").(string)

	cmd := commands.CreateUserCommand{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		CreatedBy: createdBy,
	}

	usr, err := h.createUserHandler.Handle(cmd)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to create user", err)
		return
	}

	respondJSON(w, http.StatusCreated, mapUserToDTO(usr))
}

// GetUser handles GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	usr, err := h.userRepo.FindByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found", err)
		return
	}

	respondJSON(w, http.StatusOK, mapUserToDTO(usr))
}

// ListUsers handles GET /api/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.FindAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch users", err)
		return
	}

	userDTOs := make([]dto.UserDTO, len(users))
	for i, usr := range users {
		userDTOs[i] = mapUserToDTO(usr)
	}

	respondJSON(w, http.StatusOK, userDTOs)
}

// UpdateUser handles PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	updatedBy, _ := r.Context().Value("user_id").(string)

	cmd := commands.UpdateUserCommand{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UpdatedBy: updatedBy,
	}

	if err := h.updateUserHandler.Handle(cmd); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to update user", err)
		return
	}

	usr, _ := h.userRepo.FindByID(userID)
	respondJSON(w, http.StatusOK, mapUserToDTO(usr))
}

// DeactivateUser handles POST /api/users/{id}/deactivate
func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	deactivatedBy, _ := r.Context().Value("user_id").(string)

	cmd := commands.DeactivateUserCommand{
		UserID:        userID,
		DeactivatedBy: deactivatedBy,
	}

	if err := h.deactivateUserHandler.Handle(cmd); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to deactivate user", err)
		return
	}

	respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "User deactivated successfully",
	})
}