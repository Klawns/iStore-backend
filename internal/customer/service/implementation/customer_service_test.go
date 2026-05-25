package implementation

import (
	"istore/internal/customer/domain"
	serviceContracts "istore/internal/customer/service/contracts"
	saleDomain "istore/internal/sale/domain"
	"testing"
	"time"
)

type fakeCustomerRepository struct {
	filter     domain.CustomerListFilter
	customers  []domain.Customer
	salesCount int64
	deletedIDs []int
}

func (f *fakeCustomerRepository) Create(customer *domain.Customer) error { return nil }
func (f *fakeCustomerRepository) Update(customer *domain.Customer) error { return nil }
func (f *fakeCustomerRepository) Delete(userID uint, id int) error {
	f.deletedIDs = []int{id}
	return nil
}
func (f *fakeCustomerRepository) DeleteMany(userID uint, ids []int) error {
	f.deletedIDs = ids
	return nil
}
func (f *fakeCustomerRepository) FindByID(userID uint, id int) (*domain.Customer, error) {
	return &domain.Customer{ID: id, UserID: userID}, nil
}
func (f *fakeCustomerRepository) FindByIDs(userID uint, ids []int) ([]domain.Customer, error) {
	if f.customers != nil {
		return f.customers, nil
	}

	customers := make([]domain.Customer, len(ids))
	for i, id := range ids {
		customers[i] = domain.Customer{ID: id, UserID: userID}
	}

	return customers, nil
}
func (f *fakeCustomerRepository) FindAll() ([]domain.Customer, error) { return nil, nil }
func (f *fakeCustomerRepository) List(filter domain.CustomerListFilter) (*domain.CustomerListResult, error) {
	f.filter = filter
	return &domain.CustomerListResult{Page: filter.Page, Limit: filter.Limit}, nil
}
func (f *fakeCustomerRepository) CountSalesByCustomerIDs(userID uint, ids []int) (int64, error) {
	return f.salesCount, nil
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
		{name: "page", input: serviceContracts.ListCustomersInput{UserID: 1, Page: 0, Limit: 10}},
		{name: "limit", input: serviceContracts.ListCustomersInput{UserID: 1, Page: 1, Limit: 101}},
		{name: "date range", input: serviceContracts.ListCustomersInput{UserID: 1, Page: 1, Limit: 10, Start: &start, End: &end}},
		{name: "status", input: serviceContracts.ListCustomersInput{UserID: 1, Page: 1, Limit: 10, Status: &status}},
		{name: "payment type", input: serviceContracts.ListCustomersInput{UserID: 1, Page: 1, Limit: 10, PaymentType: &paymentType}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, restErr := service.List(tt.input); restErr == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDeleteBlocksCustomerWithSales(t *testing.T) {
	repository := &fakeCustomerRepository{salesCount: 1}
	service := NewCustomerService(repository)

	restErr := service.Delete(1, 10)
	if restErr == nil || restErr.Code != 409 {
		t.Fatalf("expected conflict, got %#v", restErr)
	}
	if len(repository.deletedIDs) != 0 {
		t.Fatalf("expected no delete, got %#v", repository.deletedIDs)
	}
}

func TestDeleteManyDeduplicatesAndDeletesCustomersWithoutSales(t *testing.T) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)

	deleted, restErr := service.DeleteMany(1, []int{10, 10, 11})
	if restErr != nil {
		t.Fatalf("delete many customers: %v", restErr)
	}
	if deleted != 2 || len(repository.deletedIDs) != 2 {
		t.Fatalf("unexpected delete result: deleted=%d ids=%#v", deleted, repository.deletedIDs)
	}
}

func TestDeleteManyBlocksWholeOperationWhenAnyCustomerHasSales(t *testing.T) {
	repository := &fakeCustomerRepository{salesCount: 1}
	service := NewCustomerService(repository)

	deleted, restErr := service.DeleteMany(1, []int{10, 11})
	if restErr == nil || restErr.Code != 409 {
		t.Fatalf("expected conflict, got deleted=%d err=%#v", deleted, restErr)
	}
	if len(repository.deletedIDs) != 0 {
		t.Fatalf("expected no delete, got %#v", repository.deletedIDs)
	}
}

func TestDeleteManyReturnsNotFoundForMissingUserCustomer(t *testing.T) {
	repository := &fakeCustomerRepository{customers: []domain.Customer{{ID: 10, UserID: 1}}}
	service := NewCustomerService(repository)

	_, restErr := service.DeleteMany(1, []int{10, 99})
	if restErr == nil || restErr.Code != 404 {
		t.Fatalf("expected not found, got %#v", restErr)
	}
}

func TestListPassesNormalizedFilterToRepository(t *testing.T) {
	repository := &fakeCustomerRepository{}
	service := NewCustomerService(repository)
	start := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	status := saleDomain.PaymentApproved

	result, restErr := service.List(serviceContracts.ListCustomersInput{
		Page:   2,
		UserID: 1,
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
