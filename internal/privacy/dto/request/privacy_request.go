package request

import "istore/internal/privacy/domain"

type PrivacyRequest struct {
	Type    domain.RequestType `json:"type" binding:"required"`
	Message string             `json:"message"`
}
