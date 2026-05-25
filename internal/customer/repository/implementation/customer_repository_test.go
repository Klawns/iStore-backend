package implementation

import (
	"istore/internal/customer/domain"
	customerEntity "istore/internal/customer/repository/entity"
	saleDomain "istore/internal/sale/domain"
	saleEntity "istore/internal/sale/repository/entity"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListPaginatesOrdersAndKeepsCustomersWithoutSales(t *testing.T) {
	db := newCustomerListTestDB(t)
	repository := NewCustomerRepository(db)

	seedCustomer(t, db, "Bruno Costa", "11999990000")
	seedCustomer(t, db, "Ana Silva", "21999990000")
	seedCustomer(t, db, "Carla Mendes", "31999990000")

	result, err := repository.List(domain.CustomerListFilter{UserID: 1, Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("list customers: %v", err)
	}

	if result.TotalItems != 3 || result.TotalPages != 2 || len(result.Items) != 2 {
		t.Fatalf("unexpected pagination: total=%d pages=%d items=%d", result.TotalItems, result.TotalPages, len(result.Items))
	}

	if result.Items[0].Name != "Ana Silva" || result.Items[1].Name != "Bruno Costa" {
		t.Fatalf("expected customers ordered by name, got %#v", result.Items)
	}

	if result.Summary.TotalCustomers != 3 {
		t.Fatalf("expected summary over all customers, got %d", result.Summary.TotalCustomers)
	}
}

func TestListSearchesByNameAndPhone(t *testing.T) {
	db := newCustomerListTestDB(t)
	repository := NewCustomerRepository(db)

	seedCustomer(t, db, "Ana Silva", "11912345678")
	seedCustomer(t, db, "Bruno Costa", "21987654321")

	nameResult, err := repository.List(domain.CustomerListFilter{UserID: 1, Page: 1, Limit: 10, Search: "ana"})
	if err != nil {
		t.Fatalf("search by name: %v", err)
	}
	if nameResult.TotalItems != 1 || nameResult.Items[0].Name != "Ana Silva" {
		t.Fatalf("expected Ana by name search, got %#v", nameResult.Items)
	}

	phoneResult, err := repository.List(domain.CustomerListFilter{UserID: 1, Page: 1, Limit: 10, Search: "9876"})
	if err != nil {
		t.Fatalf("search by phone: %v", err)
	}
	if phoneResult.TotalItems != 1 || phoneResult.Items[0].Name != "Bruno Costa" {
		t.Fatalf("expected Bruno by phone search, got %#v", phoneResult.Items)
	}
}

func TestListFiltersSalesAndSummarizesAllFilteredCustomers(t *testing.T) {
	db := newCustomerListTestDB(t)
	repository := NewCustomerRepository(db)

	anaID := seedCustomer(t, db, "Ana Silva", "11999990000")
	brunoID := seedCustomer(t, db, "Bruno Costa", "21999990000")
	seedCustomer(t, db, "Carla Mendes", "31999990000")

	seedSaleWithItems(t, db, anaID, saleDomain.PaymentApproved, saleDomain.Pix, time.Date(2026, time.May, 10, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "iPhone", Quantity: 1, CostPrice: 10000, SalePrice: 15000},
	})
	seedSaleWithItems(t, db, anaID, saleDomain.PaymentApproved, saleDomain.Pix, time.Date(2026, time.May, 11, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "AirPods", Quantity: 2, CostPrice: 1000, SalePrice: 2500},
	})
	seedSaleWithItems(t, db, brunoID, saleDomain.PaymentPending, saleDomain.CreditCard, time.Date(2026, time.May, 12, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "Watch", Quantity: 1, CostPrice: 4000, SalePrice: 6000},
	})
	seedSaleWithItems(t, db, brunoID, saleDomain.PaymentApproved, saleDomain.Money, time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "Case", Quantity: 1, CostPrice: 500, SalePrice: 1000},
	})

	start := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 31, 23, 59, 59, 0, time.UTC)
	status := saleDomain.PaymentApproved
	paymentType := saleDomain.Pix

	result, err := repository.List(domain.CustomerListFilter{
		UserID:        1,
		Page:          1,
		Limit:         1,
		Start:         &start,
		End:           &end,
		PaymentStatus: &status,
		PaymentType:   &paymentType,
	})
	if err != nil {
		t.Fatalf("list filtered customers: %v", err)
	}

	if result.TotalItems != 1 || result.TotalPages != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected filtered pagination: total=%d pages=%d items=%d", result.TotalItems, result.TotalPages, len(result.Items))
	}

	item := result.Items[0]
	if item.Name != "Ana Silva" || item.SalesCount != 2 || item.Revenue != 20000 || item.Profit != 8000 || item.AverageTicket != 10000 {
		t.Fatalf("unexpected customer metrics: %#v", item)
	}

	if result.Summary.TotalCustomers != 1 || result.Summary.SalesCount != 2 || result.Summary.Revenue != 20000 || result.Summary.Profit != 8000 || result.Summary.AverageTicket != 10000 || result.Summary.RepeatRate != 100 {
		t.Fatalf("unexpected summary: %#v", result.Summary)
	}
}

func TestListScopesCustomersByUser(t *testing.T) {
	db := newCustomerListTestDB(t)
	repository := NewCustomerRepository(db)

	seedCustomer(t, db, "Ana Silva", "11999990000")
	other := customerEntity.CustomerEntity{UserID: 2, Name: "Outro Usuario", Phone: "21999990000"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("seed other customer: %v", err)
	}

	result, err := repository.List(domain.CustomerListFilter{UserID: 1, Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("list customers: %v", err)
	}

	if result.TotalItems != 1 || result.Items[0].Name != "Ana Silva" {
		t.Fatalf("expected only user customers, got %#v", result.Items)
	}
}

func TestDeleteManyDeletesOnlySelectedCustomersFromUser(t *testing.T) {
	db := newCustomerListTestDB(t)
	repository := NewCustomerRepository(db)

	firstID := seedCustomer(t, db, "Ana Silva", "11999990000")
	secondID := seedCustomer(t, db, "Bruno Costa", "21999990000")
	thirdID := seedCustomer(t, db, "Carla Mendes", "31999990000")
	other := customerEntity.CustomerEntity{UserID: 2, Name: "Outro Usuario", Phone: "21999990000"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("seed other customer: %v", err)
	}

	if err := repository.DeleteMany(1, []int{firstID, thirdID, other.ID}); err != nil {
		t.Fatalf("delete many customers: %v", err)
	}

	var remaining []customerEntity.CustomerEntity
	if err := db.Order("id ASC").Find(&remaining).Error; err != nil {
		t.Fatalf("list remaining customers: %v", err)
	}
	if len(remaining) != 2 || remaining[0].ID != secondID || remaining[1].ID != other.ID {
		t.Fatalf("unexpected remaining customers: %#v", remaining)
	}
}

func TestCountSalesByCustomerIDsScopesByUser(t *testing.T) {
	db := newCustomerListTestDB(t)
	repository := NewCustomerRepository(db)

	customerID := seedCustomer(t, db, "Ana Silva", "11999990000")
	otherCustomer := customerEntity.CustomerEntity{UserID: 2, Name: "Outro Usuario", Phone: "21999990000"}
	if err := db.Create(&otherCustomer).Error; err != nil {
		t.Fatalf("seed other customer: %v", err)
	}
	seedSaleWithItems(t, db, customerID, saleDomain.PaymentApproved, saleDomain.Pix, time.Now(), nil)
	otherSale := saleEntity.SaleEntity{
		UserID:        2,
		CustomerID:    otherCustomer.ID,
		TotalValue:    100,
		PaymentStatus: saleDomain.PaymentApproved,
		PaymentType:   saleDomain.Pix,
		SaleDate:      time.Now(),
	}
	if err := db.Create(&otherSale).Error; err != nil {
		t.Fatalf("seed other sale: %v", err)
	}

	count, err := repository.CountSalesByCustomerIDs(1, []int{customerID, otherCustomer.ID})
	if err != nil {
		t.Fatalf("count sales: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one scoped sale, got %d", count)
	}
}

func newCustomerListTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.AutoMigrate(&customerEntity.CustomerEntity{}, &saleEntity.SaleEntity{}, &saleEntity.SaleItemEntity{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	return db
}

func seedCustomer(t *testing.T, db *gorm.DB, name string, phone string) int {
	t.Helper()

	customer := customerEntity.CustomerEntity{UserID: 1, Name: name, Phone: phone}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	return customer.ID
}

func seedSaleWithItems(t *testing.T, db *gorm.DB, customerID int, status saleDomain.PaymentStatus, paymentType saleDomain.PaymentType, saleDate time.Time, items []saleEntity.SaleItemEntity) {
	t.Helper()

	total := 0
	for i := range items {
		total += items[i].SalePrice * items[i].Quantity
	}

	sale := saleEntity.SaleEntity{
		UserID:        1,
		CustomerID:    customerID,
		TotalValue:    total,
		PaymentStatus: status,
		PaymentType:   paymentType,
		SaleDate:      saleDate,
		Items:         items,
	}
	if err := db.Create(&sale).Error; err != nil {
		t.Fatalf("seed sale: %v", err)
	}
}
