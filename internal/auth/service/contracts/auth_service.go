package contracts

import "istore/pkg/rest_err"

type SignInInput struct {
	Email    string
	Password string
}

type AuthService interface {
	SignIn(input SignInInput) (string, *rest_err.RestErr)
}
