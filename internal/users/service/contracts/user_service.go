package contracts

import (
	"istore/internal/users/domain"
	"istore/pkg/rest_err"
)

type CreateUserInput struct {
	Email    string
	Password string
}

type UserService interface {
	Create(input CreateUserInput) (*domain.User, *rest_err.RestErr)
	FindByID(id uint) (*domain.User, *rest_err.RestErr)
}
