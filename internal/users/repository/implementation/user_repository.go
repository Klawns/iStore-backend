package implementation

import (
	"errors"
	customerEntity "istore/internal/customer/repository/entity"
	privacyEntity "istore/internal/privacy/repository/entity"
	saleEntity "istore/internal/sale/repository/entity"
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

func (r *userRepository) DeleteOwnAccount(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		userSales := tx.Model(&saleEntity.SaleEntity{}).Select("id").Where("user_id = ?", id)

		if err := tx.Where("sale_id IN (?)", userSales).Delete(&saleEntity.PaymentAlertEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("sale_id IN (?)", userSales).Delete(&saleEntity.SaleInstallmentEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("sale_id IN (?)", userSales).Delete(&saleEntity.SaleItemEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&saleEntity.SaleEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&customerEntity.CustomerEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&privacyEntity.PrivacyRequestEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&entity.UserEntity{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}
