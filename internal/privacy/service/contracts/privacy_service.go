package contracts

import (
	"istore/internal/privacy/domain"
	"istore/pkg/rest_err"
)

type CreatePrivacyRequestInput struct {
	UserID  uint
	Type    domain.RequestType
	Message string
}

type PrivacyService interface {
	CreateRequest(input CreatePrivacyRequestInput) (*domain.PrivacyRequest, *rest_err.RestErr)
	ListRequests(userID uint) ([]domain.PrivacyRequest, *rest_err.RestErr)
	Export(userID uint) (*domain.PrivacyExport, *rest_err.RestErr)
}
