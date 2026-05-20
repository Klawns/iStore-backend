package main

import (
	customerEntity "istore/internal/customer/repository/entity"
	userEntity "istore/internal/users/repository/entity"

	"gorm.io/gorm"
)

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&userEntity.UserEntity{},
		&customerEntity.CustomerEntity{},
	)
}
