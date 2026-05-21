package contract

import (
	"istore/internal/sale/domain"
	"istore/pkg/rest_err"
	"time"
)

type CreateSaleInput struct {
	ClienteID       int64
	TipoPagamento   domain.PaymentType
	StatusPagamento domain.PaymentStatus
	SaleDate        time.Time
	Itens           []CreateSaleItemInput
}

type CreateSaleItemInput struct {
	ProductName string
	Specs       string
	Quantity    int
	CostPrice   int // centavos
	SalePrice   int // centavos
}

type SaleService interface {
	Create(input *CreateSaleInput) (*domain.Sale, *rest_err.RestErr)

	GetByID(id int) (*domain.Sale, *rest_err.RestErr)

	List() ([]domain.Sale, *rest_err.RestErr)

	ListByPeriod(
		start time.Time,
		end time.Time,
	) ([]domain.Sale, *rest_err.RestErr)

	UpdateStatus(
		id int,
		status domain.PaymentStatus,
	) *rest_err.RestErr

	Delete(id int) *rest_err.RestErr
}
