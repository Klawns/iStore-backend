package domain

import (
	"time"
)

type PaymentStatus string
type PaymentType string

const (
	PaymentPending  PaymentStatus = "PENDING"
	PaymentApproved PaymentStatus = "APPROVED"
	PaymentCanceled PaymentStatus = "CANCELED"

	Pix   PaymentType = "PIX"
	Money PaymentType = "MONEY"
	Card  PaymentType = "CARD"
)

type Sale struct {
	ID            int
	CustomerID    int
	TotalValue    int // centavos
	PaymentStatus PaymentStatus
	PaymentType   PaymentType
	SaleDate      time.Time

	Items []SaleItem
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
