package entity

import (
	saleDomain "istore/internal/sale/domain"
)

type SaleItemEntity struct {
	ID     int `gorm:"primaryKey"`
	SaleID int `gorm:"column:sale_id;not null;index"`

	ProductName string `gorm:"column:product_name;not null"`
	Specs       string `gorm:"column:specs"`

	Quantity  int `gorm:"column:quantity;not null"`
	CostPrice int `gorm:"column:cost_price;not null"`
	SalePrice int `gorm:"column:sale_price;not null"`
}

func (SaleItemEntity) TableName() string {
	return "sale_items"
}

func FromSaleItemDomain(i *saleDomain.SaleItem) *SaleItemEntity {
	if i == nil {
		return nil
	}

	return &SaleItemEntity{
		ID:          i.ID,
		SaleID:      i.SaleID,
		ProductName: i.ProductName,
		Specs:       i.Specs,
		Quantity:    i.Quantity,
		CostPrice:   i.CostPrice,
		SalePrice:   i.SalePrice,
	}
}

func (si *SaleItemEntity) ToDomain() *saleDomain.SaleItem {
	if si == nil {
		return nil
	}

	return &saleDomain.SaleItem{
		ID:          si.ID,
		SaleID:      si.SaleID,
		ProductName: si.ProductName,
		Specs:       si.Specs,
		Quantity:    si.Quantity,
		CostPrice:   si.CostPrice,
		SalePrice:   si.SalePrice,
	}
}
