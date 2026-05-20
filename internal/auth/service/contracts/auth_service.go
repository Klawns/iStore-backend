package contracts

import (
	"istore/internal/auth/dto/request"
	"istore/pkg/rest_err"
)

type AuthService interface {
	SignIn(req request.AuthRequest) (string, *rest_err.RestErr)
}
