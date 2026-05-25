package response

import (
	"istore/internal/users/domain"
	"time"
)

type UserResponse struct {
	ID                   uint       `json:"id"`
	Email                string     `json:"email"`
	PrivacyPolicyVersion string     `json:"privacyPolicyVersion"`
	PrivacyAcceptedAt    *time.Time `json:"privacyAcceptedAt"`
	TermsVersion         string     `json:"termsVersion"`
	TermsAcceptedAt      *time.Time `json:"termsAcceptedAt"`
}

func FromDomain(user *domain.User) *UserResponse {
	if user == nil {
		return nil
	}

	return &UserResponse{
		ID:                   user.ID,
		Email:                user.Email,
		PrivacyPolicyVersion: user.PrivacyPolicyVersion,
		PrivacyAcceptedAt:    user.PrivacyAcceptedAt,
		TermsVersion:         user.TermsVersion,
		TermsAcceptedAt:      user.TermsAcceptedAt,
	}
}
