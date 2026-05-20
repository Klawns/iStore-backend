package database

import (
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	// "github.com/glebarez/sqlite" // Pure-Go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	// "github.com/libtnb/sqlite" // Pure-Go SQLite driver, checkout https://github.com/libtnb/sqlite for details
	"gorm.io/gorm"
)

func Connect(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
