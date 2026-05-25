package domain

import "time"

type User struct {
	ID                   uint
	Email                string
	PasswordHash         string
	PrivacyPolicyVersion string
	PrivacyAcceptedAt    *time.Time
	TermsVersion         string
	TermsAcceptedAt      *time.Time
}
