package contract

import (
	"istore/internal/sale/domain"
	"istore/pkg/rest_err"
	"time"
)

type CreateSaleInput struct {
	UserID          uint
	ClienteID       int64
	TipoPagamento   domain.PaymentType
	StatusPagamento domain.PaymentStatus
	SaleDate        time.Time
	Installments    *int
	BillingDay      *int
	Itens           []CreateSaleItemInput
}

type CreateSaleItemInput struct {
	ProductName string
	Specs       string
	Quantity    int
	CostPrice   int // centavos
	SalePrice   int // centavos
}

type UpdateInstallmentStatusInput struct {
	Status domain.SaleInstallmentStatus
	Notes  string
}

type ListSalesInput struct {
	UserID      uint
	Page        int
	Limit       int
	Start       *time.Time
	End         *time.Time
	Status      *domain.PaymentStatus
	PaymentType *domain.PaymentType
	CustomerID  *int
	Search      string
}

type SaleService interface {
	Create(input *CreateSaleInput) (*domain.Sale, *rest_err.RestErr)

	GetByID(userID uint, id int) (*domain.Sale, *rest_err.RestErr)

	List(input ListSalesInput) (*domain.SaleListResult, *rest_err.RestErr)

	ListByPeriod(
		userID uint,
		start time.Time,
		end time.Time,
	) ([]domain.Sale, *rest_err.RestErr)

	UpdateStatus(
		userID uint,
		id int,
		status domain.PaymentStatus,
	) *rest_err.RestErr

	Delete(userID uint, id int) *rest_err.RestErr

	ListInstallmentAlerts(userID uint, now time.Time, windowDays int) ([]domain.SaleInstallment, *rest_err.RestErr)

	ListInstallmentsBySaleID(userID uint, saleID int) ([]domain.SaleInstallment, *rest_err.RestErr)

	UpdateInstallmentStatus(userID uint, id int, input UpdateInstallmentStatusInput) (*domain.SaleInstallment, *rest_err.RestErr)
}
