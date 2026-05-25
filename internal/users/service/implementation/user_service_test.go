package implementation

import (
	"istore/internal/users/domain"
	"istore/internal/users/service/contracts"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepository struct {
	user    *domain.User
	created *domain.User
	deleted bool
}

func (f *fakeUserRepository) Create(user *domain.User) error {
	clone := *user
	f.created = &clone
	return nil
}
func (f *fakeUserRepository) FindByEmail(email string) (*domain.User, error) {
	return nil, nil
}
func (f *fakeUserRepository) FindByID(id uint) (*domain.User, error) {
	return f.user, nil
}
func (f *fakeUserRepository) DeleteOwnAccount(id uint) error {
	f.deleted = true
	return nil
}

func TestCreateRejectsMissingLegalAcceptance(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)

	_, restErr := service.Create(contracts.CreateUserInput{
		Email:               "user@example.com",
		Password:            "senha-segura",
		AcceptPrivacyPolicy: true,
		AcceptTerms:         false,
	})
	if restErr == nil || restErr.Code != 400 {
		t.Fatalf("expected bad request, got %#v", restErr)
	}
	if repository.created != nil {
		t.Fatal("expected user not to be created")
	}
}

func TestCreateUsesServerConsentVersionsAndSingleUTCTimestamp(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)
	before := time.Now().UTC()

	_, restErr := service.Create(contracts.CreateUserInput{
		Email:                " USER@example.com ",
		Password:             "senha-segura",
		AcceptPrivacyPolicy:  true,
		AcceptTerms:          true,
		PrivacyPolicyVersion: "client-privacy-version",
		TermsVersion:         "client-terms-version",
	})
	after := time.Now().UTC()

	if restErr != nil {
		t.Fatalf("expected user creation, got %#v", restErr)
	}
	if repository.created == nil {
		t.Fatal("expected user to be created")
	}
	if repository.created.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", repository.created.Email)
	}
	if repository.created.PrivacyPolicyVersion != currentPrivacyPolicyVersion {
		t.Fatalf("expected server privacy version, got %q", repository.created.PrivacyPolicyVersion)
	}
	if repository.created.TermsVersion != currentTermsVersion {
		t.Fatalf("expected server terms version, got %q", repository.created.TermsVersion)
	}
	if repository.created.PrivacyAcceptedAt == nil || repository.created.TermsAcceptedAt == nil {
		t.Fatal("expected consent timestamps")
	}
	if !repository.created.PrivacyAcceptedAt.Equal(*repository.created.TermsAcceptedAt) {
		t.Fatalf("expected a single consent timestamp, got %s and %s", repository.created.PrivacyAcceptedAt, repository.created.TermsAcceptedAt)
	}
	if repository.created.PrivacyAcceptedAt.Location() != time.UTC {
		t.Fatalf("expected UTC timestamp, got %s", repository.created.PrivacyAcceptedAt.Location())
	}
	if repository.created.PrivacyAcceptedAt.Before(before) || repository.created.PrivacyAcceptedAt.After(after) {
		t.Fatalf("expected timestamp between %s and %s, got %s", before, after, repository.created.PrivacyAcceptedAt)
	}
}

func TestDeleteOwnAccountRejectsInvalidPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repository := &fakeUserRepository{
		user: &domain.User{ID: 1, Email: "user@example.com", PasswordHash: string(hash)},
	}
	service := NewUserService(repository)

	restErr := service.DeleteOwnAccount(1, "senha-errada")
	if restErr == nil || restErr.Code != 401 {
		t.Fatalf("expected unauthorized, got %#v", restErr)
	}
	if repository.deleted {
		t.Fatal("expected account to remain")
	}
}
