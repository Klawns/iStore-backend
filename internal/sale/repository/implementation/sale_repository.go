package implementation

import (
	"errors"
	"istore/internal/sale/domain"
	"istore/internal/sale/repository/contracts"
	"istore/internal/sale/repository/entity"
	"strings"
	"time"

	"gorm.io/gorm"
)

type saleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) contracts.SaleRepository {
	return &saleRepository{db: db}
}

func (r *saleRepository) Create(sale *domain.Sale) error {
	saleEntity := entity.FromSaleDomain(sale)

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(saleEntity).Error; err != nil {
			return err
		}

		for i := range sale.InstallmentsList {
			sale.InstallmentsList[i].SaleID = saleEntity.ID
			installmentEntity := entity.FromSaleInstallmentDomain(&sale.InstallmentsList[i])
			if err := tx.Create(installmentEntity).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	var createdSale entity.SaleEntity
	if err := r.db.Preload("Customer").Preload("Items").First(&createdSale, saleEntity.ID).Error; err != nil {
		return err
	}
	*sale = *createdSale.ToDomain()

	return nil
}

func (r *saleRepository) FindByID(id int) (*domain.Sale, error) {
	var saleEntity entity.SaleEntity
	if err := r.db.Preload("Customer").Preload("Items").First(&saleEntity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return saleEntity.ToDomain(), nil
}

func (r *saleRepository) FindAll() ([]domain.Sale, error) {
	var saleEntities []entity.SaleEntity
	if err := r.db.Preload("Customer").Preload("Items").Find(&saleEntities).Error; err != nil {
		return nil, err
	}

	sales := make([]domain.Sale, len(saleEntities))
	for i, saleEntity := range saleEntities {
		sales[i] = *saleEntity.ToDomain()
	}

	return sales, nil
}

func (r *saleRepository) List(filter domain.SaleListFilter) (*domain.SaleListResult, error) {
	baseQuery := r.filteredSaleIDs(filter)

	var totalItems int64
	if err := r.db.Table("(?) AS filtered_sales", baseQuery).
		Count(&totalItems).Error; err != nil {
		return nil, err
	}

	summary, err := r.saleListSummary(baseQuery, totalItems)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int((totalItems + int64(filter.Limit) - 1) / int64(filter.Limit))
	}

	var saleEntities []entity.SaleEntity
	if totalItems > 0 {
		offset := (filter.Page - 1) * filter.Limit
		pageIDs := r.db.Table("(?) AS filtered_sales", baseQuery).
			Select("filtered_sales.id").
			Order("filtered_sales.sale_date DESC, filtered_sales.id DESC").
			Limit(filter.Limit).
			Offset(offset)

		if err := r.db.Preload("Customer").Preload("Items").
			Where("sales.id IN (?)", pageIDs).
			Order("sales.sale_date DESC, sales.id DESC").
			Find(&saleEntities).Error; err != nil {
			return nil, err
		}
	}

	sales := make([]domain.Sale, len(saleEntities))
	for i, saleEntity := range saleEntities {
		sales[i] = *saleEntity.ToDomain()
	}

	return &domain.SaleListResult{
		Items:      sales,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
		Summary:    summary,
	}, nil
}

func (r *saleRepository) ListByPeriod(start time.Time, end time.Time) ([]domain.Sale, error) {
	var saleEntities []entity.SaleEntity
	if err := r.db.Preload("Customer").Preload("Items").
		Where("sale_date BETWEEN ? AND ?", start, end).
		Find(&saleEntities).Error; err != nil {
		return nil, err
	}

	sales := make([]domain.Sale, len(saleEntities))
	for i, saleEntity := range saleEntities {
		sales[i] = *saleEntity.ToDomain()
	}

	return sales, nil
}

func (r *saleRepository) filteredSaleIDs(filter domain.SaleListFilter) *gorm.DB {
	query := r.db.Model(&entity.SaleEntity{}).
		Select("DISTINCT sales.id, sales.sale_date").
		Joins("LEFT JOIN customers ON customers.id = sales.customer_id")

	if filter.Search != "" {
		query = query.Joins("LEFT JOIN sale_items ON sale_items.sale_id = sales.id")
	}

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

	if filter.CustomerID != nil {
		query = query.Where("sales.customer_id = ?", *filter.CustomerID)
	}

	if filter.Search != "" {
		term := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(customers.name) LIKE ? OR LOWER(sale_items.product_name) LIKE ? OR LOWER(sale_items.specs) LIKE ?",
			term,
			term,
			term,
		)
	}

	return query
}

func (r *saleRepository) saleListSummary(filteredIDs *gorm.DB, totalItems int64) (domain.SaleListSummary, error) {
	var revenue int
	if err := r.db.Table("sales").
		Select("COALESCE(SUM(sales.total_value), 0)").
		Where("sales.id IN (?)", r.db.Table("(?) AS filtered_sales", filteredIDs).Select("filtered_sales.id")).
		Scan(&revenue).Error; err != nil {
		return domain.SaleListSummary{}, err
	}

	var profit int
	if err := r.db.Table("sale_items").
		Select("COALESCE(SUM((sale_items.sale_price - sale_items.cost_price) * sale_items.quantity), 0)").
		Where("sale_items.sale_id IN (?)", r.db.Table("(?) AS filtered_sales", filteredIDs).Select("filtered_sales.id")).
		Scan(&profit).Error; err != nil {
		return domain.SaleListSummary{}, err
	}

	averageTicket := 0
	if totalItems > 0 {
		averageTicket = revenue / int(totalItems)
	}

	return domain.SaleListSummary{
		Revenue:       revenue,
		Profit:        profit,
		AverageTicket: averageTicket,
	}, nil
}

func (r *saleRepository) UpdateStatus(id int, status domain.PaymentStatus) error {
	return r.db.Model(&entity.SaleEntity{}).
		Where("id = ?", id).
		Update("payment_status", status).Error
}

func (r *saleRepository) Delete(id int) error {
	return r.db.Delete(&entity.SaleEntity{}, id).Error
}

func (r *saleRepository) ListInstallmentAlerts(now time.Time, windowDays int) ([]domain.SaleInstallment, error) {
	start := dateOnly(now)
	end := start.AddDate(0, 0, windowDays)

	var installmentEntities []entity.SaleInstallmentEntity
	if err := r.db.Preload("Sale.Customer").
		Where(
			"(status = ? AND due_date <= ?) OR status = ?",
			domain.InstallmentPending,
			end,
			domain.InstallmentUnpaid,
		).
		Order("due_date ASC, id ASC").
		Find(&installmentEntities).Error; err != nil {
		return nil, err
	}

	installments := make([]domain.SaleInstallment, len(installmentEntities))
	for i := range installmentEntities {
		installments[i] = *installmentEntities[i].ToDomain()
		if installments[i].Status == domain.InstallmentPending && dateOnly(installments[i].DueDate).After(end) {
			continue
		}
		applyInstallmentProgress(r.db, &installments[i])
	}

	return installments, nil
}

func (r *saleRepository) ListInstallmentsBySaleID(saleID int) ([]domain.SaleInstallment, error) {
	var installmentEntities []entity.SaleInstallmentEntity
	if err := r.db.Preload("Sale.Customer").
		Where("sale_id = ?", saleID).
		Order("installment_number ASC").
		Find(&installmentEntities).Error; err != nil {
		return nil, err
	}

	installments := make([]domain.SaleInstallment, len(installmentEntities))
	for i := range installmentEntities {
		installments[i] = *installmentEntities[i].ToDomain()
		applyInstallmentProgress(r.db, &installments[i])
	}

	return installments, nil
}

func (r *saleRepository) UpdateInstallmentStatus(id int, status domain.SaleInstallmentStatus, notes string, validatedAt time.Time) (*domain.SaleInstallment, error) {
	var installmentEntity entity.SaleInstallmentEntity

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&installmentEntity, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return err
		}

		updates := map[string]any{
			"status":       status,
			"validated_at": validatedAt,
			"notes":        notes,
			"paid_at":      nil,
		}
		if status == domain.InstallmentPaid {
			updates["paid_at"] = validatedAt
		}

		if err := tx.Model(&entity.SaleInstallmentEntity{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return err
		}

		var pendingCount int64
		if err := tx.Model(&entity.SaleInstallmentEntity{}).
			Where("sale_id = ? AND status <> ?", installmentEntity.SaleID, domain.InstallmentPaid).
			Count(&pendingCount).Error; err != nil {
			return err
		}

		if pendingCount == 0 {
			if err := tx.Model(&entity.SaleEntity{}).
				Where("id = ?", installmentEntity.SaleID).
				Update("payment_status", domain.PaymentApproved).Error; err != nil {
				return err
			}
		}

		return tx.Preload("Sale.Customer").First(&installmentEntity, id).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	installment := installmentEntity.ToDomain()
	applyInstallmentProgress(r.db, installment)
	return installment, nil
}

func applyInstallmentProgress(db *gorm.DB, installment *domain.SaleInstallment) {
	if installment == nil || installment.SaleID == 0 {
		return
	}

	var paidCount int64
	var totalCount int64
	db.Model(&entity.SaleInstallmentEntity{}).
		Where("sale_id = ? AND status = ?", installment.SaleID, domain.InstallmentPaid).
		Count(&paidCount)
	db.Model(&entity.SaleInstallmentEntity{}).
		Where("sale_id = ?", installment.SaleID).
		Count(&totalCount)

	installment.PaidInstallments = int(paidCount)
	installment.RemainingInstallments = int(totalCount - paidCount)
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
