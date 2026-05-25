package contracts

import "istore/internal/users/domain"

type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uint) (*domain.User, error)
	DeleteOwnAccount(id uint) error
}
