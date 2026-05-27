package main

import (
	customerEntity "istore/internal/customer/repository/entity"
	privacyEntity "istore/internal/privacy/repository/entity"
	saleDomain "istore/internal/sale/domain"
	saleEntity "istore/internal/sale/repository/entity"
	userEntity "istore/internal/users/repository/entity"
	"time"

	"gorm.io/gorm"
)

func runMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&userEntity.UserEntity{},
		&customerEntity.CustomerEntity{},
		&saleEntity.SaleEntity{},
		&saleEntity.SaleItemEntity{},
		&saleEntity.SaleInstallmentEntity{},
		&saleEntity.PaymentAlertEntity{},
		&privacyEntity.PrivacyRequestEntity{},
	); err != nil {
		return err
	}

	if err := deleteOrphanUserData(db); err != nil {
		return err
	}

	installments := 1
	if err := db.Model(&saleEntity.SaleEntity{}).
		Where("payment_type = ?", "CARD").
		Updates(map[string]any{
			"payment_type": saleDomain.CreditCard,
			"installments": installments,
			"billing_day":  nil,
		}).Error; err != nil {
		return err
	}

	return backfillSaleInstallments(db)
}

func deleteOrphanUserData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		orphanSales := tx.Model(&saleEntity.SaleEntity{}).
			Select("id").
			Where("user_id = 0 OR user_id IS NULL")

		if err := tx.Where("sale_id IN (?)", orphanSales).
			Delete(&saleEntity.PaymentAlertEntity{}).Error; err != nil {
			return err
		}

		if err := tx.Where("sale_id IN (?)", orphanSales).
			Delete(&saleEntity.SaleInstallmentEntity{}).Error; err != nil {
			return err
		}

		if err := tx.Where("sale_id IN (?)", orphanSales).
			Delete(&saleEntity.SaleItemEntity{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = 0 OR user_id IS NULL").
			Delete(&saleEntity.SaleEntity{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = 0 OR user_id IS NULL").
			Delete(&customerEntity.CustomerEntity{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func backfillSaleInstallments(db *gorm.DB) error {
	var sales []saleEntity.SaleEntity
	if err := db.
		Where("payment_type = ? AND installments IS NOT NULL AND billing_day IS NOT NULL", saleDomain.CreditCard).
		Find(&sales).Error; err != nil {
		return err
	}

	for _, sale := range sales {
		var count int64
		if err := db.Model(&saleEntity.SaleInstallmentEntity{}).
			Where("sale_id = ?", sale.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 || sale.Installments == nil || sale.BillingDay == nil || *sale.Installments <= 0 {
			continue
		}

		dueDates := migrationCreditCardDueDates(sale.SaleDate, *sale.Installments, *sale.BillingDay)
		baseAmount := sale.TotalValue / len(dueDates)
		remainder := sale.TotalValue % len(dueDates)

		for i, dueDate := range dueDates {
			amount := baseAmount
			if i < remainder {
				amount++
			}

			installment := saleEntity.SaleInstallmentEntity{
				SaleID:            sale.ID,
				DueDate:           dueDate,
				InstallmentNumber: i + 1,
				TotalInstallments: len(dueDates),
				Amount:            amount,
				Status:            saleDomain.InstallmentPending,
			}
			if err := db.Create(&installment).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func migrationCreditCardDueDates(saleDate time.Time, installments int, billingDay int) []time.Time {
	first := migrationFirstBillingDate(saleDate, billingDay)
	dates := make([]time.Time, installments)
	for i := 0; i < installments; i++ {
		monthIndex := int(first.Month()) - 1 + i
		year := first.Year() + monthIndex/12
		month := time.Month(monthIndex%12 + 1)
		dates[i] = migrationDateWithClampedDay(year, month, billingDay, first.Location())
	}

	return dates
}

func migrationFirstBillingDate(saleDate time.Time, billingDay int) time.Time {
	candidate := migrationDateWithClampedDay(saleDate.Year(), saleDate.Month(), billingDay, saleDate.Location())
	if !candidate.Before(migrationDateOnly(saleDate)) {
		return candidate
	}

	nextMonth := saleDate.AddDate(0, 1, 0)
	return migrationDateWithClampedDay(nextMonth.Year(), nextMonth.Month(), billingDay, saleDate.Location())
}

func migrationDateWithClampedDay(year int, month time.Month, day int, location *time.Location) time.Time {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, location).Day()
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func migrationDateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
