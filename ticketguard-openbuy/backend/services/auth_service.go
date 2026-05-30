package services

import (
	"fmt"
	"strings"

	"ticketguard/backend/config"
	"ticketguard/backend/dao"
	"ticketguard/backend/domain"
	"ticketguard/backend/utils"
)

type AuthService struct {
	users *dao.UserDAO
	cfg   config.Config
}

type RegisterInput struct {
	Name     string          `json:"name" binding:"required"`
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=6"`
	Role     domain.UserRole `json:"role"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func NewAuthService(users *dao.UserDAO, cfg config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

func (s *AuthService) Register(input RegisterInput) (*AuthResponse, error) {
	role := input.Role
	if role == "" {
		role = domain.RoleCliente
	}
	if !domain.IsValidRole(role) {
		return nil, fmt.Errorf("%w: invalid role", utils.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Password) == "" {
		return nil, fmt.Errorf("%w: missing fields", utils.ErrInvalidInput)
	}
	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Name:         strings.TrimSpace(input.Name),
		Email:        domain.NormalizeEmail(input.Email),
		PasswordHash: hash,
		Role:         role,
		Balance:      0,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return s.issueToken(user)
}

func (s *AuthService) Login(input LoginInput) (*AuthResponse, error) {
	user, err := s.users.FindByEmail(input.Email)
	if err != nil {
		return nil, utils.ErrInvalidLogin
	}
	if !utils.CheckPassword(input.Password, user.PasswordHash) {
		return nil, utils.ErrInvalidLogin
	}
	return s.issueToken(user)
}

func (s *AuthService) issueToken(user *domain.User) (*AuthResponse, error) {
	token, err := utils.GenerateJWT(user.ID, user.Role, s.cfg.JWTSecret, s.cfg.JWTDuration())
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: *user}, nil
}
