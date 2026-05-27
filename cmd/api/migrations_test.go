package main

import (
	customerEntity "istore/internal/customer/repository/entity"
	"istore/internal/sale/domain"
	saleEntity "istore/internal/sale/repository/entity"
	userEntity "istore/internal/users/repository/entity"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunMigrationsBackfillsLegacyCardSales(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("initial migrations: %v", err)
	}

	if err := db.Create(&userEntity.UserEntity{
		ID:           1,
		Email:        "owner@example.com",
		PasswordHash: "hash",
	}).Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO sales (id, user_id, customer_id, total_value, payment_status, payment_type, sale_date, installments, billing_day)
		VALUES (1, 1, 1, 1000, 'APPROVED', 'CARD', '2026-05-23 12:00:00', NULL, NULL)
	`).Error; err != nil {
		t.Fatalf("insert legacy sale: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	var sale saleEntity.SaleEntity
	if err := db.First(&sale, 1).Error; err != nil {
		t.Fatalf("load sale: %v", err)
	}
	if sale.PaymentType != domain.CreditCard {
		t.Fatalf("expected payment type %s, got %s", domain.CreditCard, sale.PaymentType)
	}
	if sale.Installments == nil || *sale.Installments != 1 {
		t.Fatalf("expected installments to be backfilled with 1, got %#v", sale.Installments)
	}
	if sale.BillingDay != nil {
		t.Fatalf("expected billing day to remain nil, got %#v", sale.BillingDay)
	}
}

func TestRunMigrationsDeletesOrphanUserData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("initial migrations: %v", err)
	}

	if err := db.Create(&userEntity.UserEntity{
		ID:           1,
		Email:        "owner@example.com",
		PasswordHash: "hash",
	}).Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := db.Create(&customerEntity.CustomerEntity{
		ID:     1,
		UserID: 1,
		Name:   "Owned Customer",
		Phone:  "11999999999",
	}).Error; err != nil {
		t.Fatalf("insert owned customer: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO customers (id, user_id, name, phone, created_at, updated_at)
		VALUES
			(2, 0, 'Zero User Customer', '11000000000', '2026-05-23 12:00:00', '2026-05-23 12:00:00'),
			(3, NULL, 'Null User Customer', '11000000001', '2026-05-23 12:00:00', '2026-05-23 12:00:00')
	`).Error; err != nil {
		t.Fatalf("insert orphan customers: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO sales (id, user_id, customer_id, total_value, payment_status, payment_type, sale_date)
		VALUES
			(1, 1, 1, 1000, 'APPROVED', 'PIX', '2026-05-23 12:00:00'),
			(2, 0, 1, 2000, 'APPROVED', 'PIX', '2026-05-23 12:00:00'),
			(3, NULL, 1, 3000, 'APPROVED', 'PIX', '2026-05-23 12:00:00')
	`).Error; err != nil {
		t.Fatalf("insert sales: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO sale_items (id, sale_id, product_name, specs, quantity, cost_price, sale_price)
		VALUES
			(1, 2, 'Orphan Item Zero', '', 1, 100, 200),
			(2, 3, 'Orphan Item Null', '', 1, 100, 300)
	`).Error; err != nil {
		t.Fatalf("insert sale items: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO sale_installments (
			id, sale_id, due_date, installment_number, total_installments, amount, status, created_at, updated_at
		)
		VALUES
			(1, 2, '2026-06-10 00:00:00', 1, 1, 2000, 'PENDING', '2026-05-23 12:00:00', '2026-05-23 12:00:00'),
			(2, 3, '2026-06-10 00:00:00', 1, 1, 3000, 'PENDING', '2026-05-23 12:00:00', '2026-05-23 12:00:00')
	`).Error; err != nil {
		t.Fatalf("insert sale installments: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO payment_alerts (
			id, sale_id, due_date, installment_number, message, status, created_at, updated_at
		)
		VALUES
			(1, 2, '2026-06-10 00:00:00', 1, 'Pay zero', 'OPEN', '2026-05-23 12:00:00', '2026-05-23 12:00:00'),
			(2, 3, '2026-06-10 00:00:00', 1, 'Pay null', 'OPEN', '2026-05-23 12:00:00', '2026-05-23 12:00:00')
	`).Error; err != nil {
		t.Fatalf("insert payment alerts: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	assertCount(t, db, &customerEntity.CustomerEntity{}, 1)
	assertCount(t, db, &saleEntity.SaleEntity{}, 1)
	assertCount(t, db, &saleEntity.SaleItemEntity{}, 0)
	assertCount(t, db, &saleEntity.SaleInstallmentEntity{}, 0)
	assertCount(t, db, &saleEntity.PaymentAlertEntity{}, 0)

	var ownedSale saleEntity.SaleEntity
	if err := db.First(&ownedSale, 1).Error; err != nil {
		t.Fatalf("load owned sale: %v", err)
	}
	if ownedSale.UserID != 1 {
		t.Fatalf("expected owned sale user_id to remain 1, got %d", ownedSale.UserID)
	}
}

func assertCount(t *testing.T, db *gorm.DB, model any, expected int64) {
	t.Helper()

	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	if count != expected {
		t.Fatalf("expected %T count %d, got %d", model, expected, count)
	}
}
