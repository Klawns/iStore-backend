package implementation

import (
	"istore/internal/users/domain"
	repository "istore/internal/users/repository/contracts"
	"istore/internal/users/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	currentPrivacyPolicyVersion = "2026-05-25"
	currentTermsVersion         = "2026-05-25"
)

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) contracts.UserService {
	return &userService{repository: repository}
}

func (s *userService) Create(input contracts.CreateUserInput) (*domain.User, *rest_err.RestErr) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if !input.AcceptPrivacyPolicy || !input.AcceptTerms {
		return nil, rest_err.NewBadRequestError("privacy policy and terms acceptance is required")
	}

	existingUser, err := s.repository.FindByEmail(email)
	if err != nil {
		logger.Error("error finding user by email", err, zap.String("journey", "CreateUser"))
		return nil, rest_err.NewInternalServerError("error creating user")
	}
	if existingUser != nil {
		return nil, rest_err.NewBadRequestError("Nao foi possivel criar conta com os dados informados")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("error hashing password", err, zap.String("journey", "CreateUser"))
		return nil, rest_err.NewInternalServerError("error creating user")
	}

	acceptedAt := time.Now().UTC()
	user := &domain.User{
		Email:                email,
		PasswordHash:         string(passwordHash),
		PrivacyPolicyVersion: currentPrivacyPolicyVersion,
		PrivacyAcceptedAt:    ptrTime(acceptedAt),
		TermsVersion:         currentTermsVersion,
		TermsAcceptedAt:      ptrTime(acceptedAt),
	}

	if err := s.repository.Create(user); err != nil {
		logger.Error("error creating user", err, zap.String("journey", "CreateUser"))
		return nil, rest_err.NewInternalServerError("error creating user")
	}

	return user, nil
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func (s *userService) FindByID(id uint) (*domain.User, *rest_err.RestErr) {
	user, err := s.repository.FindByID(id)
	if err != nil {
		logger.Error("error finding user by id", err, zap.Uint("user_id", id), zap.String("journey", "FindUserByID"))
		return nil, rest_err.NewInternalServerError("error finding user")
	}
	if user == nil {
		return nil, rest_err.NewNotFoundError("user not found")
	}

	return user, nil
}

func (s *userService) DeleteOwnAccount(userID uint, password string) *rest_err.RestErr {
	if userID == 0 {
		return rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	user, err := s.repository.FindByID(userID)
	if err != nil {
		logger.Error("error finding user by id", err, zap.Uint("user_id", userID), zap.String("journey", "DeleteOwnAccount"))
		return rest_err.NewInternalServerError("error deleting user")
	}
	if user == nil {
		return rest_err.NewNotFoundError("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return rest_err.NewUnauthorizedRequestError("Senha invalida")
	}

	if err := s.repository.DeleteOwnAccount(userID); err != nil {
		logger.Error("error deleting user account", err, zap.Uint("user_id", userID), zap.String("journey", "DeleteOwnAccount"))
		return rest_err.NewInternalServerError("error deleting user")
	}

	return nil
}
