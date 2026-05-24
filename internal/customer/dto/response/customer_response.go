package response

import "istore/internal/customer/domain"

type CustomerResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type CustomerListItemResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	SalesCount    int    `json:"salesCount"`
	Revenue       int    `json:"revenue"`
	Profit        int    `json:"profit"`
	AverageTicket int    `json:"averageTicket"`
}

type CustomerListSummaryResponse struct {
	TotalCustomers int `json:"totalCustomers"`
	SalesCount     int `json:"salesCount"`
	Revenue        int `json:"revenue"`
	Profit         int `json:"profit"`
	AverageTicket  int `json:"averageTicket"`
	RepeatRate     int `json:"repeatRate"`
}

type CustomerListResponse struct {
	Items      []CustomerListItemResponse  `json:"items"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	TotalItems int                         `json:"totalItems"`
	TotalPages int                         `json:"totalPages"`
	Summary    CustomerListSummaryResponse `json:"summary"`
}

func FromDomain(customer *domain.Customer) *CustomerResponse {
	if customer == nil {
		return nil
	}

	return &CustomerResponse{
		ID:    customer.ID,
		Name:  customer.Name,
		Phone: customer.Phone,
	}
}

func ListFromDomain(result *domain.CustomerListResult) *CustomerListResponse {
	if result == nil {
		return nil
	}

	items := make([]CustomerListItemResponse, len(result.Items))
	for i := range result.Items {
		item := result.Items[i]
		items[i] = CustomerListItemResponse{
			ID:            item.ID,
			Name:          item.Name,
			Phone:         item.Phone,
			SalesCount:    item.SalesCount,
			Revenue:       item.Revenue,
			Profit:        item.Profit,
			AverageTicket: item.AverageTicket,
		}
	}

	return &CustomerListResponse{
		Items:      items,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalItems: result.TotalItems,
		TotalPages: result.TotalPages,
		Summary: CustomerListSummaryResponse{
			TotalCustomers: result.Summary.TotalCustomers,
			SalesCount:     result.Summary.SalesCount,
			Revenue:        result.Summary.Revenue,
			Profit:         result.Summary.Profit,
			AverageTicket:  result.Summary.AverageTicket,
			RepeatRate:     result.Summary.RepeatRate,
		},
	}
}
