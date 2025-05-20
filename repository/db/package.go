package db

import "gorm.io/gorm"

type PackageRepository struct {
	DB *gorm.DB
}

func NewPackageRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{
		DB: db,
	}
}
