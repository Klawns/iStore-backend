package response

import "istore/internal/customer/domain"

type CustomerResponse struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func FromDomain(customer *domain.Customer) *CustomerResponse {
	if customer == nil {
		return nil
	}

	return &CustomerResponse{
		Name:  customer.Name,
		Phone: customer.Phone,
	}
}
