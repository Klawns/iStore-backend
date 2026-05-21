package implementation

import (
	"errors"
	"istore/internal/sale/domain"
	"istore/internal/sale/repository/contracts"
	"istore/internal/sale/repository/entity"
	"time"

	"gorm.io/gorm"
)

type saleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) contracts.SaleRepository {
	return &saleRepository{db: db}
}

func (r *saleRepository) Create(sale *domain.Sale) error {
	saleEntity := entity.FromSaleDomain(sale)
	if err := r.db.Create(saleEntity).Error; err != nil {
		return err
	}

	sale.ID = saleEntity.ID
	for i := range sale.Items {
		if i < len(saleEntity.Items) {
			sale.Items[i].ID = saleEntity.Items[i].ID
			sale.Items[i].SaleID = saleEntity.ID
		}
	}

	return nil
}

func (r *saleRepository) FindByID(id int) (*domain.Sale, error) {
	var saleEntity entity.SaleEntity
	if err := r.db.Preload("Items").First(&saleEntity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return saleEntity.ToDomain(), nil
}

func (r *saleRepository) FindAll() ([]domain.Sale, error) {
	var saleEntities []entity.SaleEntity
	if err := r.db.Preload("Items").Find(&saleEntities).Error; err != nil {
		return nil, err
	}

	sales := make([]domain.Sale, len(saleEntities))
	for i, saleEntity := range saleEntities {
		sales[i] = *saleEntity.ToDomain()
	}

	return sales, nil
}

func (r *saleRepository) ListByPeriod(start time.Time, end time.Time) ([]domain.Sale, error) {
	var saleEntities []entity.SaleEntity
	if err := r.db.Preload("Items").
		Where("sale_date BETWEEN ? AND ?", start, end).
		Find(&saleEntities).Error; err != nil {
		return nil, err
	}

	sales := make([]domain.Sale, len(saleEntities))
	for i, saleEntity := range saleEntities {
		sales[i] = *saleEntity.ToDomain()
	}

	return sales, nil
}

func (r *saleRepository) UpdateStatus(id int, status domain.PaymentStatus) error {
	return r.db.Model(&entity.SaleEntity{}).
		Where("id = ?", id).
		Update("payment_status", status).Error
}

func (r *saleRepository) Delete(id int) error {
	return r.db.Delete(&entity.SaleEntity{}, id).Error
}
