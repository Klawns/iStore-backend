package implementation

import (
	"istore/internal/analytics/domain"
	"istore/internal/analytics/repository/contracts"
	serviceContracts "istore/internal/analytics/service/contracts"
	saleDomain "istore/internal/sale/domain"
	"istore/pkg/logger"
	"istore/pkg/rest_err"

	"go.uber.org/zap"
)

const (
	defaultAnalyticsLimit = 10
	maxAnalyticsLimit     = 100
)

type AnalyticsService struct {
	repository contracts.AnalyticsRepository
}

func NewAnalyticsService(repository contracts.AnalyticsRepository) serviceContracts.AnalyticsService {
	return &AnalyticsService{repository: repository}
}

func (s *AnalyticsService) GetDashboard(filter domain.AnalyticsFilter) (*domain.DashboardMetrics, *rest_err.RestErr) {
	normalized, restErr := normalizeFilter(filter)
	if restErr != nil {
		return nil, restErr
	}

	metrics, err := s.repository.GetDashboardMetrics(normalized)
	if err != nil {
		logger.Error("Erro ao buscar dashboard de analytics: ", err, zap.String("journey", "GetAnalyticsDashboard"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar dashboard de analytics")
	}

	return metrics, nil
}

func (s *AnalyticsService) GetRevenue(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, *rest_err.RestErr) {
	normalized, restErr := normalizeFilter(filter)
	if restErr != nil {
		return nil, restErr
	}

	metrics, err := s.repository.GetRevenue(normalized)
	if err != nil {
		logger.Error("Erro ao buscar receita de analytics: ", err, zap.String("journey", "GetAnalyticsRevenue"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar receita de analytics")
	}

	return metrics, nil
}

func (s *AnalyticsService) GetProfit(filter domain.AnalyticsFilter) ([]domain.FinancialMetric, *rest_err.RestErr) {
	normalized, restErr := normalizeFilter(filter)
	if restErr != nil {
		return nil, restErr
	}

	metrics, err := s.repository.GetProfit(normalized)
	if err != nil {
		logger.Error("Erro ao buscar lucro de analytics: ", err, zap.String("journey", "GetAnalyticsProfit"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar lucro de analytics")
	}

	return metrics, nil
}

func (s *AnalyticsService) GetTopProducts(filter domain.AnalyticsFilter) ([]domain.ProductMetric, *rest_err.RestErr) {
	normalized, restErr := normalizeFilter(filter)
	if restErr != nil {
		return nil, restErr
	}

	metrics, err := s.repository.GetTopProducts(normalized)
	if err != nil {
		logger.Error("Erro ao buscar produtos de analytics: ", err, zap.String("journey", "GetAnalyticsTopProducts"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar produtos de analytics")
	}

	return metrics, nil
}

func (s *AnalyticsService) GetPaymentMethods(filter domain.AnalyticsFilter) ([]domain.PaymentMetric, *rest_err.RestErr) {
	normalized, restErr := normalizeFilter(filter)
	if restErr != nil {
		return nil, restErr
	}

	metrics, err := s.repository.GetPaymentMethods(normalized)
	if err != nil {
		logger.Error("Erro ao buscar pagamentos de analytics: ", err, zap.String("journey", "GetAnalyticsPayments"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar pagamentos de analytics")
	}

	return metrics, nil
}

func (s *AnalyticsService) GetTopCustomers(filter domain.AnalyticsFilter) ([]domain.CustomerMetric, *rest_err.RestErr) {
	normalized, restErr := normalizeFilter(filter)
	if restErr != nil {
		return nil, restErr
	}

	metrics, err := s.repository.GetTopCustomers(normalized)
	if err != nil {
		logger.Error("Erro ao buscar clientes de analytics: ", err, zap.String("journey", "GetAnalyticsTopCustomers"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar clientes de analytics")
	}

	return metrics, nil
}

func (s *AnalyticsService) GetStatuses(filter domain.AnalyticsFilter) ([]domain.StatusMetric, *rest_err.RestErr) {
	normalized, restErr := normalizeFilter(filter)
	if restErr != nil {
		return nil, restErr
	}

	metrics, err := s.repository.GetStatuses(normalized)
	if err != nil {
		logger.Error("Erro ao buscar status de analytics: ", err, zap.String("journey", "GetAnalyticsStatuses"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar status de analytics")
	}

	return metrics, nil
}

func normalizeFilter(filter domain.AnalyticsFilter) (domain.AnalyticsFilter, *rest_err.RestErr) {
	if !filter.StartDate.IsZero() && !filter.EndDate.IsZero() && filter.EndDate.Before(filter.StartDate) {
		return domain.AnalyticsFilter{}, rest_err.NewBadRequestError("Data de termino deve ser posterior a data de inicio")
	}

	if filter.Limit < 0 {
		return domain.AnalyticsFilter{}, rest_err.NewBadRequestError("Limite deve ser maior que zero")
	}

	if filter.Limit == 0 {
		filter.Limit = defaultAnalyticsLimit
	}
	if filter.Limit > maxAnalyticsLimit {
		return domain.AnalyticsFilter{}, rest_err.NewBadRequestError("Limite maximo permitido e 100")
	}

	if filter.Status == "" {
		filter.Status = saleDomain.PaymentApproved
	}
	if !isValidPaymentStatus(filter.Status) {
		return domain.AnalyticsFilter{}, rest_err.NewBadRequestError("Status de pagamento invalido")
	}

	if filter.GroupBy == "" {
		filter.GroupBy = domain.GroupByDaily
	}
	if filter.GroupBy != domain.GroupByDaily && filter.GroupBy != domain.GroupByMonthly {
		return domain.AnalyticsFilter{}, rest_err.NewBadRequestError("Agrupamento invalido")
	}

	return filter, nil
}

func isValidPaymentStatus(status saleDomain.PaymentStatus) bool {
	switch status {
	case saleDomain.PaymentPending, saleDomain.PaymentApproved, saleDomain.PaymentCanceled:
		return true
	default:
		return false
	}
}
