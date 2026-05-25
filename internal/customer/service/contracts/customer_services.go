package service

import (
	"istore/internal/customer/domain"
	saleDomain "istore/internal/sale/domain"
	"istore/pkg/rest_err"
	"time"
)

type CreateCustomerInput struct {
	UserID uint
	Name   string
	Phone  string
}

type UpdateCustomerInput struct {
	UserID uint
	Name   string
	Phone  string
}

type ListCustomersInput struct {
	UserID      uint
	Page        int
	Limit       int
	Start       *time.Time
	End         *time.Time
	Status      *saleDomain.PaymentStatus
	PaymentType *saleDomain.PaymentType
	Search      string
}

type CustomerService interface {
	Create(input CreateCustomerInput) (*domain.Customer, *rest_err.RestErr)
	Update(id int, input UpdateCustomerInput) (*domain.Customer, *rest_err.RestErr)
	Delete(userID uint, id int) *rest_err.RestErr
	DeleteMany(userID uint, ids []int) (int, *rest_err.RestErr)

	GetByID(userID uint, id int) (*domain.Customer, *rest_err.RestErr)

	List(input ListCustomersInput) (*domain.CustomerListResult, *rest_err.RestErr)
}
