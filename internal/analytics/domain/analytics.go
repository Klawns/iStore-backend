package domain

import (
	saleDomain "istore/internal/sale/domain"
	"time"
)

const (
	GroupByDaily   = "daily"
	GroupByMonthly = "monthly"
)

type AnalyticsFilter struct {
	UserID      uint
	StartDate   time.Time
	EndDate     time.Time
	Limit       int
	Status      saleDomain.PaymentStatus
	PaymentType saleDomain.PaymentType
	GroupBy     string
}

type DashboardMetrics struct {
	Revenue            int
	Profit             int
	ProfitMargin       float64
	ApprovedSalesCount int
	AverageTicket      float64
	ItemsSold          int
	PendingSalesCount  int
	CanceledSalesCount int
}

type ProductMetric struct {
	ProductName string
	Quantity    int
	Revenue     int
	Profit      int
	SalesCount  int
}

type PaymentMetric struct {
	PaymentType saleDomain.PaymentType
	SalesCount  int
	TotalValue  int
}

type FinancialMetric struct {
	Period  string
	Revenue int
	Profit  int
}

type CustomerMetric struct {
	CustomerID   int
	CustomerName string
	SalesCount   int
	Revenue      int
	Profit       int
}

type StatusMetric struct {
	Status     saleDomain.PaymentStatus
	SalesCount int
	TotalValue int
}
