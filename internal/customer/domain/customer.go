package domain

import (
	saleDomain "istore/internal/sale/domain"
	"time"
)

type Customer struct {
	ID     int
	UserID uint
	Name   string
	Phone  string
}

type CustomerListFilter struct {
	UserID        uint
	Page          int
	Limit         int
	Start         *time.Time
	End           *time.Time
	PaymentStatus *saleDomain.PaymentStatus
	PaymentType   *saleDomain.PaymentType
	Search        string
}

type CustomerListItem struct {
	ID            int
	Name          string
	Phone         string
	SalesCount    int
	Revenue       int
	Profit        int
	AverageTicket int
}

type CustomerListSummary struct {
	TotalCustomers int
	SalesCount     int
	Revenue        int
	Profit         int
	AverageTicket  int
	RepeatRate     int
}

type CustomerListResult struct {
	Items      []CustomerListItem
	Page       int
	Limit      int
	TotalItems int
	TotalPages int
	Summary    CustomerListSummary
}
