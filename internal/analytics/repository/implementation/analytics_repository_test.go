package implementation

import (
	"istore/internal/analytics/domain"
	customerEntity "istore/internal/customer/repository/entity"
	saleDomain "istore/internal/sale/domain"
	saleEntity "istore/internal/sale/repository/entity"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAnalyticsRepositoryFiltersByPeriodStatusAndPaymentType(t *testing.T) {
	db := newAnalyticsTestDB(t)
	repository := NewAnalyticsRepository(db)
	anaID := seedAnalyticsCustomer(t, db, "Ana Silva")
	brunoID := seedAnalyticsCustomer(t, db, "Bruno Costa")

	seedAnalyticsSale(t, db, anaID, saleDomain.PaymentApproved, saleDomain.Pix, time.Date(2026, time.May, 5, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "iPhone", Quantity: 1, CostPrice: 6000, SalePrice: 10000},
	})
	seedAnalyticsSale(t, db, anaID, saleDomain.PaymentPending, saleDomain.Pix, time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "MacBook", Quantity: 1, CostPrice: 20000, SalePrice: 30000},
	})
	seedAnalyticsSale(t, db, brunoID, saleDomain.PaymentApproved, saleDomain.Money, time.Date(2026, time.May, 7, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "AirPods", Quantity: 2, CostPrice: 1000, SalePrice: 2500},
	})
	seedAnalyticsSale(t, db, brunoID, saleDomain.PaymentApproved, saleDomain.Pix, time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "Case", Quantity: 1, CostPrice: 500, SalePrice: 1000},
	})

	filter := domain.AnalyticsFilter{
		StartDate:   time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, time.May, 31, 23, 59, 59, 0, time.UTC),
		Status:      saleDomain.PaymentApproved,
		PaymentType: saleDomain.Pix,
		GroupBy:     domain.GroupByMonthly,
		Limit:       5,
	}

	dashboard, err := repository.GetDashboardMetrics(filter)
	if err != nil {
		t.Fatalf("dashboard metrics: %v", err)
	}
	if dashboard.Revenue != 10000 || dashboard.Profit != 4000 || dashboard.ApprovedSalesCount != 1 || dashboard.PendingSalesCount != 1 {
		t.Fatalf("unexpected dashboard metrics: %#v", dashboard)
	}

	revenue, err := repository.GetRevenue(filter)
	if err != nil {
		t.Fatalf("revenue metrics: %v", err)
	}
	if len(revenue) != 1 || revenue[0].Period != "2026-05" || revenue[0].Revenue != 10000 {
		t.Fatalf("unexpected revenue metrics: %#v", revenue)
	}

	profit, err := repository.GetProfit(filter)
	if err != nil {
		t.Fatalf("profit metrics: %v", err)
	}
	if len(profit) != 1 || profit[0].Period != "2026-05" || profit[0].Profit != 4000 {
		t.Fatalf("unexpected profit metrics: %#v", profit)
	}

	products, err := repository.GetTopProducts(filter)
	if err != nil {
		t.Fatalf("top products: %v", err)
	}
	if len(products) != 1 || products[0].ProductName != "iPhone" || products[0].Revenue != 10000 {
		t.Fatalf("unexpected product metrics: %#v", products)
	}

	payments, err := repository.GetPaymentMethods(filter)
	if err != nil {
		t.Fatalf("payment methods: %v", err)
	}
	if len(payments) != 1 || payments[0].PaymentType != saleDomain.Pix || payments[0].TotalValue != 10000 {
		t.Fatalf("unexpected payment metrics: %#v", payments)
	}

	customers, err := repository.GetTopCustomers(filter)
	if err != nil {
		t.Fatalf("top customers: %v", err)
	}
	if len(customers) != 1 || customers[0].CustomerName != "Ana Silva" || customers[0].Revenue != 10000 {
		t.Fatalf("unexpected customer metrics: %#v", customers)
	}

	statuses, err := repository.GetStatuses(filter)
	if err != nil {
		t.Fatalf("statuses: %v", err)
	}
	if len(statuses) != 2 || statuses[0].Status != saleDomain.PaymentApproved || statuses[0].TotalValue != 10000 || statuses[1].Status != saleDomain.PaymentPending || statuses[1].TotalValue != 30000 {
		t.Fatalf("unexpected status metrics: %#v", statuses)
	}
}

func TestAnalyticsRepositoryDefaultsToApprovedStatus(t *testing.T) {
	db := newAnalyticsTestDB(t)
	repository := NewAnalyticsRepository(db)
	customerID := seedAnalyticsCustomer(t, db, "Ana Silva")

	seedAnalyticsSale(t, db, customerID, saleDomain.PaymentApproved, saleDomain.Pix, time.Date(2026, time.May, 5, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "iPhone", Quantity: 1, CostPrice: 6000, SalePrice: 10000},
	})
	seedAnalyticsSale(t, db, customerID, saleDomain.PaymentPending, saleDomain.Pix, time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC), []saleEntity.SaleItemEntity{
		{ProductName: "MacBook", Quantity: 1, CostPrice: 20000, SalePrice: 30000},
	})

	revenue, err := repository.GetRevenue(domain.AnalyticsFilter{GroupBy: domain.GroupByMonthly})
	if err != nil {
		t.Fatalf("revenue metrics: %v", err)
	}

	if len(revenue) != 1 || revenue[0].Revenue != 10000 {
		t.Fatalf("expected only approved revenue by default, got %#v", revenue)
	}
}

func TestPeriodExpressionUsesPostgresDateFormatting(t *testing.T) {
	if got := periodExpression(domain.GroupByDaily, "postgres"); got != "to_char(sales.sale_date, 'YYYY-MM-DD')" {
		t.Fatalf("unexpected daily expression: %s", got)
	}

	if got := periodExpression(domain.GroupByMonthly, "postgres"); got != "to_char(sales.sale_date, 'YYYY-MM')" {
		t.Fatalf("unexpected monthly expression: %s", got)
	}
}

func newAnalyticsTestDB(t *testing.T) *gorm.DB {
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

func seedAnalyticsCustomer(t *testing.T, db *gorm.DB, name string) int {
	t.Helper()

	customer := customerEntity.CustomerEntity{Name: name}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	return customer.ID
}

func seedAnalyticsSale(t *testing.T, db *gorm.DB, customerID int, status saleDomain.PaymentStatus, paymentType saleDomain.PaymentType, saleDate time.Time, items []saleEntity.SaleItemEntity) {
	t.Helper()

	total := 0
	for i := range items {
		total += items[i].SalePrice * items[i].Quantity
	}

	sale := saleEntity.SaleEntity{
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
