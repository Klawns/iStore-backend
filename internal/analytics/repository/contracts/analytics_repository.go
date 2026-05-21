package contracts

import "istore/internal/analytics/domain"

type AnalyticsRepository interface {
	GetDashboardMetrics(filter domain.AnalyticsFilter) (*domain.DashboardMetrics, error)

	GetRevenue(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, error)

	GetProfit(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, error)

	GetTopProducts(filter domain.AnalyticsFilter) ([]domain.ProductMetric, error)

	GetPaymentMethods(filter domain.AnalyticsFilter) ([]domain.PaymentMetric, error)

	GetTopCustomers(filter domain.AnalyticsFilter) ([]domain.CustomerMetric, error)

	GetStatuses(filter domain.AnalyticsFilter) ([]domain.StatusMetric, error)
}
