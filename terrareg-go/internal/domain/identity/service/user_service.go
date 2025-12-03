package service

import (
	"context"
	"errors"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// Using identity model errors instead of redefining here

// UserService manages user operations
type UserService struct {
	userRepo   repository.UserRepository
	groupRepo  repository.UserGroupRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, groupRepo repository.UserGroupRepository) *UserService {
	return &UserService{
		userRepo:  userRepo,
		groupRepo: groupRepo,
	}
}

// Authenticate authenticates a user with username/password
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	// For Phase 4, we'll implement a simple password check
	// In a full implementation, this would use proper password hashing
	if user == nil || !user.Active() {
		return nil, model.ErrInvalidCredentials
	}

	// TODO: Implement proper password verification
	// For now, return user if found (placeholder for Phase 4)
	return user, nil
}

// AuthenticateByToken authenticates a user using API key or session token
func (s *UserService) AuthenticateByToken(ctx context.Context, token string) (*model.User, error) {
	// Try to find user by API key
	user, err := s.userRepo.FindByAccessToken(ctx, token)
	if err == nil && user != nil {
		return user, nil
	}

	// For Phase 4, we could also check session tokens
	// But for now, focus on API key authentication
	return nil, model.ErrUserNotFound
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// GetUserByToken retrieves a user by authentication token
func (s *UserService) GetUserByToken(ctx context.Context, token string) (*model.User, error) {
	return s.AuthenticateByToken(ctx, token)
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, username, email, displayName string, authMethod model.AuthMethod) (*model.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, model.ErrUserAlreadyExists
	}

	existingUser, err = s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, model.ErrUserAlreadyExists
	}

	// Create new user
	user, err := model.NewUser(username, displayName, email, authMethod)
	if err != nil {
		return nil, err
	}

	// Save user
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, userID, displayName, email string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	// Update user profile
	err = user.UpdateProfile(displayName, email)
	if err != nil {
		return nil, err
	}

	// Save updated user
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeactivateUser deactivates a user
func (s *UserService) DeactivateUser(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return model.ErrUserNotFound
	}

	return user.Deactivate()
}

// AddUserToGroup adds a user to a user group
func (s *UserService) AddUserToGroup(ctx context.Context, userID, userGroupID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return model.ErrUserNotFound
	}

	userGroup, err := s.groupRepo.FindByID(ctx, userGroupID)
	if err != nil {
		return errors.New("user group not found")
	}

	return userGroup.AddMember(user)
}

// RemoveUserFromGroup removes a user from a user group
func (s *UserService) RemoveUserFromGroup(ctx context.Context, userID, userGroupID string) error {
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return model.ErrUserNotFound
	}

	userGroup, err := s.groupRepo.FindByID(ctx, userGroupID)
	if err != nil {
		return errors.New("user group not found")
	}

	return userGroup.RemoveMember(userID)
}

// CheckPermission checks if a user has a specific permission
func (s *UserService) CheckPermission(ctx context.Context, userID string, resourceType model.ResourceType, resourceID string, action model.Action) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, model.ErrUserNotFound
	}

	// Check direct user permissions
	if user.HasPermission(resourceType, resourceID, action) {
		return true, nil
	}

	// Check permissions through user groups
	userGroups, err := s.groupRepo.FindByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, userGroup := range userGroups {
		if userGroup.HasPermission(resourceType, resourceID, action) {
			return true, nil
		}
	}

	return false, nil
}

// ValidateToken validates if a token is valid for the given auth method
func (s *UserService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	user, err := s.AuthenticateByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Additional token validation logic can be added here
	if !user.Active() {
		return nil, errors.New("user is inactive")
	}

	return user, nil
}

// Logout invalidates a user's session/token
func (s *UserService) Logout(ctx context.Context, userID string, token string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return model.ErrUserNotFound
	}

	// For Phase 4, we'll implement basic logout
	// In a full implementation, this would invalidate sessions/tokens
	_ = user
	_ = token

	return nil
}

// RefreshToken generates a new access token for an authenticated user
func (s *UserService) RefreshToken(ctx context.Context, currentToken string) (*model.IDPAccessToken, error) {
	// Validate current token first
	user, err := s.ValidateToken(ctx, currentToken)
	if err != nil {
		return nil, err
	}

	// Generate new access token with 24 hour expiry
	// In a real implementation, you'd use proper token generation with crypto/rand
	newToken, err := model.NewIDPAccessToken(
		model.AuthMethodAPIKey, // This would be the appropriate auth method
		"terrareg", // provider name
		generateTokenString(),
		"Bearer",
		"", // client ID - would be populated based on auth method
		[]string{"read", "write"}, // default scopes
		user.ID(),
		24*time.Hour, // TTL
	)
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

// Helper function to generate a random token string
func generateTokenString() string {
	// Simple implementation - in production, use crypto/rand or similar
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}