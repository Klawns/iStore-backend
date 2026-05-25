package contracts

import "istore/internal/customer/domain"

type CustomerRepository interface {
	Create(customer *domain.Customer) error
	Update(customer *domain.Customer) error
	Delete(userID uint, id int) error
	DeleteMany(userID uint, ids []int) error
	FindByID(userID uint, id int) (*domain.Customer, error)
	FindByIDs(userID uint, ids []int) ([]domain.Customer, error)
	FindAll() ([]domain.Customer, error)
	List(filter domain.CustomerListFilter) (*domain.CustomerListResult, error)
	CountSalesByCustomerIDs(userID uint, ids []int) (int64, error)
}
