package implementation

import (
	"istore/internal/customer/domain"
	serviceContracts "istore/internal/customer/service/contracts"
	saleDomain "istore/internal/sale/domain"
	"testing"
	"time"
)

type fakeCustomerRepository struct {
	filter domain.CustomerListFilter
}

func (f *fakeCustomerRepository) Create(customer *domain.Customer) error { return nil }
func (f *fakeCustomerRepository) Update(customer *domain.Customer) error { return nil }
func (f *fakeCustomerRepository) Delete(id int) error                    { return nil }
func (f *fakeCustomerRepository) FindByID(id int) (*domain.Customer, error) {
	return &domain.Customer{ID: id}, nil
}
func (f *fakeCustomerRepository) FindAll() ([]domain.Customer, error) { return nil, nil }
func (f *fakeCustomerRepository) List(filter domain.CustomerListFilter) (*domain.CustomerListResult, error) {
	f.filter = filter
	return &domain.CustomerListResult{Page: filter.Page, Limit: filter.Limit}, nil
}

func TestListRejectsInvalidPaginationAndFilters(t *testing.T) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	start := time.Date(2026, time.May, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	status := saleDomain.PaymentStatus("INVALID")
	paymentType := saleDomain.PaymentType("INVALID")

	tests := []struct {
		name  string
		input serviceContracts.ListCustomersInput
	}{
		{name: "page", input: serviceContracts.ListCustomersInput{Page: 0, Limit: 10}},
		{name: "limit", input: serviceContracts.ListCustomersInput{Page: 1, Limit: 101}},
		{name: "date range", input: serviceContracts.ListCustomersInput{Page: 1, Limit: 10, Start: &start, End: &end}},
		{name: "status", input: serviceContracts.ListCustomersInput{Page: 1, Limit: 10, Status: &status}},
		{name: "payment type", input: serviceContracts.ListCustomersInput{Page: 1, Limit: 10, PaymentType: &paymentType}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, restErr := service.List(tt.input); restErr == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestListPassesNormalizedFilterToRepository(t *testing.T) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)
	start := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	status := saleDomain.PaymentApproved

	result, restErr := service.List(serviceContracts.ListCustomersInput{
		Page:   2,
		Limit:  25,
		Start:  &start,
		Status: &status,
		Search: "ana",
	})
	if restErr != nil {
		t.Fatalf("list customers: %v", restErr)
	}

	if result.Page != 2 || repository.filter.Limit != 25 || repository.filter.Start == nil || *repository.filter.PaymentStatus != saleDomain.PaymentApproved || repository.filter.Search != "ana" {
		t.Fatalf("unexpected filter: %#v", repository.filter)
	}
}
