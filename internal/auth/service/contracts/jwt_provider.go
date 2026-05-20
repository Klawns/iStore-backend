package contracts

import (
	"istore/internal/auth/domain"
	"istore/pkg/rest_err"
)

type JwtProvider interface {
	Generate(userID uint, email string) (string, *rest_err.RestErr)
	Validate(token string) (*domain.TokenPayload, *rest_err.RestErr)
}
