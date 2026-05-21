package contracts

import (
	"istore/internal/sale/domain"
	"time"
)

type SaleRepository interface {
	Create(sale *domain.Sale) error

	FindByID(id int) (*domain.Sale, error)

	FindAll() ([]domain.Sale, error)

	ListByPeriod(
		start time.Time,
		end time.Time,
	) ([]domain.Sale, error)

	UpdateStatus(
		id int,
		status domain.PaymentStatus,
	) error

	Delete(id int) error
}
