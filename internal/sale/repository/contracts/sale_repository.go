package contracts

import (
	"istore/internal/sale/domain"
	"time"
)

type SaleRepository interface {
	Create(sale *domain.Sale) error

	FindByID(userID uint, id int) (*domain.Sale, error)

	FindAll() ([]domain.Sale, error)

	List(filter domain.SaleListFilter) (*domain.SaleListResult, error)

	ListByPeriod(
		userID uint,
		start time.Time,
		end time.Time,
	) ([]domain.Sale, error)

	UpdateStatus(
		userID uint,
		id int,
		status domain.PaymentStatus,
	) error

	Delete(userID uint, id int) error

	ListInstallmentAlerts(userID uint, now time.Time, windowDays int) ([]domain.SaleInstallment, error)

	ListInstallmentsBySaleID(userID uint, saleID int) ([]domain.SaleInstallment, error)

	UpdateInstallmentStatus(userID uint, id int, status domain.SaleInstallmentStatus, notes string, validatedAt time.Time) (*domain.SaleInstallment, error)
}
