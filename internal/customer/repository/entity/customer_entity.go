package entity

import (
	"istore/internal/customer/domain"
	"time"
)

type CustomerEntity struct {
	ID        int    `gorm:"primaryKey"`
	UserID    uint   `gorm:"column:user_id;index"`
	Name      string `gorm:"not null"`
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (CustomerEntity) TableName() string {
	return "customers"
}

// Transforma um domínio Customer em uma entidade CustomerEntity
func FromDomain(customer *domain.Customer) *CustomerEntity {
	if customer == nil {
		return nil
	}
	return &CustomerEntity{
		ID:     customer.ID,
		UserID: customer.UserID,
		Name:   customer.Name,
		Phone:  customer.Phone,
	}
}

// Transforma uma entidade CustomerEntity em um domínio Customer
func (u *CustomerEntity) ToDomain() *domain.Customer {
	if u == nil {
		return nil
	}

	return &domain.Customer{
		ID:     u.ID,
		UserID: u.UserID,
		Name:   u.Name,
		Phone:  u.Phone,
	}
}
