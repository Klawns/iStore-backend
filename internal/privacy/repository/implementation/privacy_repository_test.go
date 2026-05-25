package implementation

import (
	"istore/internal/privacy/domain"
	privacyEntity "istore/internal/privacy/repository/entity"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateAndListRequestsByUser(t *testing.T) {
	db := newPrivacyTestDB(t)
	repository := NewPrivacyRepository(db)

	first := &domain.PrivacyRequest{UserID: 1, Type: domain.RequestAccess, Status: domain.RequestOpen, Message: "dados"}
	second := &domain.PrivacyRequest{UserID: 2, Type: domain.RequestDeletion, Status: domain.RequestOpen, Message: "excluir"}

	if err := repository.Create(first); err != nil {
		t.Fatalf("create first request: %v", err)
	}
	if err := repository.Create(second); err != nil {
		t.Fatalf("create second request: %v", err)
	}

	requests, err := repository.ListByUserID(1)
	if err != nil {
		t.Fatalf("list requests: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one request, got %d", len(requests))
	}
	if requests[0].UserID != 1 || requests[0].Type != domain.RequestAccess || requests[0].Status != domain.RequestOpen {
		t.Fatalf("unexpected request: %+v", requests[0])
	}
}

func newPrivacyTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&privacyEntity.PrivacyRequestEntity{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db
}
