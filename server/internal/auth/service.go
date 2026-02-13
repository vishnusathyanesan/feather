package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/feather-chat/feather/internal/model"
)

var (
	ErrEmailTaken      = errors.New("email already registered")
	ErrInvalidCreds    = errors.New("invalid email or password")
	ErrInvalidToken    = errors.New("invalid or expired refresh token")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserDeactivated = errors.New("user account is deactivated")
)

type Service struct {
	repo   *Repository
	tokens *TokenService
}

func NewService(repo *Repository, tokens *TokenService) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	existing, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hash),
		Role:         model.RoleMember,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *Service) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCreds
	}
	if !user.IsActive {
		return nil, ErrUserDeactivated
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *Service) GoogleLogin(ctx context.Context, googleID, email, name string) (*model.AuthResponse, error) {
	// Branch 1: user found by google_id → login
	user, err := s.repo.GetUserByGoogleID(ctx, googleID)
	if err != nil {
		return nil, fmt.Errorf("find user by google id: %w", err)
	}
	if user != nil {
		if !user.IsActive {
			return nil, ErrUserDeactivated
		}
		return s.generateAuthResponse(ctx, user)
	}

	// Branch 2: user found by email (no google_id) → link Google account
	user, err = s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	if user != nil {
		if !user.IsActive {
			return nil, ErrUserDeactivated
		}
		if err := s.repo.LinkGoogleID(ctx, user.ID, googleID); err != nil {
			return nil, fmt.Errorf("link google id: %w", err)
		}
		user.GoogleID = &googleID
		return s.generateAuthResponse(ctx, user)
	}

	// Branch 3: no user → create new user with Google info
	now := time.Now()
	user = &model.User{
		ID:           uuid.New(),
		Email:        email,
		Name:         name,
		PasswordHash: "",
		GoogleID:     &googleID,
		Role:         model.RoleMember,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create google user: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (*model.AuthResponse, error) {
	tokenHash := HashToken(rawToken)

	rec, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if rec == nil || rec.RevokedAt != nil || rec.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	// Revoke old token
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoke old token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, rec.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, ErrUserNotFound
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	tokenHash := HashToken(rawToken)
	return s.repo.RevokeRefreshToken(ctx, tokenHash)
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *Service) generateAuthResponse(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	accessToken, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	rawRefresh, refreshHash, expiresAt, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.repo.StoreRefreshToken(ctx, user.ID, refreshHash, expiresAt); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &model.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}
