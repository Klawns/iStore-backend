package contracts

import (
	"istore/internal/sale/domain"
	"time"
)

type SaleRepository interface {
	Create(sale *domain.Sale) error

	FindByID(id int) (*domain.Sale, error)

	FindAll() ([]domain.Sale, error)

	List(filter domain.SaleListFilter) (*domain.SaleListResult, error)

	ListByPeriod(
		start time.Time,
		end time.Time,
	) ([]domain.Sale, error)

	UpdateStatus(
		id int,
		status domain.PaymentStatus,
	) error

	Delete(id int) error

	ListInstallmentAlerts(now time.Time, windowDays int) ([]domain.SaleInstallment, error)

	ListInstallmentsBySaleID(saleID int) ([]domain.SaleInstallment, error)

	UpdateInstallmentStatus(id int, status domain.SaleInstallmentStatus, notes string, validatedAt time.Time) (*domain.SaleInstallment, error)
}
