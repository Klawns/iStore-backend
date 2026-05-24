package entity

import (
	saleDomain "istore/internal/sale/domain"
	"time"
)

type SaleInstallmentEntity struct {
	ID                int                              `gorm:"primaryKey"`
	SaleID            int                              `gorm:"column:sale_id;not null;uniqueIndex:idx_sale_installment_unique"`
	Sale              SaleEntity                       `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE"`
	DueDate           time.Time                        `gorm:"column:due_date;not null;index"`
	InstallmentNumber int                              `gorm:"column:installment_number;not null;uniqueIndex:idx_sale_installment_unique"`
	TotalInstallments int                              `gorm:"column:total_installments;not null"`
	Amount            int                              `gorm:"column:amount;not null"`
	Status            saleDomain.SaleInstallmentStatus `gorm:"column:status;not null;index"`
	PaidAt            *time.Time                       `gorm:"column:paid_at"`
	ValidatedAt       *time.Time                       `gorm:"column:validated_at"`
	Notes             string                           `gorm:"column:notes"`
	CreatedAt         time.Time                        `gorm:"column:created_at"`
	UpdatedAt         time.Time                        `gorm:"column:updated_at"`
}

func (SaleInstallmentEntity) TableName() string {
	return "sale_installments"
}

func FromSaleInstallmentDomain(installment *saleDomain.SaleInstallment) *SaleInstallmentEntity {
	if installment == nil {
		return nil
	}

	return &SaleInstallmentEntity{
		ID:                installment.ID,
		SaleID:            installment.SaleID,
		DueDate:           installment.DueDate,
		InstallmentNumber: installment.InstallmentNumber,
		TotalInstallments: installment.TotalInstallments,
		Amount:            installment.Amount,
		Status:            installment.Status,
		PaidAt:            installment.PaidAt,
		ValidatedAt:       installment.ValidatedAt,
		Notes:             installment.Notes,
		CreatedAt:         installment.CreatedAt,
		UpdatedAt:         installment.UpdatedAt,
	}
}

func (s *SaleInstallmentEntity) ToDomain() *saleDomain.SaleInstallment {
	if s == nil {
		return nil
	}

	return &saleDomain.SaleInstallment{
		ID:                s.ID,
		SaleID:            s.SaleID,
		CustomerName:      s.Sale.Customer.Name,
		DueDate:           s.DueDate,
		InstallmentNumber: s.InstallmentNumber,
		TotalInstallments: s.TotalInstallments,
		Amount:            s.Amount,
		Status:            s.Status,
		PaidAt:            s.PaidAt,
		ValidatedAt:       s.ValidatedAt,
		Notes:             s.Notes,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}
