package entity

import (
	saleDomain "istore/internal/sale/domain"
	"time"
)

type PaymentAlertEntity struct {
	ID                int                           `gorm:"primaryKey"`
	SaleID            int                           `gorm:"column:sale_id;not null;uniqueIndex:idx_payment_alert_unique"`
	Sale              SaleEntity                    `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE"`
	DueDate           time.Time                     `gorm:"column:due_date;not null;uniqueIndex:idx_payment_alert_unique"`
	InstallmentNumber int                           `gorm:"column:installment_number;not null;uniqueIndex:idx_payment_alert_unique"`
	Message           string                        `gorm:"column:message;not null"`
	Status            saleDomain.PaymentAlertStatus `gorm:"column:status;not null;index"`
	CreatedAt         time.Time                     `gorm:"column:created_at"`
	UpdatedAt         time.Time                     `gorm:"column:updated_at"`
}

func (PaymentAlertEntity) TableName() string {
	return "payment_alerts"
}

func FromPaymentAlertDomain(alert *saleDomain.PaymentAlert) *PaymentAlertEntity {
	if alert == nil {
		return nil
	}

	return &PaymentAlertEntity{
		ID:                alert.ID,
		SaleID:            alert.SaleID,
		DueDate:           alert.DueDate,
		InstallmentNumber: alert.InstallmentNumber,
		Message:           alert.Message,
		Status:            alert.Status,
		CreatedAt:         alert.CreatedAt,
		UpdatedAt:         alert.UpdatedAt,
	}
}

func (p *PaymentAlertEntity) ToDomain() *saleDomain.PaymentAlert {
	if p == nil {
		return nil
	}

	return &saleDomain.PaymentAlert{
		ID:                p.ID,
		SaleID:            p.SaleID,
		DueDate:           p.DueDate,
		InstallmentNumber: p.InstallmentNumber,
		Message:           p.Message,
		Status:            p.Status,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}
