package implementation

import (
	"istore/internal/users/domain"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepository struct {
	user    *domain.User
	deleted bool
}

func (f *fakeUserRepository) Create(user *domain.User) error { return nil }
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
