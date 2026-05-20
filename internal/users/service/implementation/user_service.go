package implementation

import (
	"istore/internal/users/domain"
	"istore/internal/users/dto/request"
	"istore/internal/users/dto/response"
	repository "istore/internal/users/repository/contracts"
	"istore/internal/users/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) contracts.UserService {
	return &userService{repository: repository}
}

func (s *userService) Create(req request.UserRequest) (*response.UserResponse, *rest_err.RestErr) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	existingUser, err := s.repository.FindByEmail(email)
	if err != nil {
		logger.Error("error finding user by email", err, zap.String("journey", "CreateUser"))
		return nil, rest_err.NewInternalServerError("error creating user")
	}
	if existingUser != nil {
		return nil, rest_err.NewBadRequestError("email already registered")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("error hashing password", err, zap.String("journey", "CreateUser"))
		return nil, rest_err.NewInternalServerError("error creating user")
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
	}

	if err := s.repository.Create(user); err != nil {
		logger.Error("error creating user", err, zap.String("journey", "CreateUser"))
		return nil, rest_err.NewInternalServerError("error creating user")
	}

	return response.FromDomain(user), nil
}

func (s *userService) FindByID(id uint) (*response.UserResponse, *rest_err.RestErr) {
	user, err := s.repository.FindByID(id)
	if err != nil {
		logger.Error("error finding user by id", err, zap.Uint("user_id", id), zap.String("journey", "FindUserByID"))
		return nil, rest_err.NewInternalServerError("error finding user")
	}
	if user == nil {
		return nil, rest_err.NewNotFoundError("user not found")
	}

	return response.FromDomain(user), nil
}
