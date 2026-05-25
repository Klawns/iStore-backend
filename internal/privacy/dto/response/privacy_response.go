package response

import (
	"istore/internal/privacy/domain"
	"time"
)

type PrivacyRequestResponse struct {
	ID        uint                 `json:"id"`
	Type      domain.RequestType   `json:"type"`
	Status    domain.RequestStatus `json:"status"`
	Message   string               `json:"message"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

func FromDomain(request *domain.PrivacyRequest) *PrivacyRequestResponse {
	if request == nil {
		return nil
	}

	return &PrivacyRequestResponse{
		ID:        request.ID,
		Type:      request.Type,
		Status:    request.Status,
		Message:   request.Message,
		CreatedAt: request.CreatedAt,
		UpdatedAt: request.UpdatedAt,
	}
}

func ListFromDomain(requests []domain.PrivacyRequest) []PrivacyRequestResponse {
	response := make([]PrivacyRequestResponse, len(requests))
	for i := range requests {
		response[i] = *FromDomain(&requests[i])
	}

	return response
}
