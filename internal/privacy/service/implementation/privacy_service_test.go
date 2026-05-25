package implementation

import (
	"istore/internal/privacy/domain"
	serviceContracts "istore/internal/privacy/service/contracts"
	"testing"
)

type fakePrivacyRepository struct {
	created *domain.PrivacyRequest
}

func (f *fakePrivacyRepository) Create(request *domain.PrivacyRequest) error {
	request.ID = 1
	f.created = request
	return nil
}
func (f *fakePrivacyRepository) ListByUserID(userID uint) ([]domain.PrivacyRequest, error) {
	return []domain.PrivacyRequest{{ID: 1, UserID: userID, Type: domain.RequestAccess}}, nil
}
func (f *fakePrivacyRepository) ExportByUserID(userID uint) (*domain.PrivacyExport, error) {
	return &domain.PrivacyExport{Account: domain.AccountExport{ID: userID}}, nil
}

func TestCreateRequestDefaultsStatusOpen(t *testing.T) {
	repository := &fakePrivacyRepository{}
	service := NewPrivacyService(repository)

	request, restErr := service.CreateRequest(serviceContracts.CreatePrivacyRequestInput{
		UserID:  1,
		Type:    domain.RequestDeletion,
		Message: " revisar exclusao ",
	})
	if restErr != nil {
		t.Fatalf("create request: %v", restErr)
	}
	if request.Status != domain.RequestOpen || request.Message != "revisar exclusao" {
		t.Fatalf("unexpected request: %+v", request)
	}
}

func TestCreateRequestRejectsInvalidType(t *testing.T) {
	service := NewPrivacyService(&fakePrivacyRepository{})

	if _, restErr := service.CreateRequest(serviceContracts.CreatePrivacyRequestInput{
		UserID: 1,
		Type:   domain.RequestType("INVALID"),
	}); restErr == nil {
		t.Fatal("expected validation error")
	}
}
