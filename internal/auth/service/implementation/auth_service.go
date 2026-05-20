package implementation

import (
	"istore/internal/auth/dto/request"
	authContracts "istore/internal/auth/service/contracts"
	userRepository "istore/internal/users/repository/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepository userRepository.UserRepository
	jwtProvider    authContracts.JwtProvider
}

func NewAuthService(userRepository userRepository.UserRepository, jwtProvider authContracts.JwtProvider) authContracts.AuthService {
	return &authService{
		userRepository: userRepository,
		jwtProvider:    jwtProvider,
	}
}

func (s *authService) SignIn(req request.AuthRequest) (string, *rest_err.RestErr) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		logger.Error("error finding user by email", err, zap.String("journey", "SignIn"))
		return "", rest_err.NewInternalServerError("error signing in")
	}
	if user == nil {
		return "", rest_err.NewUnauthorizedRequestError("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", rest_err.NewUnauthorizedRequestError("invalid credentials")
	}

	token, restErr := s.jwtProvider.Generate(user.ID, user.Email)
	if restErr != nil {
		return "", restErr
	}

	return token, nil
}
