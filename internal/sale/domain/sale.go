package domain

import (
	"time"
)

type PaymentStatus string
type PaymentType string
type SaleInstallmentStatus string

const (
	PaymentPending  PaymentStatus = "PENDING"
	PaymentApproved PaymentStatus = "APPROVED"
	PaymentCanceled PaymentStatus = "CANCELED"

	Pix        PaymentType = "PIX"
	Money      PaymentType = "MONEY"
	CreditCard PaymentType = "CREDIT_CARD"
	DebitCard  PaymentType = "DEBIT_CARD"
)

const (
	InstallmentPending SaleInstallmentStatus = "PENDING"
	InstallmentPaid    SaleInstallmentStatus = "PAID"
	InstallmentUnpaid  SaleInstallmentStatus = "UNPAID"
)

type Sale struct {
	ID            int
	CustomerID    int
	CustomerName  string
	TotalValue    int // centavos
	PaymentStatus PaymentStatus
	PaymentType   PaymentType
	SaleDate      time.Time
	Installments  *int
	BillingDay    *int

	Items            []SaleItem
	InstallmentsList []SaleInstallment
}

type SaleListFilter struct {
	Page          int
	Limit         int
	Start         *time.Time
	End           *time.Time
	PaymentStatus *PaymentStatus
	PaymentType   *PaymentType
	CustomerID    *int
	Search        string
}

type SaleListSummary struct {
	Revenue       int
	Profit        int
	AverageTicket int
}

type SaleListResult struct {
	Items      []Sale
	Page       int
	Limit      int
	TotalItems int
	TotalPages int
	Summary    SaleListSummary
}

type SaleInstallment struct {
	ID                    int
	SaleID                int
	CustomerName          string
	DueDate               time.Time
	InstallmentNumber     int
	TotalInstallments     int
	Amount                int
	Status                SaleInstallmentStatus
	PaidAt                *time.Time
	ValidatedAt           *time.Time
	Notes                 string
	PaidInstallments      int
	RemainingInstallments int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type PaymentAlertStatus string

const (
	PaymentAlertOpen PaymentAlertStatus = "OPEN"
)

type PaymentAlert struct {
	ID                int
	SaleID            int
	DueDate           time.Time
	InstallmentNumber int
	Message           string
	Status            PaymentAlertStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *Sale) CalculateTotal() int {
	total := 0

	for _, item := range s.Items {
		total += item.Total()
	}

	return total
}

func (s *Sale) CalculateProfit() int {
	total := 0

	for _, item := range s.Items {
		total += item.Profit()
	}

	return total
}
