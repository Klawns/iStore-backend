package main

import (
	"istore/internal/sale/domain"
	saleEntity "istore/internal/sale/repository/entity"
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

	if err := db.Exec(`
		INSERT INTO sales (id, customer_id, total_value, payment_status, payment_type, sale_date, installments, billing_day)
		VALUES (1, 1, 1000, 'APPROVED', 'CARD', '2026-05-23 12:00:00', NULL, NULL)
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
