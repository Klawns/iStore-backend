package implementation

import (
	"istore/internal/sale/domain"
	"istore/internal/sale/service/contract"
	"testing"
	"time"
)

type fakeSaleRepository struct {
	createdSale  *domain.Sale
	installments []domain.SaleInstallment
}

func (f *fakeSaleRepository) Create(sale *domain.Sale) error {
	sale.ID = 1
	f.createdSale = sale
	return nil
}

func (f *fakeSaleRepository) FindByID(id int) (*domain.Sale, error) { return nil, nil }
func (f *fakeSaleRepository) FindAll() ([]domain.Sale, error)       { return nil, nil }
func (f *fakeSaleRepository) List(filter domain.SaleListFilter) (*domain.SaleListResult, error) {
	return &domain.SaleListResult{Page: filter.Page, Limit: filter.Limit}, nil
}
func (f *fakeSaleRepository) ListByPeriod(start time.Time, end time.Time) ([]domain.Sale, error) {
	return nil, nil
}
func (f *fakeSaleRepository) UpdateStatus(id int, status domain.PaymentStatus) error { return nil }
func (f *fakeSaleRepository) Delete(id int) error                                    { return nil }
func (f *fakeSaleRepository) ListInstallmentAlerts(now time.Time, windowDays int) ([]domain.SaleInstallment, error) {
	return f.installments, nil
}
func (f *fakeSaleRepository) ListInstallmentsBySaleID(saleID int) ([]domain.SaleInstallment, error) {
	return f.installments, nil
}
func (f *fakeSaleRepository) UpdateInstallmentStatus(id int, status domain.SaleInstallmentStatus, notes string, validatedAt time.Time) (*domain.SaleInstallment, error) {
	return &domain.SaleInstallment{ID: id, Status: status, Notes: notes, ValidatedAt: &validatedAt}, nil
}

func TestCreateDebitCardForcesApprovedAndRejectsCardFields(t *testing.T) {
	repo := &fakeSaleRepository{}
	service := NewSaleService(repo)

	input := validCreateInput()
	input.TipoPagamento = domain.DebitCard
	input.StatusPagamento = domain.PaymentPending

	sale, restErr := service.Create(input)
	if restErr != nil {
		t.Fatalf("expected debit sale to be accepted: %v", restErr)
	}
	if sale.PaymentStatus != domain.PaymentApproved {
		t.Fatalf("expected debit sale to be approved, got %s", sale.PaymentStatus)
	}

	installments := 1
	input = validCreateInput()
	input.TipoPagamento = domain.DebitCard
	input.Installments = &installments
	if _, restErr = service.Create(input); restErr == nil {
		t.Fatal("expected debit sale with installments to be rejected")
	}
}

func TestCreateCreditCardValidatesInstallmentsAndBillingDay(t *testing.T) {
	repo := &fakeSaleRepository{}
	service := NewSaleService(repo)

	input := validCreateInput()
	input.TipoPagamento = domain.CreditCard
	if _, restErr := service.Create(input); restErr == nil {
		t.Fatal("expected credit sale without installments and billing day to be rejected")
	}

	installments := 24
	billingDay := 31
	input.Installments = &installments
	input.BillingDay = &billingDay
	if _, restErr := service.Create(input); restErr != nil {
		t.Fatalf("expected valid credit sale to be accepted: %v", restErr)
	}
	if len(repo.createdSale.InstallmentsList) != 24 {
		t.Fatalf("expected 24 installments, got %d", len(repo.createdSale.InstallmentsList))
	}
}

func TestCreatePixAndMoneyRejectCardFields(t *testing.T) {
	service := NewSaleService(&fakeSaleRepository{})
	installments := 2

	for _, paymentType := range []domain.PaymentType{domain.Pix, domain.Money} {
		input := validCreateInput()
		input.TipoPagamento = paymentType
		input.Installments = &installments

		if _, restErr := service.Create(input); restErr == nil {
			t.Fatalf("expected %s with card fields to be rejected", paymentType)
		}
	}
}

func TestListRejectsInvalidPaginationAndFilters(t *testing.T) {
	service := NewSaleService(&fakeSaleRepository{})
	status := domain.PaymentStatus("INVALID")
	paymentType := domain.PaymentType("BOLETO")
	customerID := 0
	start := time.Date(2026, time.May, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		input contract.ListSalesInput
	}{
		{name: "page", input: contract.ListSalesInput{Page: 0, Limit: 10}},
		{name: "limit", input: contract.ListSalesInput{Page: 1, Limit: 101}},
		{name: "date range", input: contract.ListSalesInput{Page: 1, Limit: 10, Start: &start, End: &end}},
		{name: "status", input: contract.ListSalesInput{Page: 1, Limit: 10, Status: &status}},
		{name: "payment type", input: contract.ListSalesInput{Page: 1, Limit: 10, PaymentType: &paymentType}},
		{name: "customer", input: contract.ListSalesInput{Page: 1, Limit: 10, CustomerID: &customerID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, restErr := service.List(tt.input); restErr == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCreditCardDueDatesClampBillingDayToMonthEnd(t *testing.T) {
	installments := 3
	billingDay := 31
	sale := domain.Sale{
		SaleDate:     time.Date(2026, time.January, 30, 12, 0, 0, 0, time.UTC),
		Installments: &installments,
		BillingDay:   &billingDay,
	}

	dates := creditCardDueDates(sale)
	expected := []string{"2026-01-31", "2026-02-28", "2026-03-31"}
	for i, expectedDate := range expected {
		if dates[i].Format(time.DateOnly) != expectedDate {
			t.Fatalf("due date %d: expected %s, got %s", i, expectedDate, dates[i].Format(time.DateOnly))
		}
	}
}

func TestBuildSaleInstallmentsSplitsAmountsAndDueDates(t *testing.T) {
	installments := 3
	billingDay := 31
	repo := &fakeSaleRepository{}
	service := NewSaleService(repo)

	input := validCreateInput()
	input.TipoPagamento = domain.CreditCard
	input.SaleDate = time.Date(2026, time.January, 30, 12, 0, 0, 0, time.UTC)
	input.Installments = &installments
	input.BillingDay = &billingDay
	input.Itens[0].SalePrice = 100

	if _, restErr := service.Create(input); restErr != nil {
		t.Fatalf("expected credit sale to be accepted: %v", restErr)
	}

	got := repo.createdSale.InstallmentsList
	expectedDates := []string{"2026-01-31", "2026-02-28", "2026-03-31"}
	expectedAmounts := []int{34, 33, 33}
	for i := range got {
		if got[i].DueDate.Format(time.DateOnly) != expectedDates[i] {
			t.Fatalf("installment %d due date: expected %s, got %s", i, expectedDates[i], got[i].DueDate.Format(time.DateOnly))
		}
		if got[i].Amount != expectedAmounts[i] {
			t.Fatalf("installment %d amount: expected %d, got %d", i, expectedAmounts[i], got[i].Amount)
		}
	}
}

func validCreateInput() *contract.CreateSaleInput {
	return &contract.CreateSaleInput{
		ClienteID:       1,
		TipoPagamento:   domain.Pix,
		StatusPagamento: domain.PaymentPending,
		SaleDate:        time.Date(2026, time.May, 23, 12, 0, 0, 0, time.UTC),
		Itens: []contract.CreateSaleItemInput{
			{
				ProductName: "iPhone",
				Quantity:    1,
				CostPrice:   100,
				SalePrice:   200,
			},
		},
	}
}
