package response

import (
	"istore/internal/analytics/domain"
	saleDomain "istore/internal/sale/domain"
)

type DashboardMetricsResponse struct {
	Revenue            int     `json:"revenue"`
	Profit             int     `json:"profit"`
	ProfitMargin       float64 `json:"profitMargin"`
	ApprovedSalesCount int     `json:"approvedSalesCount"`
	AverageTicket      float64 `json:"averageTicket"`
	ItemsSold          int     `json:"itemsSold"`
	PendingSalesCount  int     `json:"pendingSalesCount"`
	CanceledSalesCount int     `json:"canceledSalesCount"`
}

func FromDashboardDomain(metrics *domain.DashboardMetrics) *DashboardMetricsResponse {
	if metrics == nil {
		return nil
	}

	return &DashboardMetricsResponse{
		Revenue:            metrics.Revenue,
		Profit:             metrics.Profit,
		ProfitMargin:       metrics.ProfitMargin,
		ApprovedSalesCount: metrics.ApprovedSalesCount,
		AverageTicket:      metrics.AverageTicket,
		ItemsSold:          metrics.ItemsSold,
		PendingSalesCount:  metrics.PendingSalesCount,
		CanceledSalesCount: metrics.CanceledSalesCount,
	}
}

type ProductMetricResponse struct {
	ProductName string `json:"productName"`
	Quantity    int    `json:"quantity"`
	Revenue     int    `json:"revenue"`
	Profit      int    `json:"profit"`
	SalesCount  int    `json:"salesCount"`
}

func FromProductDomain(metric *domain.ProductMetric) *ProductMetricResponse {
	if metric == nil {
		return nil
	}

	return &ProductMetricResponse{
		ProductName: metric.ProductName,
		Quantity:    metric.Quantity,
		Revenue:     metric.Revenue,
		Profit:      metric.Profit,
		SalesCount:  metric.SalesCount,
	}
}

func FromProductDomainSlice(metrics []domain.ProductMetric) []ProductMetricResponse {
	responses := make([]ProductMetricResponse, len(metrics))
	for i := range metrics {
		responses[i] = *FromProductDomain(&metrics[i])
	}

	return responses
}

type PaymentMetricResponse struct {
	PaymentType saleDomain.PaymentType `json:"paymentType"`
	SalesCount  int                    `json:"salesCount"`
	TotalValue  int                    `json:"totalValue"`
}

func FromPaymentDomain(metric *domain.PaymentMetric) *PaymentMetricResponse {
	if metric == nil {
		return nil
	}

	return &PaymentMetricResponse{
		PaymentType: metric.PaymentType,
		SalesCount:  metric.SalesCount,
		TotalValue:  metric.TotalValue,
	}
}

func FromPaymentDomainSlice(metrics []domain.PaymentMetric) []PaymentMetricResponse {
	responses := make([]PaymentMetricResponse, len(metrics))
	for i := range metrics {
		responses[i] = *FromPaymentDomain(&metrics[i])
	}

	return responses
}

type FinancialMetricResponse struct {
	Period  string `json:"period"`
	Revenue int    `json:"revenue,omitempty"`
	Profit  int    `json:"profit,omitempty"`
}

func FromFinancialDomain(metric *domain.FinancialMetric) *FinancialMetricResponse {
	if metric == nil {
		return nil
	}

	return &FinancialMetricResponse{
		Period:  metric.Period,
		Revenue: metric.Revenue,
		Profit:  metric.Profit,
	}
}

func FromFinancialDomainSlice(metrics []domain.FinancialMetric) []FinancialMetricResponse {
	responses := make([]FinancialMetricResponse, len(metrics))
	for i := range metrics {
		responses[i] = *FromFinancialDomain(&metrics[i])
	}

	return responses
}

type CustomerMetricResponse struct {
	CustomerID   int    `json:"customerId"`
	CustomerName string `json:"customerName"`
	SalesCount   int    `json:"salesCount"`
	Revenue      int    `json:"revenue"`
	Profit       int    `json:"profit"`
}

func FromCustomerDomain(metric *domain.CustomerMetric) *CustomerMetricResponse {
	if metric == nil {
		return nil
	}

	return &CustomerMetricResponse{
		CustomerID:   metric.CustomerID,
		CustomerName: metric.CustomerName,
		SalesCount:   metric.SalesCount,
		Revenue:      metric.Revenue,
		Profit:       metric.Profit,
	}
}

func FromCustomerDomainSlice(metrics []domain.CustomerMetric) []CustomerMetricResponse {
	responses := make([]CustomerMetricResponse, len(metrics))
	for i := range metrics {
		responses[i] = *FromCustomerDomain(&metrics[i])
	}

	return responses
}

type StatusMetricResponse struct {
	Status     saleDomain.PaymentStatus `json:"status"`
	SalesCount int                      `json:"salesCount"`
	TotalValue int                      `json:"totalValue"`
}

func FromStatusDomain(metric *domain.StatusMetric) *StatusMetricResponse {
	if metric == nil {
		return nil
	}

	return &StatusMetricResponse{
		Status:     metric.Status,
		SalesCount: metric.SalesCount,
		TotalValue: metric.TotalValue,
	}
}

func FromStatusDomainSlice(metrics []domain.StatusMetric) []StatusMetricResponse {
	responses := make([]StatusMetricResponse, len(metrics))
	for i := range metrics {
		responses[i] = *FromStatusDomain(&metrics[i])
	}

	return responses
}
