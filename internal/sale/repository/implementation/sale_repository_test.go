package implementation

import (
	customerEntity "istore/internal/customer/repository/entity"
	"istore/internal/sale/domain"
	"istore/internal/sale/repository/entity"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreatePersistsSaleInstallments(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&customerEntity.CustomerEntity{}, &entity.SaleEntity{}, &entity.SaleItemEntity{}, &entity.SaleInstallmentEntity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	customerID := seedCustomer(t, db, "Ana Silva")

	repository := NewSaleRepository(db)
	installments := 2
	billingDay := 25
	sale := &domain.Sale{
		UserID:        1,
		CustomerID:    customerID,
		TotalValue:    101,
		PaymentStatus: domain.PaymentPending,
		PaymentType:   domain.CreditCard,
		SaleDate:      time.Date(2026, time.May, 23, 12, 0, 0, 0, time.UTC),
		Installments:  &installments,
		BillingDay:    &billingDay,
		Items: []domain.SaleItem{
			{ProductName: "iPhone", Quantity: 1, CostPrice: 50, SalePrice: 101},
		},
		InstallmentsList: []domain.SaleInstallment{
			{DueDate: time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC), InstallmentNumber: 1, TotalInstallments: 2, Amount: 51, Status: domain.InstallmentPending},
			{DueDate: time.Date(2026, time.June, 25, 0, 0, 0, 0, time.UTC), InstallmentNumber: 2, TotalInstallments: 2, Amount: 50, Status: domain.InstallmentPending},
		},
	}

	if err := repository.Create(sale); err != nil {
		t.Fatalf("create sale: %v", err)
	}

	var count int64
	if err := db.Model(&entity.SaleInstallmentEntity{}).Count(&count).Error; err != nil {
		t.Fatalf("count installments: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 installments, got %d", count)
	}
}

func TestListFiltersSalesAndSummarizesAllFilteredRows(t *testing.T) {
	db := newSaleListTestDB(t)
	repository := NewSaleRepository(db)
	firstCustomerID := seedCustomer(t, db, "Ana Silva")
	secondCustomerID := seedCustomer(t, db, "Bruno Costa")
	seedSaleWithItems(t, db, firstCustomerID, domain.PaymentApproved, domain.Pix, time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC), []entity.SaleItemEntity{
		{ProductName: "iPhone 15", Specs: "128GB", Quantity: 1, CostPrice: 300000, SalePrice: 450000},
	})
	seedSaleWithItems(t, db, firstCustomerID, domain.PaymentPending, domain.CreditCard, time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC), []entity.SaleItemEntity{
		{ProductName: "MacBook Air", Specs: "M3", Quantity: 1, CostPrice: 700000, SalePrice: 900000},
	})
	seedSaleWithItems(t, db, secondCustomerID, domain.PaymentApproved, domain.Pix, time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC), []entity.SaleItemEntity{
		{ProductName: "AirPods", Specs: "Pro", Quantity: 2, CostPrice: 80000, SalePrice: 120000},
	})

	start := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 31, 23, 59, 59, 0, time.UTC)
	status := domain.PaymentApproved
	paymentType := domain.Pix
	result, err := repository.List(domain.SaleListFilter{
		UserID:        1,
		Page:          1,
		Limit:         10,
		Start:         &start,
		End:           &end,
		PaymentStatus: &status,
		PaymentType:   &paymentType,
		CustomerID:    &firstCustomerID,
	})
	if err != nil {
		t.Fatalf("list sales: %v", err)
	}

	if result.TotalItems != 1 || len(result.Items) != 1 {
		t.Fatalf("expected one filtered sale, got total=%d items=%d", result.TotalItems, len(result.Items))
	}
	if result.Items[0].CustomerName != "Ana Silva" {
		t.Fatalf("expected customer name loaded, got %q", result.Items[0].CustomerName)
	}
	if result.Summary.Revenue != 450000 || result.Summary.Profit != 150000 || result.Summary.AverageTicket != 450000 {
		t.Fatalf("unexpected summary: %+v", result.Summary)
	}
}

func TestListSearchesCustomerProductAndSpecs(t *testing.T) {
	db := newSaleListTestDB(t)
	repository := NewSaleRepository(db)
	customerID := seedCustomer(t, db, "Carla Mendes")
	seedSaleWithItems(t, db, customerID, domain.PaymentApproved, domain.Pix, time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC), []entity.SaleItemEntity{
		{ProductName: "Apple Watch", Specs: "GPS", Quantity: 1, CostPrice: 100000, SalePrice: 160000},
	})

	customerResult, err := repository.List(domain.SaleListFilter{UserID: 1, Page: 1, Limit: 10, Search: "carla"})
	if err != nil {
		t.Fatalf("search customer: %v", err)
	}
	if customerResult.TotalItems != 1 {
		t.Fatalf("expected customer search to match one sale, got %d", customerResult.TotalItems)
	}

	productResult, err := repository.List(domain.SaleListFilter{UserID: 1, Page: 1, Limit: 10, Search: "watch"})
	if err != nil {
		t.Fatalf("search product: %v", err)
	}
	if productResult.TotalItems != 1 {
		t.Fatalf("expected product search to match one sale, got %d", productResult.TotalItems)
	}

	specsResult, err := repository.List(domain.SaleListFilter{UserID: 1, Page: 1, Limit: 10, Search: "gps"})
	if err != nil {
		t.Fatalf("search specs: %v", err)
	}
	if specsResult.TotalItems != 1 {
		t.Fatalf("expected specs search to match one sale, got %d", specsResult.TotalItems)
	}
}

func TestListPaginatesAndOrdersSales(t *testing.T) {
	db := newSaleListTestDB(t)
	repository := NewSaleRepository(db)
	customerID := seedCustomer(t, db, "Diego Ramos")

	var newestID int
	for i := 0; i < 12; i++ {
		sale := seedSaleWithItems(t, db, customerID, domain.PaymentApproved, domain.Pix, time.Date(2026, time.May, i+1, 12, 0, 0, 0, time.UTC), []entity.SaleItemEntity{
			{ProductName: "Produto", Quantity: 1, CostPrice: 100, SalePrice: 200 + i},
		})
		if i == 11 {
			newestID = sale.ID
		}
	}

	firstPage, err := repository.List(domain.SaleListFilter{UserID: 1, Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if firstPage.TotalItems != 12 || firstPage.TotalPages != 2 || len(firstPage.Items) != 10 {
		t.Fatalf("unexpected pagination: total=%d pages=%d items=%d", firstPage.TotalItems, firstPage.TotalPages, len(firstPage.Items))
	}
	if firstPage.Items[0].ID != newestID {
		t.Fatalf("expected newest sale first, got id %d want %d", firstPage.Items[0].ID, newestID)
	}

	secondPage, err := repository.List(domain.SaleListFilter{UserID: 1, Page: 2, Limit: 10})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 2 {
		t.Fatalf("expected two items on second page, got %d", len(secondPage.Items))
	}
}

func newSaleListTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&customerEntity.CustomerEntity{}, &entity.SaleEntity{}, &entity.SaleItemEntity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db
}

func seedCustomer(t *testing.T, db *gorm.DB, name string) int {
	t.Helper()

	customer := customerEntity.CustomerEntity{UserID: 1, Name: name}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	return customer.ID
}

func seedSaleWithItems(t *testing.T, db *gorm.DB, customerID int, status domain.PaymentStatus, paymentType domain.PaymentType, saleDate time.Time, items []entity.SaleItemEntity) entity.SaleEntity {
	t.Helper()

	total := 0
	for _, item := range items {
		total += item.SalePrice * item.Quantity
	}

	sale := entity.SaleEntity{
		UserID:        1,
		CustomerID:    customerID,
		TotalValue:    total,
		PaymentStatus: status,
		PaymentType:   paymentType,
		SaleDate:      saleDate,
	}
	if err := db.Create(&sale).Error; err != nil {
		t.Fatalf("seed sale: %v", err)
	}

	for i := range items {
		items[i].SaleID = sale.ID
		if err := db.Create(&items[i]).Error; err != nil {
			t.Fatalf("seed sale item: %v", err)
		}
	}

	return sale
}

func TestUpdateInstallmentStatusApprovesSaleWhenAllPaid(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entity.SaleEntity{}, &entity.SaleInstallmentEntity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repository := NewSaleRepository(db)
	sale := entity.SaleEntity{
		UserID:        1,
		CustomerID:    1,
		TotalValue:    100,
		PaymentStatus: domain.PaymentPending,
		PaymentType:   domain.CreditCard,
		SaleDate:      time.Date(2026, time.May, 23, 12, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&sale).Error; err != nil {
		t.Fatalf("seed sale: %v", err)
	}
	first := entity.SaleInstallmentEntity{
		SaleID:            sale.ID,
		DueDate:           time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC),
		InstallmentNumber: 1,
		TotalInstallments: 2,
		Amount:            50,
		Status:            domain.InstallmentPaid,
	}
	second := entity.SaleInstallmentEntity{
		SaleID:            sale.ID,
		DueDate:           time.Date(2026, time.June, 25, 0, 0, 0, 0, time.UTC),
		InstallmentNumber: 2,
		TotalInstallments: 2,
		Amount:            50,
		Status:            domain.InstallmentPending,
	}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("seed first installment: %v", err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("seed second installment: %v", err)
	}

	if _, err := repository.UpdateInstallmentStatus(1, second.ID, domain.InstallmentPaid, "", time.Now()); err != nil {
		t.Fatalf("pay installment: %v", err)
	}

	var updatedSale entity.SaleEntity
	if err := db.First(&updatedSale, sale.ID).Error; err != nil {
		t.Fatalf("load sale: %v", err)
	}
	if updatedSale.PaymentStatus != domain.PaymentApproved {
		t.Fatalf("expected sale approved, got %s", updatedSale.PaymentStatus)
	}
}
