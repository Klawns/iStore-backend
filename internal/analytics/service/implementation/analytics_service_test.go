package implementation

import (
	"istore/internal/analytics/domain"
	saleDomain "istore/internal/sale/domain"
	"testing"
)

func TestNormalizeFilterRejectsInvalidPaymentType(t *testing.T) {
	_, restErr := normalizeFilter(domain.AnalyticsFilter{
		PaymentType: saleDomain.PaymentType("BOLETO"),
	})

	if restErr == nil {
		t.Fatal("expected invalid payment type error")
	}
}

func TestNormalizeFilterKeepsApprovedAsDefaultStatus(t *testing.T) {
	filter, restErr := normalizeFilter(domain.AnalyticsFilter{})
	if restErr != nil {
		t.Fatalf("normalize filter: %v", restErr)
	}

	if filter.Status != saleDomain.PaymentApproved {
		t.Fatalf("expected default status %s, got %s", saleDomain.PaymentApproved, filter.Status)
	}
}
