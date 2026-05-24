package response

import (
	"istore/internal/sale/domain"
	"time"
)

type SaleResponse struct {
	ID            int                  `json:"id"`
	CustomerID    int                  `json:"customerId"`
	CustomerName  string               `json:"customerName"`
	TotalValue    int                  `json:"totalValue"`
	PaymentStatus domain.PaymentStatus `json:"paymentStatus"`
	PaymentType   domain.PaymentType   `json:"paymentType"`
	SaleDate      time.Time            `json:"saleDate"`
	Installments  *int                 `json:"installments,omitempty"`
	BillingDay    *int                 `json:"billingDay,omitempty"`
	Items         []SaleItemResponse   `json:"items"`
}

type SalesSummaryResponse struct {
	Revenue       int `json:"revenue"`
	Profit        int `json:"profit"`
	AverageTicket int `json:"averageTicket"`
}

type SalesListResponse struct {
	Items      []SaleResponse       `json:"items"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalItems int                  `json:"totalItems"`
	TotalPages int                  `json:"totalPages"`
	Summary    SalesSummaryResponse `json:"summary"`
}

type SaleInstallmentResponse struct {
	ID                    int                          `json:"id"`
	SaleID                int                          `json:"saleId"`
	CustomerName          string                       `json:"customerName"`
	DueDate               time.Time                    `json:"dueDate"`
	InstallmentNumber     int                          `json:"installmentNumber"`
	TotalInstallments     int                          `json:"totalInstallments"`
	Amount                int                          `json:"amount"`
	Status                domain.SaleInstallmentStatus `json:"status"`
	PaidAt                *time.Time                   `json:"paidAt,omitempty"`
	ValidatedAt           *time.Time                   `json:"validatedAt,omitempty"`
	Notes                 string                       `json:"notes,omitempty"`
	PaidInstallments      int                          `json:"paidInstallments"`
	RemainingInstallments int                          `json:"remainingInstallments"`
	CreatedAt             time.Time                    `json:"createdAt"`
	UpdatedAt             time.Time                    `json:"updatedAt"`
}

type SaleItemResponse struct {
	ID          int    `json:"id"`
	SaleID      int    `json:"saleId"`
	ProductName string `json:"productName"`
	Specs       string `json:"specs"`
	Quantity    int    `json:"quantity"`
	CostPrice   int    `json:"costPrice"`
	SalePrice   int    `json:"salePrice"`
}

func ListFromDomain(result *domain.SaleListResult) *SalesListResponse {
	if result == nil {
		return nil
	}

	items := make([]SaleResponse, len(result.Items))
	for i := range result.Items {
		items[i] = *FromDomain(&result.Items[i])
	}

	return &SalesListResponse{
		Items:      items,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalItems: result.TotalItems,
		TotalPages: result.TotalPages,
		Summary: SalesSummaryResponse{
			Revenue:       result.Summary.Revenue,
			Profit:        result.Summary.Profit,
			AverageTicket: result.Summary.AverageTicket,
		},
	}
}

func FromDomain(sale *domain.Sale) *SaleResponse {
	if sale == nil {
		return nil
	}

	items := make([]SaleItemResponse, len(sale.Items))
	for i := range sale.Items {
		item := sale.Items[i]
		items[i] = SaleItemResponse{
			ID:          item.ID,
			SaleID:      item.SaleID,
			ProductName: item.ProductName,
			Specs:       item.Specs,
			Quantity:    item.Quantity,
			CostPrice:   item.CostPrice,
			SalePrice:   item.SalePrice,
		}
	}

	return &SaleResponse{
		ID:            sale.ID,
		CustomerID:    sale.CustomerID,
		CustomerName:  sale.CustomerName,
		TotalValue:    sale.TotalValue,
		PaymentStatus: sale.PaymentStatus,
		PaymentType:   sale.PaymentType,
		SaleDate:      sale.SaleDate,
		Installments:  sale.Installments,
		BillingDay:    sale.BillingDay,
		Items:         items,
	}
}

func SaleInstallmentFromDomain(installment *domain.SaleInstallment) *SaleInstallmentResponse {
	if installment == nil {
		return nil
	}

	return &SaleInstallmentResponse{
		ID:                    installment.ID,
		SaleID:                installment.SaleID,
		CustomerName:          installment.CustomerName,
		DueDate:               installment.DueDate,
		InstallmentNumber:     installment.InstallmentNumber,
		TotalInstallments:     installment.TotalInstallments,
		Amount:                installment.Amount,
		Status:                installment.Status,
		PaidAt:                installment.PaidAt,
		ValidatedAt:           installment.ValidatedAt,
		Notes:                 installment.Notes,
		PaidInstallments:      installment.PaidInstallments,
		RemainingInstallments: installment.RemainingInstallments,
		CreatedAt:             installment.CreatedAt,
		UpdatedAt:             installment.UpdatedAt,
	}
}
