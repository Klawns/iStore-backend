package contracts

import (
	"istore/internal/analytics/domain"
	"istore/pkg/rest_err"
)

type AnalyticsService interface {
	GetDashboard(filter domain.AnalyticsFilter) (*domain.DashboardMetrics, *rest_err.RestErr)

	GetRevenue(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, *rest_err.RestErr)

	GetProfit(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, *rest_err.RestErr)

	GetTopProducts(filter domain.AnalyticsFilter) ([]domain.ProductMetric, *rest_err.RestErr)

	GetPaymentMethods(filter domain.AnalyticsFilter) ([]domain.PaymentMetric, *rest_err.RestErr)

	GetTopCustomers(filter domain.AnalyticsFilter) ([]domain.CustomerMetric, *rest_err.RestErr)

	GetStatuses(filter domain.AnalyticsFilter) ([]domain.StatusMetric, *rest_err.RestErr)
}
