package implementation

import (
	"istore/internal/privacy/domain"
	"istore/internal/privacy/repository/contracts"
	serviceContracts "istore/internal/privacy/service/contracts"
	"istore/pkg/rest_err"
	"strings"
)

type privacyService struct {
	repository contracts.PrivacyRepository
}

func NewPrivacyService(repository contracts.PrivacyRepository) serviceContracts.PrivacyService {
	return &privacyService{repository: repository}
}

func (s *privacyService) CreateRequest(input serviceContracts.CreatePrivacyRequestInput) (*domain.PrivacyRequest, *rest_err.RestErr) {
	if input.UserID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("usuario invalido")
	}
	if !isValidRequestType(input.Type) {
		return nil, rest_err.NewBadRequestError("tipo de solicitacao invalido")
	}

	request := &domain.PrivacyRequest{
		UserID:  input.UserID,
		Type:    input.Type,
		Status:  domain.RequestOpen,
		Message: strings.TrimSpace(input.Message),
	}
	if err := s.repository.Create(request); err != nil {
		return nil, rest_err.NewInternalServerError("erro ao criar solicitacao LGPD")
	}

	return request, nil
}

func (s *privacyService) ListRequests(userID uint) ([]domain.PrivacyRequest, *rest_err.RestErr) {
	if userID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("usuario invalido")
	}

	requests, err := s.repository.ListByUserID(userID)
	if err != nil {
		return nil, rest_err.NewInternalServerError("erro ao listar solicitacoes LGPD")
	}

	return requests, nil
}

func isValidRequestType(requestType domain.RequestType) bool {
	switch requestType {
	case domain.RequestAccess, domain.RequestCorrection, domain.RequestExport, domain.RequestDeletion, domain.RequestPortability:
		return true
	default:
		return false
	}
}
