package implementation

import (
	"errors"
	"istore/internal/auth/domain"
	"istore/internal/auth/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type jwtService struct {
	secret string
}

func NewJwtService(secret string) contracts.JwtProvider {
	return &jwtService{
		secret: secret,
	}
}

func (j *jwtService) Generate(userID uint) (string, *rest_err.RestErr) {
	logger.Info("generating JWT for user", zap.Uint("user_id", userID), zap.String("journey", "GenerateToken"))

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		logger.Error("error generating JWT", err, zap.String("journey", "GenerateToken"))
		return "", rest_err.NewInternalServerError("error generating JWT")
	}
	return tokenString, nil
}

func (j *jwtService) Validate(tokenString string) (*domain.TokenPayload, *rest_err.RestErr) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {

			_, ok := token.Method.(*jwt.SigningMethodHMAC)

			if !ok {
				return nil, errors.New(
					"invalid signing method",
				)
			}

			return []byte(j.secret), nil
		},
	)
	if err != nil {
		logger.Error("error validating JWT", err, zap.String("journey", "ValidateToken"))
		return nil, rest_err.NewUnauthorizedRequestError("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		logger.Error("invalid JWT claims", err, zap.String("journey", "ValidateToken"))
		return nil, rest_err.NewUnauthorizedRequestError("invalid token")
	}

	userID, ok := readUintClaim(claims["user_id"])
	if !ok {
		return nil, rest_err.NewUnauthorizedRequestError("invalid token")
	}

	exp, ok := readInt64Claim(claims["exp"])
	if !ok {
		return nil, rest_err.NewUnauthorizedRequestError("invalid token")
	}

	payload := &domain.TokenPayload{
		UserID: userID,
		Exp:    exp,
	}

	return payload, nil
}

func readUintClaim(value interface{}) (uint, bool) {
	switch v := value.(type) {
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(parsed), true
	default:
		return 0, false
	}
}

func readInt64Claim(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
