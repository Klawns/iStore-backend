package implementation

import (
	"errors"
	"istore/internal/customer/domain"
	"istore/internal/customer/repository/contracts"
	"istore/internal/customer/repository/entity"
	"strings"

	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) contracts.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(customer *domain.Customer) error {
	// Primeiro, convertemos o domínio para a entidade do banco de dados.
	customerEntity := entity.FromDomain(customer)

	// Segundo, salvamos a entidade usando o GORM e retornamos o erro bruto, se houver.
	if err := r.db.Create(customerEntity).Error; err != nil {
		return err
	}

	// Terceiro, devolvemos o ID gerado para o domínio recebido.
	if customerEntity != nil {
		customer.ID = customerEntity.ID
	}

	return nil
}

func (r *customerRepository) Update(customer *domain.Customer) error {
	// Primeiro, convertemos o domínio já validado pelo service para entidade.
	customerEntity := entity.FromDomain(customer)

	// Segundo, salvamos a entidade no banco e retornamos apenas o erro do GORM.
	return r.db.Save(customerEntity).Error
}

func (r *customerRepository) Delete(userID uint, id int) error {
	// Deletamos pelo ID e retornamos apenas o erro do GORM.
	return r.db.Where("user_id = ? AND id = ?", userID, id).Delete(&entity.CustomerEntity{}).Error
}

func (r *customerRepository) DeleteMany(userID uint, ids []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Where("user_id = ? AND id IN ?", userID, ids).Delete(&entity.CustomerEntity{}).Error
	})
}

func (r *customerRepository) FindByID(userID uint, id int) (*domain.Customer, error) {
	// Primeiro, buscamos o cliente no banco usando o ID recebido.
	var customerEntity entity.CustomerEntity
	if err := r.db.Where("user_id = ? AND id = ?", userID, id).First(&customerEntity).Error; err != nil {
		// Quando não encontrar registro, retornamos nil sem erro para o service decidir o 404.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	// Segundo, convertemos a entidade encontrada para domínio.
	return customerEntity.ToDomain(), nil
}

func (r *customerRepository) FindByIDs(userID uint, ids []int) ([]domain.Customer, error) {
	var customerEntities []entity.CustomerEntity
	if err := r.db.Where("user_id = ? AND id IN ?", userID, ids).Find(&customerEntities).Error; err != nil {
		return nil, err
	}

	customers := make([]domain.Customer, len(customerEntities))
	for i, customerEntity := range customerEntities {
		customers[i] = *customerEntity.ToDomain()
	}

	return customers, nil
}

func (r *customerRepository) FindAll() ([]domain.Customer, error) {
	// Primeiro, buscamos todos os clientes cadastrados no banco.
	var customerEntities []entity.CustomerEntity
	if err := r.db.Find(&customerEntities).Error; err != nil {
		return nil, err
	}

	// Segundo, convertemos a lista de entidades para domínios.
	customers := make([]domain.Customer, len(customerEntities))
	for i, customerEntity := range customerEntities {
		customers[i] = *customerEntity.ToDomain()
	}

	return customers, nil
}

func (r *customerRepository) List(filter domain.CustomerListFilter) (*domain.CustomerListResult, error) {
	metrics := r.customerMetricsQuery(filter)
	base := r.filteredCustomersQuery(filter, metrics)

	var totalItems int64
	if err := r.db.Table("(?) AS filtered_customers", base).Count(&totalItems).Error; err != nil {
		return nil, err
	}

	summary, err := r.customerListSummary(base, totalItems)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int((totalItems + int64(filter.Limit) - 1) / int64(filter.Limit))
	}

	var items []customerListProjection
	if totalItems > 0 {
		offset := (filter.Page - 1) * filter.Limit
		if err := r.db.Table("(?) AS filtered_customers", base).
			Select(`
				id,
				name,
				phone,
				sales_count,
				revenue,
				profit,
				CASE WHEN sales_count > 0 THEN revenue / sales_count ELSE 0 END AS average_ticket`).
			Order("name ASC, id ASC").
			Limit(filter.Limit).
			Offset(offset).
			Scan(&items).Error; err != nil {
			return nil, err
		}
	}

	resultItems := make([]domain.CustomerListItem, len(items))
	for i := range items {
		resultItems[i] = items[i].ToDomain()
	}

	return &domain.CustomerListResult{
		Items:      resultItems,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
		Summary:    summary,
	}, nil
}

func (r *customerRepository) CountSalesByCustomerIDs(userID uint, ids []int) (int64, error) {
	var count int64
	if err := r.db.Table("sales").
		Where("user_id = ? AND customer_id IN ?", userID, ids).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *customerRepository) filteredCustomersQuery(filter domain.CustomerListFilter, metrics *gorm.DB) *gorm.DB {
	joinType := "LEFT JOIN"
	if hasSaleFilters(filter) {
		joinType = "JOIN"
	}

	query := r.db.Table("customers").
		Select(`
			customers.id,
			customers.name,
			customers.phone,
			COALESCE(customer_metrics.sales_count, 0) AS sales_count,
			COALESCE(customer_metrics.revenue, 0) AS revenue,
			COALESCE(customer_metrics.profit, 0) AS profit`).
		Joins(joinType+" (?) AS customer_metrics ON customer_metrics.customer_id = customers.id", metrics).
		Where("customers.user_id = ?", filter.UserID)

	if filter.Search != "" {
		term := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(customers.name) LIKE ? OR LOWER(customers.phone) LIKE ?", term, term)
	}

	return query
}

func (r *customerRepository) customerMetricsQuery(filter domain.CustomerListFilter) *gorm.DB {
	profitBySale := r.db.Table("sale_items").
		Select("sale_items.sale_id, COALESCE(SUM((sale_items.sale_price - sale_items.cost_price) * sale_items.quantity), 0) AS profit").
		Group("sale_items.sale_id")

	query := r.db.Table("sales").
		Select(`
			sales.customer_id AS customer_id,
			COUNT(sales.id) AS sales_count,
			COALESCE(SUM(sales.total_value), 0) AS revenue,
			COALESCE(SUM(sale_profit.profit), 0) AS profit`).
		Joins("LEFT JOIN (?) AS sale_profit ON sale_profit.sale_id = sales.id", profitBySale).
		Where("sales.user_id = ?", filter.UserID).
		Group("sales.customer_id")

	if filter.Start != nil {
		query = query.Where("sales.sale_date >= ?", *filter.Start)
	}

	if filter.End != nil {
		query = query.Where("sales.sale_date <= ?", *filter.End)
	}

	if filter.PaymentStatus != nil {
		query = query.Where("sales.payment_status = ?", *filter.PaymentStatus)
	}

	if filter.PaymentType != nil {
		query = query.Where("sales.payment_type = ?", *filter.PaymentType)
	}

	return query
}

func (r *customerRepository) customerListSummary(filteredCustomers *gorm.DB, totalItems int64) (domain.CustomerListSummary, error) {
	var projection customerSummaryProjection
	if err := r.db.Table("(?) AS filtered_customers", filteredCustomers).
		Select(`
			COALESCE(SUM(sales_count), 0) AS sales_count,
			COALESCE(SUM(revenue), 0) AS revenue,
			COALESCE(SUM(profit), 0) AS profit,
			COALESCE(SUM(CASE WHEN sales_count > 1 THEN 1 ELSE 0 END), 0) AS repeat_customers`).
		Scan(&projection).Error; err != nil {
		return domain.CustomerListSummary{}, err
	}

	averageTicket := 0
	if projection.SalesCount > 0 {
		averageTicket = projection.Revenue / projection.SalesCount
	}

	repeatRate := 0
	if totalItems > 0 {
		repeatRate = int((int64(projection.RepeatCustomers) * 100) / totalItems)
	}

	return domain.CustomerListSummary{
		TotalCustomers: int(totalItems),
		SalesCount:     projection.SalesCount,
		Revenue:        projection.Revenue,
		Profit:         projection.Profit,
		AverageTicket:  averageTicket,
		RepeatRate:     repeatRate,
	}, nil
}

func hasSaleFilters(filter domain.CustomerListFilter) bool {
	return filter.Start != nil || filter.End != nil || filter.PaymentStatus != nil || filter.PaymentType != nil
}

type customerListProjection struct {
	ID            int    `gorm:"column:id"`
	Name          string `gorm:"column:name"`
	Phone         string `gorm:"column:phone"`
	SalesCount    int    `gorm:"column:sales_count"`
	Revenue       int    `gorm:"column:revenue"`
	Profit        int    `gorm:"column:profit"`
	AverageTicket int    `gorm:"column:average_ticket"`
}

func (p customerListProjection) ToDomain() domain.CustomerListItem {
	return domain.CustomerListItem{
		ID:            p.ID,
		Name:          p.Name,
		Phone:         p.Phone,
		SalesCount:    p.SalesCount,
		Revenue:       p.Revenue,
		Profit:        p.Profit,
		AverageTicket: p.AverageTicket,
	}
}

type customerSummaryProjection struct {
	SalesCount      int `gorm:"column:sales_count"`
	Revenue         int `gorm:"column:revenue"`
	Profit          int `gorm:"column:profit"`
	RepeatCustomers int `gorm:"column:repeat_customers"`
}
