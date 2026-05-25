package implementation

import (
	"encoding/json"
	customerEntity "istore/internal/customer/repository/entity"
	"istore/internal/privacy/domain"
	privacyEntity "istore/internal/privacy/repository/entity"
	saleDomain "istore/internal/sale/domain"
	saleEntity "istore/internal/sale/repository/entity"
	userEntity "istore/internal/users/repository/entity"
	"strings"
	"testing"
	"time"

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

func TestExportByUserIDReturnsOnlyOwnedDataWithoutPasswordHash(t *testing.T) {
	db := newPrivacyTestDB(t)
	repository := NewPrivacyRepository(db)

	user := seedPrivacyUser(t, db, "owner@example.com", "secret-hash")
	otherUser := seedPrivacyUser(t, db, "other@example.com", "other-hash")
	customerID := seedPrivacyCustomer(t, db, user.ID, "Ana")
	otherCustomerID := seedPrivacyCustomer(t, db, otherUser.ID, "Bruno")
	saleID := seedPrivacySale(t, db, user.ID, customerID, "iPhone")
	seedPrivacySale(t, db, otherUser.ID, otherCustomerID, "MacBook")
	seedPrivacyInstallment(t, db, saleID)

	export, err := repository.ExportByUserID(user.ID)
	if err != nil {
		t.Fatalf("export user data: %v", err)
	}

	if export.Account.ID != user.ID || export.Account.Email != "owner@example.com" {
		t.Fatalf("unexpected account export: %+v", export.Account)
	}
	if len(export.Customers) != 1 || export.Customers[0].Name != "Ana" {
		t.Fatalf("expected only owned customer, got %+v", export.Customers)
	}
	if len(export.Sales) != 1 || export.Sales[0].Items[0].ProductName != "iPhone" {
		t.Fatalf("expected only owned sale with item, got %+v", export.Sales)
	}
	if len(export.Sales[0].InstallmentsList) != 1 {
		t.Fatalf("expected owned installment, got %+v", export.Sales[0].InstallmentsList)
	}

	payload, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("marshal export: %v", err)
	}
	if strings.Contains(string(payload), "secret-hash") || strings.Contains(string(payload), "other-hash") || strings.Contains(string(payload), "other@example.com") {
		t.Fatalf("export leaked another user or password hash: %s", payload)
	}
}

func newPrivacyTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&userEntity.UserEntity{},
		&customerEntity.CustomerEntity{},
		&saleEntity.SaleEntity{},
		&saleEntity.SaleItemEntity{},
		&saleEntity.SaleInstallmentEntity{},
		&privacyEntity.PrivacyRequestEntity{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db
}

func seedPrivacyUser(t *testing.T, db *gorm.DB, email string, passwordHash string) userEntity.UserEntity {
	t.Helper()

	now := time.Now().UTC()
	user := userEntity.UserEntity{
		Email:                email,
		PasswordHash:         passwordHash,
		PrivacyPolicyVersion: "2026-05-01",
		PrivacyAcceptedAt:    &now,
		TermsVersion:         "2026-05-01",
		TermsAcceptedAt:      &now,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return user
}

func seedPrivacyCustomer(t *testing.T, db *gorm.DB, userID uint, name string) int {
	t.Helper()

	customer := customerEntity.CustomerEntity{UserID: userID, Name: name, Phone: "11999999999"}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	return customer.ID
}

func seedPrivacySale(t *testing.T, db *gorm.DB, userID uint, customerID int, productName string) int {
	t.Helper()

	sale := saleEntity.SaleEntity{
		UserID:        userID,
		CustomerID:    customerID,
		TotalValue:    200,
		PaymentStatus: saleDomain.PaymentApproved,
		PaymentType:   saleDomain.Pix,
		SaleDate:      time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&sale).Error; err != nil {
		t.Fatalf("seed sale: %v", err)
	}

	item := saleEntity.SaleItemEntity{
		SaleID:      sale.ID,
		ProductName: productName,
		Quantity:    1,
		CostPrice:   100,
		SalePrice:   200,
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed sale item: %v", err)
	}

	return sale.ID
}

func seedPrivacyInstallment(t *testing.T, db *gorm.DB, saleID int) {
	t.Helper()

	installment := saleEntity.SaleInstallmentEntity{
		SaleID:            saleID,
		DueDate:           time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC),
		InstallmentNumber: 1,
		TotalInstallments: 1,
		Amount:            200,
		Status:            saleDomain.InstallmentPending,
	}
	if err := db.Create(&installment).Error; err != nil {
		t.Fatalf("seed installment: %v", err)
	}
}
