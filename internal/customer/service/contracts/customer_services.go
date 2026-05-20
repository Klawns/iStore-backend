package service

import (
	"istore/internal/customer/domain"
	"istore/pkg/rest_err"
)

type CreateCustomerInput struct {
	Name  string
	Phone string
}

type UpdateCustomerInput struct {
	Name  string
	Phone string
}

type CustomerService interface {
	Create(input CreateCustomerInput) *rest_err.RestErr
	Update(id int, input UpdateCustomerInput) *rest_err.RestErr
	Delete(id int) *rest_err.RestErr

	GetByID(id int) (*domain.Customer, *rest_err.RestErr)

	List() ([]domain.Customer, *rest_err.RestErr)
}
