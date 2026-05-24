package entity

import (
	customerEntity "istore/internal/customer/repository/entity"
	saleDomain "istore/internal/sale/domain"
	"time"
)

type SaleEntity struct {
	ID            int                           `gorm:"primaryKey"`
	CustomerID    int                           `gorm:"column:customer_id;not null;index"`
	Customer      customerEntity.CustomerEntity `gorm:"foreignKey:CustomerID"`
	TotalValue    int                           `gorm:"column:total_value;not null"`
	PaymentStatus saleDomain.PaymentStatus      `gorm:"column:payment_status;not null"`
	PaymentType   saleDomain.PaymentType        `gorm:"column:payment_type;not null"`
	SaleDate      time.Time                     `gorm:"column:sale_date;not null"`
	Installments  *int                          `gorm:"column:installments"`
	BillingDay    *int                          `gorm:"column:billing_day"`

	Items            []SaleItemEntity        `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE"`
	InstallmentsList []SaleInstallmentEntity `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE"`
}

func (SaleEntity) TableName() string {
	return "sales"
}

func FromSaleDomain(sale *saleDomain.Sale) *SaleEntity {
	if sale == nil {
		return nil
	}

	items := make([]SaleItemEntity, len(sale.Items))
	for i := range sale.Items {
		item := sale.Items[i]
		itemEntity := FromSaleItemDomain(&item)
		if itemEntity != nil {
			items[i] = *itemEntity
		}
	}

	return &SaleEntity{
		ID:            sale.ID,
		CustomerID:    sale.CustomerID,
		TotalValue:    sale.TotalValue,
		PaymentStatus: sale.PaymentStatus,
		PaymentType:   sale.PaymentType,
		SaleDate:      sale.SaleDate,
		Installments:  sale.Installments,
		BillingDay:    sale.BillingDay,
		Items:         items,
	}
}

func (s *SaleEntity) ToDomain() *saleDomain.Sale {
	if s == nil {
		return nil
	}

	items := make([]saleDomain.SaleItem, len(s.Items))
	for i := range s.Items {
		items[i] = *s.Items[i].ToDomain()
	}

	return &saleDomain.Sale{
		ID:            s.ID,
		CustomerID:    s.CustomerID,
		CustomerName:  s.Customer.Name,
		TotalValue:    s.TotalValue,
		PaymentStatus: s.PaymentStatus,
		PaymentType:   s.PaymentType,
		SaleDate:      s.SaleDate,
		Installments:  s.Installments,
		BillingDay:    s.BillingDay,
		Items:         items,
	}
}
