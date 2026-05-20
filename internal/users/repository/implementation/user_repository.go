package implementation

import (
	"errors"
	"istore/internal/users/domain"
	"istore/internal/users/repository/contracts"
	"istore/internal/users/repository/entity"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) contracts.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	userEntity := entity.FromDomain(user)
	if err := r.db.Create(userEntity).Error; err != nil {
		return err
	}

	user.ID = userEntity.ID
	return nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var userEntity entity.UserEntity
	if err := r.db.Where("email = ?", email).First(&userEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return userEntity.ToDomain(), nil
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var userEntity entity.UserEntity
	if err := r.db.First(&userEntity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return userEntity.ToDomain(), nil
}
