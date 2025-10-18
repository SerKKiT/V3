package service

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/SerKKiT/streaming-platform/auth-service/internal/models"
	"github.com/SerKKiT/streaming-platform/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

// Register creates a new user
func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetUserByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	existingUser, _ = s.userRepo.GetUserByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user using repository method
	user, err := s.userRepo.CreateUser(req.Username, req.Email, string(hashedPassword))
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login authenticates a user (email OR username)
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Determine identifier (email or username)
	identifier := ""
	isEmail := false

	if req.Email != "" {
		identifier = req.Email
		isEmail = true
	} else if req.Username != "" {
		identifier = req.Username
		// Check if username is actually an email
		isEmail = strings.Contains(identifier, "@")
	}

	if identifier == "" {
		return nil, errors.New("email or username is required")
	}

	// Get user from database
	var user *models.User
	var err error

	if isEmail {
		user, err = s.userRepo.GetUserByEmail(identifier)
	} else {
		user, err = s.userRepo.GetUserByUsername(identifier)
	}

	if err != nil || user == nil {
		log.Printf("❌ Login failed: user not found (identifier: %s)", identifier)
		return nil, errors.New("invalid credentials")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("❌ Login failed: invalid password (user_id: %s)", user.ID)
		return nil, errors.New("invalid credentials")
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	log.Printf("✅ Login successful: user_id=%s, username=%s", user.ID, user.Username)

	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetProfile retrieves user profile
func (s *AuthService) GetProfile(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(userID)
}

// UpdateProfile updates user profile
func (s *AuthService) UpdateProfile(userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if new username is taken
	if req.Username != "" && req.Username != user.Username {
		existingUser, _ := s.userRepo.GetUserByUsername(req.Username)
		if existingUser != nil {
			return nil, errors.New("username already taken")
		}
		user.Username = req.Username
	}

	// Check if new email is taken
	if req.Email != "" && req.Email != user.Email {
		existingUser, _ := s.userRepo.GetUserByEmail(req.Email)
		if existingUser != nil {
			return nil, errors.New("email already registered")
		}
		user.Email = req.Email
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return user, nil
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID uuid.UUID, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password using repository method
	if err := s.userRepo.UpdatePassword(userID, string(hashedPassword)); err != nil {
		return errors.New("failed to change password")
	}

	return nil
}

// generateToken creates a JWT token
func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
