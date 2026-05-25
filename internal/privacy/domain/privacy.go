package domain

import "time"

type RequestType string

const (
	RequestAccess      RequestType = "ACCESS"
	RequestCorrection  RequestType = "CORRECTION"
	RequestExport      RequestType = "EXPORT"
	RequestDeletion    RequestType = "DELETION"
	RequestPortability RequestType = "PORTABILITY"
)

type RequestStatus string

const (
	RequestOpen     RequestStatus = "OPEN"
	RequestInReview RequestStatus = "IN_REVIEW"
	RequestDone     RequestStatus = "DONE"
	RequestRejected RequestStatus = "REJECTED"
)

type PrivacyRequest struct {
	ID        uint
	UserID    uint
	Type      RequestType
	Status    RequestStatus
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AccountExport struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ConsentExport struct {
	PrivacyPolicyVersion string     `json:"privacyPolicyVersion"`
	PrivacyAcceptedAt    *time.Time `json:"privacyAcceptedAt"`
	TermsVersion         string     `json:"termsVersion"`
	TermsAcceptedAt      *time.Time `json:"termsAcceptedAt"`
}

type CustomerExport struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SaleItemExport struct {
	ID          int    `json:"id"`
	ProductName string `json:"productName"`
	Specs       string `json:"specs"`
	Quantity    int    `json:"quantity"`
	CostPrice   int    `json:"costPrice"`
	SalePrice   int    `json:"salePrice"`
}

type SaleInstallmentExport struct {
	ID                int        `json:"id"`
	DueDate           time.Time  `json:"dueDate"`
	InstallmentNumber int        `json:"installmentNumber"`
	TotalInstallments int        `json:"totalInstallments"`
	Amount            int        `json:"amount"`
	Status            string     `json:"status"`
	PaidAt            *time.Time `json:"paidAt"`
	ValidatedAt       *time.Time `json:"validatedAt"`
	Notes             string     `json:"notes"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type SaleExport struct {
	ID               int                     `json:"id"`
	CustomerID       int                     `json:"customerId"`
	CustomerName     string                  `json:"customerName"`
	TotalValue       int                     `json:"totalValue"`
	PaymentStatus    string                  `json:"paymentStatus"`
	PaymentType      string                  `json:"paymentType"`
	SaleDate         time.Time               `json:"saleDate"`
	Installments     *int                    `json:"installments"`
	BillingDay       *int                    `json:"billingDay"`
	Items            []SaleItemExport        `json:"items"`
	InstallmentsList []SaleInstallmentExport `json:"installmentsList"`
}

type PrivacyExport struct {
	ExportedAt time.Time        `json:"exportedAt"`
	Account    AccountExport    `json:"account"`
	Consents   ConsentExport    `json:"consents"`
	Customers  []CustomerExport `json:"customers"`
	Sales      []SaleExport     `json:"sales"`
}
