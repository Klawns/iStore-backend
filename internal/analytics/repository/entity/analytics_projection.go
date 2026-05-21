package entity

import (
	"istore/internal/analytics/domain"
	saleDomain "istore/internal/sale/domain"
)

type DashboardTotalsProjection struct {
	Revenue            int `gorm:"column:revenue"`
	ApprovedSalesCount int `gorm:"column:approved_sales_count"`
}

type DashboardItemTotalsProjection struct {
	Profit    int `gorm:"column:profit"`
	ItemsSold int `gorm:"column:items_sold"`
}

func (p *DashboardTotalsProjection) ToDomain(itemTotals *DashboardItemTotalsProjection, pendingSalesCount int, canceledSalesCount int) *domain.DashboardMetrics {
	if p == nil {
		return nil
	}
	if itemTotals == nil {
		itemTotals = &DashboardItemTotalsProjection{}
	}

	var averageTicket float64
	if p.ApprovedSalesCount > 0 {
		averageTicket = float64(p.Revenue) / float64(p.ApprovedSalesCount)
	}

	var profitMargin float64
	if p.Revenue > 0 {
		profitMargin = (float64(itemTotals.Profit) / float64(p.Revenue)) * 100
	}

	return &domain.DashboardMetrics{
		Revenue:            p.Revenue,
		Profit:             itemTotals.Profit,
		ProfitMargin:       profitMargin,
		ApprovedSalesCount: p.ApprovedSalesCount,
		AverageTicket:      averageTicket,
		ItemsSold:          itemTotals.ItemsSold,
		PendingSalesCount:  pendingSalesCount,
		CanceledSalesCount: canceledSalesCount,
	}
}

type FinancialMetricProjection struct {
	Period  string `gorm:"column:period"`
	Revenue int    `gorm:"column:revenue"`
	Profit  int    `gorm:"column:profit"`
}

func (p *FinancialMetricProjection) ToDomain() *domain.FinancialMetric {
	if p == nil {
		return nil
	}

	return &domain.FinancialMetric{
		Period:  p.Period,
		Revenue: p.Revenue,
		Profit:  p.Profit,
	}
}

func FinancialMetricProjectionsToDomain(projections []FinancialMetricProjection) []domain.FinancialMetric {
	metrics := make([]domain.FinancialMetric, len(projections))
	for i := range projections {
		metrics[i] = *projections[i].ToDomain()
	}

	return metrics
}

type ProductMetricProjection struct {
	ProductName string `gorm:"column:product_name"`
	Quantity    int    `gorm:"column:quantity"`
	Revenue     int    `gorm:"column:revenue"`
	Profit      int    `gorm:"column:profit"`
	SalesCount  int    `gorm:"column:sales_count"`
}

func (p *ProductMetricProjection) ToDomain() *domain.ProductMetric {
	if p == nil {
		return nil
	}

	return &domain.ProductMetric{
		ProductName: p.ProductName,
		Quantity:    p.Quantity,
		Revenue:     p.Revenue,
		Profit:      p.Profit,
		SalesCount:  p.SalesCount,
	}
}

func ProductMetricProjectionsToDomain(projections []ProductMetricProjection) []domain.ProductMetric {
	metrics := make([]domain.ProductMetric, len(projections))
	for i := range projections {
		metrics[i] = *projections[i].ToDomain()
	}

	return metrics
}

type PaymentMetricProjection struct {
	PaymentType saleDomain.PaymentType `gorm:"column:payment_type"`
	SalesCount  int                    `gorm:"column:sales_count"`
	TotalValue  int                    `gorm:"column:total_value"`
}

func (p *PaymentMetricProjection) ToDomain() *domain.PaymentMetric {
	if p == nil {
		return nil
	}

	return &domain.PaymentMetric{
		PaymentType: p.PaymentType,
		SalesCount:  p.SalesCount,
		TotalValue:  p.TotalValue,
	}
}

func PaymentMetricProjectionsToDomain(projections []PaymentMetricProjection) []domain.PaymentMetric {
	metrics := make([]domain.PaymentMetric, len(projections))
	for i := range projections {
		metrics[i] = *projections[i].ToDomain()
	}

	return metrics
}

type CustomerMetricProjection struct {
	CustomerID   int    `gorm:"column:customer_id"`
	CustomerName string `gorm:"column:customer_name"`
	SalesCount   int    `gorm:"column:sales_count"`
	Revenue      int    `gorm:"column:revenue"`
	Profit       int    `gorm:"column:profit"`
}

func (p *CustomerMetricProjection) ToDomain() *domain.CustomerMetric {
	if p == nil {
		return nil
	}

	return &domain.CustomerMetric{
		CustomerID:   p.CustomerID,
		CustomerName: p.CustomerName,
		SalesCount:   p.SalesCount,
		Revenue:      p.Revenue,
		Profit:       p.Profit,
	}
}

func CustomerMetricProjectionsToDomain(projections []CustomerMetricProjection) []domain.CustomerMetric {
	metrics := make([]domain.CustomerMetric, len(projections))
	for i := range projections {
		metrics[i] = *projections[i].ToDomain()
	}

	return metrics
}

type StatusMetricProjection struct {
	Status     saleDomain.PaymentStatus `gorm:"column:status"`
	SalesCount int                      `gorm:"column:sales_count"`
	TotalValue int                      `gorm:"column:total_value"`
}

func (p *StatusMetricProjection) ToDomain() *domain.StatusMetric {
	if p == nil {
		return nil
	}

	return &domain.StatusMetric{
		Status:     p.Status,
		SalesCount: p.SalesCount,
		TotalValue: p.TotalValue,
	}
}

func StatusMetricProjectionsToDomain(projections []StatusMetricProjection) []domain.StatusMetric {
	metrics := make([]domain.StatusMetric, len(projections))
	for i := range projections {
		metrics[i] = *projections[i].ToDomain()
	}

	return metrics
}
