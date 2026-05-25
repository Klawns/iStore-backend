package implementation

import (
	customerEntity "istore/internal/customer/repository/entity"
	"istore/internal/privacy/domain"
	"istore/internal/privacy/repository/contracts"
	privacyEntity "istore/internal/privacy/repository/entity"
	saleEntity "istore/internal/sale/repository/entity"
	userEntity "istore/internal/users/repository/entity"
	"time"

	"gorm.io/gorm"
)

type privacyRepository struct {
	db *gorm.DB
}

func NewPrivacyRepository(db *gorm.DB) contracts.PrivacyRepository {
	return &privacyRepository{db: db}
}

func (r *privacyRepository) Create(request *domain.PrivacyRequest) error {
	requestEntity := privacyEntity.FromDomain(request)
	if err := r.db.Create(requestEntity).Error; err != nil {
		return err
	}

	*request = *requestEntity.ToDomain()
	return nil
}

func (r *privacyRepository) ListByUserID(userID uint) ([]domain.PrivacyRequest, error) {
	var requestEntities []privacyEntity.PrivacyRequestEntity
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC, id DESC").Find(&requestEntities).Error; err != nil {
		return nil, err
	}

	requests := make([]domain.PrivacyRequest, len(requestEntities))
	for i := range requestEntities {
		requests[i] = *requestEntities[i].ToDomain()
	}

	return requests, nil
}

func (r *privacyRepository) ExportByUserID(userID uint) (*domain.PrivacyExport, error) {
	var user userEntity.UserEntity
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var customerEntities []customerEntity.CustomerEntity
	if err := r.db.Where("user_id = ?", userID).Order("id ASC").Find(&customerEntities).Error; err != nil {
		return nil, err
	}

	var saleEntities []saleEntity.SaleEntity
	if err := r.db.Preload("Customer").Preload("Items").Preload("InstallmentsList").
		Where("user_id = ?", userID).
		Order("sale_date ASC, id ASC").
		Find(&saleEntities).Error; err != nil {
		return nil, err
	}

	customers := make([]domain.CustomerExport, len(customerEntities))
	for i := range customerEntities {
		customers[i] = domain.CustomerExport{
			ID:        customerEntities[i].ID,
			Name:      customerEntities[i].Name,
			Phone:     customerEntities[i].Phone,
			CreatedAt: customerEntities[i].CreatedAt,
			UpdatedAt: customerEntities[i].UpdatedAt,
		}
	}

	sales := make([]domain.SaleExport, len(saleEntities))
	for i := range saleEntities {
		items := make([]domain.SaleItemExport, len(saleEntities[i].Items))
		for j := range saleEntities[i].Items {
			items[j] = domain.SaleItemExport{
				ID:          saleEntities[i].Items[j].ID,
				ProductName: saleEntities[i].Items[j].ProductName,
				Specs:       saleEntities[i].Items[j].Specs,
				Quantity:    saleEntities[i].Items[j].Quantity,
				CostPrice:   saleEntities[i].Items[j].CostPrice,
				SalePrice:   saleEntities[i].Items[j].SalePrice,
			}
		}

		installments := make([]domain.SaleInstallmentExport, len(saleEntities[i].InstallmentsList))
		for j := range saleEntities[i].InstallmentsList {
			installments[j] = domain.SaleInstallmentExport{
				ID:                saleEntities[i].InstallmentsList[j].ID,
				DueDate:           saleEntities[i].InstallmentsList[j].DueDate,
				InstallmentNumber: saleEntities[i].InstallmentsList[j].InstallmentNumber,
				TotalInstallments: saleEntities[i].InstallmentsList[j].TotalInstallments,
				Amount:            saleEntities[i].InstallmentsList[j].Amount,
				Status:            string(saleEntities[i].InstallmentsList[j].Status),
				PaidAt:            saleEntities[i].InstallmentsList[j].PaidAt,
				ValidatedAt:       saleEntities[i].InstallmentsList[j].ValidatedAt,
				Notes:             saleEntities[i].InstallmentsList[j].Notes,
				CreatedAt:         saleEntities[i].InstallmentsList[j].CreatedAt,
				UpdatedAt:         saleEntities[i].InstallmentsList[j].UpdatedAt,
			}
		}

		sales[i] = domain.SaleExport{
			ID:               saleEntities[i].ID,
			CustomerID:       saleEntities[i].CustomerID,
			CustomerName:     saleEntities[i].Customer.Name,
			TotalValue:       saleEntities[i].TotalValue,
			PaymentStatus:    string(saleEntities[i].PaymentStatus),
			PaymentType:      string(saleEntities[i].PaymentType),
			SaleDate:         saleEntities[i].SaleDate,
			Installments:     saleEntities[i].Installments,
			BillingDay:       saleEntities[i].BillingDay,
			Items:            items,
			InstallmentsList: installments,
		}
	}

	return &domain.PrivacyExport{
		ExportedAt: time.Now().UTC(),
		Account: domain.AccountExport{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Consents: domain.ConsentExport{
			PrivacyPolicyVersion: user.PrivacyPolicyVersion,
			PrivacyAcceptedAt:    user.PrivacyAcceptedAt,
			TermsVersion:         user.TermsVersion,
			TermsAcceptedAt:      user.TermsAcceptedAt,
		},
		Customers: customers,
		Sales:     sales,
	}, nil
}
