package service

import (
	"fmt"

	"github.com/SerKKiT/streaming-platform/auth-service/internal/models"
	"github.com/SerKKiT/streaming-platform/auth-service/internal/repository"
	"github.com/SerKKiT/streaming-platform/auth-service/pkg/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user account
func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Check if username already exists
	existingUser, _ := s.userRepo.GetUserByUsername(req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingEmail, _ := s.userRepo.GetUserByEmail(req.Email)
	if existingEmail != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.userRepo.CreateUser(req.Username, req.Email, passwordHash)
	if err != nil {
		return nil, err
	}

	// Generate JWT token (без duration параметра)
	token, expiresAt, err := utils.GenerateToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token (без duration параметра)
	token, expiresAt, err := utils.GenerateToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// GetProfile returns user profile information
func (s *AuthService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Don't return password hash
	user.PasswordHash = ""

	return user, nil
}

// UpdateProfile updates user profile information
func (s *AuthService) UpdateProfile(userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	// Get current user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if new username is already taken by another user
	if req.Username != "" && req.Username != user.Username {
		existingUser, _ := s.userRepo.GetUserByUsername(req.Username)
		if existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("username already taken")
		}
		user.Username = req.Username
	}

	// Check if new email is already taken by another user
	if req.Email != "" && req.Email != user.Email {
		existingEmail, _ := s.userRepo.GetUserByEmail(req.Email)
		if existingEmail != nil && existingEmail.ID != userID {
			return nil, fmt.Errorf("email already taken")
		}
		user.Email = req.Email
	}

	// Update user
	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	// Don't return password hash
	user.PasswordHash = ""

	return user, nil
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID uuid.UUID, req *models.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify old password
	if !utils.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(userID, newPasswordHash); err != nil {
		return err
	}

	return nil
}
