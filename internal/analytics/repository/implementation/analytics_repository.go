package implementation

import (
	"istore/internal/analytics/domain"
	"istore/internal/analytics/repository/contracts"
	"istore/internal/analytics/repository/entity"
	saleDomain "istore/internal/sale/domain"

	"gorm.io/gorm"
)

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) contracts.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetDashboardMetrics(filter domain.AnalyticsFilter) (*domain.DashboardMetrics, error) {
	status := statusOrDefault(filter.Status)

	var totals entity.DashboardTotalsProjection
	query := applyPaymentTypeFilter(applyPeriodFilters(r.db.Table("sales"), filter), filter).
		Where("payment_status = ?", status).
		Select("COALESCE(SUM(total_value), 0) AS revenue, COUNT(*) AS approved_sales_count")
	if err := query.Scan(&totals).Error; err != nil {
		return nil, err
	}

	var itemTotals entity.DashboardItemTotalsProjection
	itemQuery := applyPaymentTypeFilter(applyPeriodFilters(r.db.Table("sales"), filter), filter).
		Joins("JOIN sale_items ON sale_items.sale_id = sales.id").
		Where("sales.payment_status = ?", status).
		Select("COALESCE(SUM((sale_items.sale_price - sale_items.cost_price) * sale_items.quantity), 0) AS profit, COALESCE(SUM(sale_items.quantity), 0) AS items_sold")
	if err := itemQuery.Scan(&itemTotals).Error; err != nil {
		return nil, err
	}

	pendingCount, err := r.countSalesByStatus(filter, saleDomain.PaymentPending)
	if err != nil {
		return nil, err
	}

	canceledCount, err := r.countSalesByStatus(filter, saleDomain.PaymentCanceled)
	if err != nil {
		return nil, err
	}

	return totals.ToDomain(&itemTotals, pendingCount, canceledCount), nil
}

func (r *analyticsRepository) GetRevenue(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, error) {
	var metrics []entity.FinancialMetricProjection
	query := applyAnalyticsSaleFilters(r.db.Table("sales"), filter).
		Select(periodExpression(filter.GroupBy, r.db.Dialector.Name()) + " AS period, COALESCE(SUM(total_value), 0) AS revenue").
		Group("period").
		Order("period ASC")

	if err := query.Scan(&metrics).Error; err != nil {
		return nil, err
	}

	return entity.FinancialMetricProjectionsToDomain(metrics), nil
}

func (r *analyticsRepository) GetProfit(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, error) {
	var metrics []entity.FinancialMetricProjection
	query := applyAnalyticsSaleFilters(r.db.Table("sales"), filter).
		Joins("JOIN sale_items ON sale_items.sale_id = sales.id").
		Select(periodExpression(filter.GroupBy, r.db.Dialector.Name()) + " AS period, COALESCE(SUM((sale_items.sale_price - sale_items.cost_price) * sale_items.quantity), 0) AS profit").
		Group("period").
		Order("period ASC")

	if err := query.Scan(&metrics).Error; err != nil {
		return nil, err
	}

	return entity.FinancialMetricProjectionsToDomain(metrics), nil
}

func (r *analyticsRepository) GetTopProducts(filter domain.AnalyticsFilter) ([]domain.ProductMetric, error) {
	var metrics []entity.ProductMetricProjection
	query := applyAnalyticsSaleFilters(r.db.Table("sales"), filter).
		Joins("JOIN sale_items ON sale_items.sale_id = sales.id").
		Select(`
			sale_items.product_name AS product_name,
			COALESCE(SUM(sale_items.quantity), 0) AS quantity,
			COALESCE(SUM(sale_items.sale_price * sale_items.quantity), 0) AS revenue,
			COALESCE(SUM((sale_items.sale_price - sale_items.cost_price) * sale_items.quantity), 0) AS profit,
			COUNT(DISTINCT sales.id) AS sales_count`).
		Group("sale_items.product_name").
		Order("quantity DESC, revenue DESC, product_name ASC")
	query = applyLimit(query, filter)

	if err := query.Scan(&metrics).Error; err != nil {
		return nil, err
	}

	return entity.ProductMetricProjectionsToDomain(metrics), nil
}

func (r *analyticsRepository) GetPaymentMethods(filter domain.AnalyticsFilter) ([]domain.PaymentMetric, error) {
	var metrics []entity.PaymentMetricProjection
	query := applyAnalyticsSaleFilters(r.db.Table("sales"), filter).
		Select("payment_type AS payment_type, COUNT(*) AS sales_count, COALESCE(SUM(total_value), 0) AS total_value").
		Group("payment_type").
		Order("total_value DESC, sales_count DESC, payment_type ASC")

	if err := query.Scan(&metrics).Error; err != nil {
		return nil, err
	}

	return entity.PaymentMetricProjectionsToDomain(metrics), nil
}

func (r *analyticsRepository) GetTopCustomers(filter domain.AnalyticsFilter) ([]domain.CustomerMetric, error) {
	var metrics []entity.CustomerMetricProjection
	profitBySale := r.db.Table("sale_items").
		Select("sale_id, COALESCE(SUM((sale_price - cost_price) * quantity), 0) AS profit").
		Group("sale_id")

	query := applyAnalyticsSaleFilters(r.db.Table("sales"), filter).
		Joins("LEFT JOIN customers ON customers.id = sales.customer_id").
		Joins("LEFT JOIN (?) AS sale_profit ON sale_profit.sale_id = sales.id", profitBySale).
		Select(`
			sales.customer_id AS customer_id,
			COALESCE(customers.name, '') AS customer_name,
			COUNT(sales.id) AS sales_count,
			COALESCE(SUM(sales.total_value), 0) AS revenue,
			COALESCE(SUM(sale_profit.profit), 0) AS profit`).
		Group("sales.customer_id, customers.name").
		Order("revenue DESC, sales_count DESC, customer_name ASC")
	query = applyLimit(query, filter)

	if err := query.Scan(&metrics).Error; err != nil {
		return nil, err
	}

	return entity.CustomerMetricProjectionsToDomain(metrics), nil
}

func (r *analyticsRepository) GetStatuses(filter domain.AnalyticsFilter) ([]domain.StatusMetric, error) {
	var metrics []entity.StatusMetricProjection
	query := applyPaymentTypeFilter(applyPeriodFilters(r.db.Table("sales"), filter), filter).
		Select("payment_status AS status, COUNT(*) AS sales_count, COALESCE(SUM(total_value), 0) AS total_value").
		Group("payment_status").
		Order("payment_status ASC")

	if err := query.Scan(&metrics).Error; err != nil {
		return nil, err
	}

	return entity.StatusMetricProjectionsToDomain(metrics), nil
}

func (r *analyticsRepository) countSalesByStatus(filter domain.AnalyticsFilter, status saleDomain.PaymentStatus) (int, error) {
	var count int64
	query := applyPaymentTypeFilter(applyPeriodFilters(r.db.Model(nil).Table("sales"), filter), filter).
		Where("payment_status = ?", status)
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func applyPeriodFilters(query *gorm.DB, filter domain.AnalyticsFilter) *gorm.DB {
	query = query.Where("sales.user_id = ?", filter.UserID)
	if !filter.StartDate.IsZero() {
		query = query.Where("sales.sale_date >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("sales.sale_date <= ?", filter.EndDate)
	}

	return query
}

func applyAnalyticsSaleFilters(query *gorm.DB, filter domain.AnalyticsFilter) *gorm.DB {
	return applyPaymentTypeFilter(applyStatusFilter(applyPeriodFilters(query, filter), filter), filter)
}

func applyStatusFilter(query *gorm.DB, filter domain.AnalyticsFilter) *gorm.DB {
	return query.Where("sales.payment_status = ?", statusOrDefault(filter.Status))
}

func applyPaymentTypeFilter(query *gorm.DB, filter domain.AnalyticsFilter) *gorm.DB {
	if filter.PaymentType == "" {
		return query
	}

	return query.Where("sales.payment_type = ?", filter.PaymentType)
}

func applyLimit(query *gorm.DB, filter domain.AnalyticsFilter) *gorm.DB {
	if filter.Limit > 0 {
		return query.Limit(filter.Limit)
	}

	return query
}

func statusOrDefault(status saleDomain.PaymentStatus) saleDomain.PaymentStatus {
	if status == "" {
		return saleDomain.PaymentApproved
	}

	return status
}

func periodExpression(groupBy string, dialect string) string {
	if dialect == "sqlite" {
		if groupBy == domain.GroupByMonthly {
			return "strftime('%Y-%m', sales.sale_date)"
		}

		return "strftime('%Y-%m-%d', sales.sale_date)"
	}

	if groupBy == domain.GroupByMonthly {
		return "to_char(sales.sale_date, 'YYYY-MM')"
	}

	return "to_char(sales.sale_date, 'YYYY-MM-DD')"
}
