package entity

import (
	privacyDomain "istore/internal/privacy/domain"
	"time"
)

type PrivacyRequestEntity struct {
	ID        uint                        `gorm:"primaryKey"`
	UserID    uint                        `gorm:"column:user_id;not null;index"`
	Type      privacyDomain.RequestType   `gorm:"column:type;not null"`
	Status    privacyDomain.RequestStatus `gorm:"column:status;not null;index"`
	Message   string                      `gorm:"column:message"`
	CreatedAt time.Time                   `gorm:"column:created_at"`
	UpdatedAt time.Time                   `gorm:"column:updated_at"`
}

func (PrivacyRequestEntity) TableName() string {
	return "lgpd_requests"
}

func FromDomain(request *privacyDomain.PrivacyRequest) *PrivacyRequestEntity {
	if request == nil {
		return nil
	}

	return &PrivacyRequestEntity{
		ID:      request.ID,
		UserID:  request.UserID,
		Type:    request.Type,
		Status:  request.Status,
		Message: request.Message,
	}
}

func (r *PrivacyRequestEntity) ToDomain() *privacyDomain.PrivacyRequest {
	if r == nil {
		return nil
	}

	return &privacyDomain.PrivacyRequest{
		ID:        r.ID,
		UserID:    r.UserID,
		Type:      r.Type,
		Status:    r.Status,
		Message:   r.Message,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
