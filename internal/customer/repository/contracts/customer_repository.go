package contracts

import "istore/internal/customer/domain"

type CustomerRepository interface {
	Create(customer *domain.Customer) error
	Update(customer *domain.Customer) error
	Delete(id int) error
	FindByID(id int) (*domain.Customer, error)
	FindAll() ([]domain.Customer, error)
}
