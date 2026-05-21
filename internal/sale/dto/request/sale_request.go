package request

import (
	saleDomain "istore/internal/sale/domain"
	"time"
)

type SaleRequest struct {
	ClienteID       int64                    `json:"customerId" validate:"required,gt=0"`
	TipoPagamento   saleDomain.PaymentType   `json:"paymentType" validate:"required"`
	StatusPagamento saleDomain.PaymentStatus `json:"paymentStatus" validate:"required"`
	SaleDate        time.Time                `json:"saleDate"`
	Itens           []SaleItemRequest        `json:"items" validate:"required,min=1,dive"`
}

type SaleItemRequest struct {
	ProductName string `json:"productName" validate:"required"`
	Specs       string `json:"specs"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	CostPrice   int    `json:"costPrice" validate:"required,gte=0"`
	SalePrice   int    `json:"salePrice" validate:"required,gte=0"`
}

type SaleStatusRequest struct {
	Status saleDomain.PaymentStatus `json:"status" validate:"required"`
}
