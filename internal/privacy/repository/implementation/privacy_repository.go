package implementation

import (
	"istore/internal/privacy/domain"
	"istore/internal/privacy/repository/contracts"
	privacyEntity "istore/internal/privacy/repository/entity"

	"gorm.io/gorm"
)

type privacyRepository struct {
	db *gorm.DB
}

func NewPrivacyRepository(db *gorm.DB) contracts.PrivacyRepository {
	return &privacyRepository{db: db}
}

func (r *privacyRepository) Create(request *domain.PrivacyRequest) error {
	requestEntity := privacyEntity.FromDomain(request)
	if err := r.db.Create(requestEntity).Error; err != nil {
		return err
	}

	*request = *requestEntity.ToDomain()
	return nil
}

func (r *privacyRepository) ListByUserID(userID uint) ([]domain.PrivacyRequest, error) {
	var requestEntities []privacyEntity.PrivacyRequestEntity
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC, id DESC").Find(&requestEntities).Error; err != nil {
		return nil, err
	}

	requests := make([]domain.PrivacyRequest, len(requestEntities))
	for i := range requestEntities {
		requests[i] = *requestEntities[i].ToDomain()
	}

	return requests, nil
}
