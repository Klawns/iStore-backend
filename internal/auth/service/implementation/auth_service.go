package implementation

import (
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

func (s *authService) SignIn(input authContracts.SignInInput) (string, *rest_err.RestErr) {
	email := strings.TrimSpace(strings.ToLower(input.Email))

	logger.Info("sign in user lookup starting", zap.String("journey", "SignIn"), zap.String("stage", "find_user_by_email"))
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		logger.Error("error finding user by email", err, zap.String("journey", "SignIn"), zap.String("stage", "find_user_by_email"))
		return "", rest_err.NewInternalServerError("error signing in")
	}
	if user == nil {
		logger.Info("sign in invalid credentials", zap.String("journey", "SignIn"), zap.String("stage", "user_not_found"))
		return "", rest_err.NewUnauthorizedRequestError("invalid credentials")
	}

	logger.Info("sign in password check starting", zap.String("journey", "SignIn"), zap.String("stage", "compare_password"), zap.Uint("user_id", user.ID))
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		logger.Info("sign in invalid credentials", zap.String("journey", "SignIn"), zap.String("stage", "password_mismatch"), zap.Uint("user_id", user.ID))
		return "", rest_err.NewUnauthorizedRequestError("invalid credentials")
	}

	logger.Info("sign in token generation starting", zap.String("journey", "SignIn"), zap.String("stage", "generate_token"), zap.Uint("user_id", user.ID))
	token, restErr := s.jwtProvider.Generate(user.ID)
	if restErr != nil {
		return "", restErr
	}

	return token, nil
}
