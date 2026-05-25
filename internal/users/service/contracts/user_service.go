package contracts

import (
	"istore/internal/users/domain"
	"istore/pkg/rest_err"
)

type CreateUserInput struct {
	Email                string
	Password             string
	AcceptPrivacyPolicy  bool
	AcceptTerms          bool
	PrivacyPolicyVersion string
	TermsVersion         string
}

type UserService interface {
	Create(input CreateUserInput) (*domain.User, *rest_err.RestErr)
	FindByID(id uint) (*domain.User, *rest_err.RestErr)
	DeleteOwnAccount(userID uint, password string) *rest_err.RestErr
}
