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

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func NewAuthService(users *dao.UserDAO, cfg config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

func (s *AuthService) Register(input RegisterInput) (*AuthResponse, error) {
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
		Role:         domain.RoleUsuario,
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
	user.Role = domain.NormalizeRole(user.Role)
	return s.issueToken(user)
}

func (s *AuthService) ChangePassword(userID uint, input ChangePasswordInput) error {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	if !utils.CheckPassword(input.CurrentPassword, user.PasswordHash) {
		return utils.ErrInvalidLogin
	}
	if strings.TrimSpace(input.NewPassword) == "" || len(input.NewPassword) < 6 {
		return fmt.Errorf("%w: new password must have at least 6 characters", utils.ErrInvalidInput)
	}
	hash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.users.Update(user)
}

func (s *AuthService) issueToken(user *domain.User) (*AuthResponse, error) {
	user.Role = domain.NormalizeRole(user.Role)
	token, err := utils.GenerateJWT(user.ID, user.Role, s.cfg.JWTSecret, s.cfg.JWTDuration())
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: *user}, nil
}
