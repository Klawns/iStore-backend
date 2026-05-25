package implementation

import (
	customerEntity "istore/internal/customer/repository/entity"
	privacyDomain "istore/internal/privacy/domain"
	privacyEntity "istore/internal/privacy/repository/entity"
	saleDomain "istore/internal/sale/domain"
	saleEntity "istore/internal/sale/repository/entity"
	userEntity "istore/internal/users/repository/entity"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeleteOwnAccountDeletesOnlyAuthenticatedUserData(t *testing.T) {
	db := newUserRepositoryTestDB(t)
	repository := NewUserRepository(db)

	user := seedUserEntity(t, db, "user@example.com")
	otherUser := seedUserEntity(t, db, "other@example.com")
	customer := seedUserCustomer(t, db, user.ID, "Ana")
	otherCustomer := seedUserCustomer(t, db, otherUser.ID, "Bruno")
	sale := seedUserSale(t, db, user.ID, customer.ID)
	otherSale := seedUserSale(t, db, otherUser.ID, otherCustomer.ID)
	seedSaleChildren(t, db, sale.ID)
	seedSaleChildren(t, db, otherSale.ID)
	seedPrivacyRequest(t, db, user.ID)
	seedPrivacyRequest(t, db, otherUser.ID)

	if err := repository.DeleteOwnAccount(user.ID); err != nil {
		t.Fatalf("delete own account: %v", err)
	}

	assertCount(t, db, &userEntity.UserEntity{}, "id = ?", user.ID, 0)
	assertCount(t, db, &customerEntity.CustomerEntity{}, "user_id = ?", user.ID, 0)
	assertCount(t, db, &saleEntity.SaleEntity{}, "user_id = ?", user.ID, 0)
	assertCount(t, db, &saleEntity.SaleItemEntity{}, "sale_id = ?", sale.ID, 0)
	assertCount(t, db, &saleEntity.SaleInstallmentEntity{}, "sale_id = ?", sale.ID, 0)
	assertCount(t, db, &saleEntity.PaymentAlertEntity{}, "sale_id = ?", sale.ID, 0)
	assertCount(t, db, &privacyEntity.PrivacyRequestEntity{}, "user_id = ?", user.ID, 0)

	assertCount(t, db, &userEntity.UserEntity{}, "id = ?", otherUser.ID, 1)
	assertCount(t, db, &customerEntity.CustomerEntity{}, "user_id = ?", otherUser.ID, 1)
	assertCount(t, db, &saleEntity.SaleEntity{}, "user_id = ?", otherUser.ID, 1)
	assertCount(t, db, &saleEntity.SaleItemEntity{}, "sale_id = ?", otherSale.ID, 1)
	assertCount(t, db, &saleEntity.SaleInstallmentEntity{}, "sale_id = ?", otherSale.ID, 1)
	assertCount(t, db, &saleEntity.PaymentAlertEntity{}, "sale_id = ?", otherSale.ID, 1)
	assertCount(t, db, &privacyEntity.PrivacyRequestEntity{}, "user_id = ?", otherUser.ID, 1)
}

func newUserRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&userEntity.UserEntity{},
		&customerEntity.CustomerEntity{},
		&saleEntity.SaleEntity{},
		&saleEntity.SaleItemEntity{},
		&saleEntity.SaleInstallmentEntity{},
		&saleEntity.PaymentAlertEntity{},
		&privacyEntity.PrivacyRequestEntity{},
	); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	return db
}

func seedUserEntity(t *testing.T, db *gorm.DB, email string) userEntity.UserEntity {
	t.Helper()

	user := userEntity.UserEntity{Email: email, PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}

func seedUserCustomer(t *testing.T, db *gorm.DB, userID uint, name string) customerEntity.CustomerEntity {
	t.Helper()

	customer := customerEntity.CustomerEntity{UserID: userID, Name: name, Phone: "11999990000"}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	return customer
}

func seedUserSale(t *testing.T, db *gorm.DB, userID uint, customerID int) saleEntity.SaleEntity {
	t.Helper()

	sale := saleEntity.SaleEntity{
		UserID:        userID,
		CustomerID:    customerID,
		TotalValue:    100,
		PaymentStatus: saleDomain.PaymentApproved,
		PaymentType:   saleDomain.Pix,
		SaleDate:      time.Now(),
	}
	if err := db.Create(&sale).Error; err != nil {
		t.Fatalf("seed sale: %v", err)
	}
	return sale
}

func seedSaleChildren(t *testing.T, db *gorm.DB, saleID int) {
	t.Helper()

	if err := db.Create(&saleEntity.SaleItemEntity{
		SaleID:      saleID,
		ProductName: "iPhone",
		Quantity:    1,
		CostPrice:   10,
		SalePrice:   100,
	}).Error; err != nil {
		t.Fatalf("seed sale item: %v", err)
	}
	if err := db.Create(&saleEntity.SaleInstallmentEntity{
		SaleID:            saleID,
		DueDate:           time.Now(),
		InstallmentNumber: 1,
		TotalInstallments: 1,
		Amount:            100,
		Status:            saleDomain.InstallmentPending,
	}).Error; err != nil {
		t.Fatalf("seed sale installment: %v", err)
	}
	if err := db.Create(&saleEntity.PaymentAlertEntity{
		SaleID:            saleID,
		DueDate:           time.Now(),
		InstallmentNumber: 1,
		Message:           "Vencimento proximo",
		Status:            saleDomain.PaymentAlertOpen,
	}).Error; err != nil {
		t.Fatalf("seed payment alert: %v", err)
	}
}

func seedPrivacyRequest(t *testing.T, db *gorm.DB, userID uint) {
	t.Helper()

	request := privacyEntity.PrivacyRequestEntity{
		UserID:  userID,
		Type:    privacyDomain.RequestAccess,
		Status:  privacyDomain.RequestOpen,
		Message: "Acesso",
	}
	if err := db.Create(&request).Error; err != nil {
		t.Fatalf("seed privacy request: %v", err)
	}
}

func assertCount(t *testing.T, db *gorm.DB, model any, query string, value any, expected int64) {
	t.Helper()

	var count int64
	if err := db.Model(model).Where(query, value).Count(&count).Error; err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	if count != expected {
		t.Fatalf("expected %d records for %T, got %d", expected, model, count)
	}
}
