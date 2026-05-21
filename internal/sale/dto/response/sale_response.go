package response

import (
	"istore/internal/sale/domain"
	"time"
)

type SaleResponse struct {
	ID            int                  `json:"id"`
	CustomerID    int                  `json:"customerId"`
	TotalValue    int                  `json:"totalValue"`
	PaymentStatus domain.PaymentStatus `json:"paymentStatus"`
	PaymentType   domain.PaymentType   `json:"paymentType"`
	SaleDate      time.Time            `json:"saleDate"`
	Items         []SaleItemResponse   `json:"items"`
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
		TotalValue:    sale.TotalValue,
		PaymentStatus: sale.PaymentStatus,
		PaymentType:   sale.PaymentType,
		SaleDate:      sale.SaleDate,
		Items:         items,
	}
}
