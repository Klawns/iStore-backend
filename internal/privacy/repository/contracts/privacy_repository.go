package contracts

import "istore/internal/privacy/domain"

type PrivacyRepository interface {
	Create(request *domain.PrivacyRequest) error
	ListByUserID(userID uint) ([]domain.PrivacyRequest, error)
}
