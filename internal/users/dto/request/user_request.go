package request

type UserRequest struct {
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=6"`
	AcceptPrivacyPolicy  bool   `json:"acceptPrivacyPolicy" validate:"required"`
	AcceptTerms          bool   `json:"acceptTerms" validate:"required"`
	PrivacyPolicyVersion string `json:"privacyPolicyVersion"`
	TermsVersion         string `json:"termsVersion"`
}

type DeleteOwnAccountRequest struct {
	Password string `json:"password" validate:"required"`
}
