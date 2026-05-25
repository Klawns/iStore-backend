package entity

import (
	"istore/internal/users/domain"
	"time"
)

type UserEntity struct {
	ID                   uint   `gorm:"primaryKey"`
	Email                string `gorm:"uniqueIndex;not null"`
	PasswordHash         string `gorm:"not null"`
	PrivacyPolicyVersion string `gorm:"column:privacy_policy_version"`
	PrivacyAcceptedAt    *time.Time
	TermsVersion         string `gorm:"column:terms_version"`
	TermsAcceptedAt      *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (UserEntity) TableName() string {
	return "users"
}

func FromDomain(user *domain.User) *UserEntity {
	if user == nil {
		return nil
	}

	return &UserEntity{
		ID:                   user.ID,
		Email:                user.Email,
		PasswordHash:         user.PasswordHash,
		PrivacyPolicyVersion: user.PrivacyPolicyVersion,
		PrivacyAcceptedAt:    user.PrivacyAcceptedAt,
		TermsVersion:         user.TermsVersion,
		TermsAcceptedAt:      user.TermsAcceptedAt,
	}
}

func (u *UserEntity) ToDomain() *domain.User {
	if u == nil {
		return nil
	}

	return &domain.User{
		ID:                   u.ID,
		Email:                u.Email,
		PasswordHash:         u.PasswordHash,
		PrivacyPolicyVersion: u.PrivacyPolicyVersion,
		PrivacyAcceptedAt:    u.PrivacyAcceptedAt,
		TermsVersion:         u.TermsVersion,
		TermsAcceptedAt:      u.TermsAcceptedAt,
	}
}
