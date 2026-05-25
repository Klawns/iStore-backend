package implementation

import (
	"errors"
	"istore/internal/sale/domain"
	"istore/internal/sale/repository/contracts"
	"istore/internal/sale/service/contract"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SaleService struct {
	repository contracts.SaleRepository
}

func NewSaleService(repository contracts.SaleRepository) contract.SaleService {
	return &SaleService{repository: repository}
}

func (s *SaleService) Create(input *contract.CreateSaleInput) (*domain.Sale, *rest_err.RestErr) {
	if input == nil {
		return nil, rest_err.NewBadRequestError("Dados da venda são obrigatórios")
	}

	if input.UserID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if input.ClienteID <= 0 {
		return nil, rest_err.NewBadRequestError("Cliente inválido")
	}

	if !isValidPaymentType(input.TipoPagamento) {
		return nil, rest_err.NewBadRequestError("Tipo de pagamento inválido")
	}

	if !isValidPaymentStatus(input.StatusPagamento) {
		return nil, rest_err.NewBadRequestError("Status de pagamento inválido")
	}

	if restErr := normalizeCardFields(input); restErr != nil {
		return nil, restErr
	}

	if len(input.Itens) == 0 {
		return nil, rest_err.NewBadRequestError("Venda deve possuir ao menos um item")
	}

	sale := &domain.Sale{
		UserID:        input.UserID,
		CustomerID:    int(input.ClienteID),
		PaymentType:   input.TipoPagamento,
		PaymentStatus: input.StatusPagamento,
		SaleDate:      input.SaleDate,
		Installments:  input.Installments,
		BillingDay:    input.BillingDay,
		Items:         make([]domain.SaleItem, len(input.Itens)),
	}

	if sale.SaleDate.IsZero() {
		sale.SaleDate = time.Now()
	}

	for i, item := range input.Itens {
		if item.ProductName == "" {
			return nil, rest_err.NewBadRequestError("Nome do produto é obrigatório")
		}
		if item.Quantity <= 0 {
			return nil, rest_err.NewBadRequestError("Quantidade do item deve ser maior que zero")
		}
		if item.CostPrice < 0 {
			return nil, rest_err.NewBadRequestError("Preço de custo não pode ser negativo")
		}
		if item.SalePrice < 0 {
			return nil, rest_err.NewBadRequestError("Preço de venda não pode ser negativo")
		}

		sale.Items[i] = domain.SaleItem{
			ProductName: item.ProductName,
			Specs:       item.Specs,
			Quantity:    item.Quantity,
			CostPrice:   item.CostPrice,
			SalePrice:   item.SalePrice,
		}
	}

	sale.TotalValue = sale.CalculateTotal()
	sale.InstallmentsList = buildSaleInstallments(sale)

	err := s.repository.Create(sale)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, rest_err.NewBadRequestError("Cliente invalido")
		}
		logger.Error("Erro ao criar venda: ", err, zap.String("customer_id", strconv.Itoa(int(input.ClienteID))), zap.String("payment_type", string(input.TipoPagamento)), zap.String("payment_status", string(input.StatusPagamento)), zap.String("journey", "CreateSale"))
		return nil, rest_err.NewInternalServerError("Erro ao criar venda")
	}
	return sale, nil

}

func (s *SaleService) GetByID(userID uint, id int) (*domain.Sale, *rest_err.RestErr) {
	if userID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if id <= 0 {
		return nil, rest_err.NewBadRequestError("ID inválido")
	}

	// Primeiro, precisamos usar o repositório para buscar a venda pelo ID
	sale, err := s.repository.FindByID(userID, id)
	if err != nil {
		logger.Error("Erro ao buscar venda por ID: ", err, zap.Int("sale_id", id), zap.String("journey", "GetSaleByID"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar venda por ID")
	}
	if sale == nil {
		return nil, rest_err.NewNotFoundError("Venda não encontrada")
	}
	return sale, nil
}

func (s *SaleService) List(input contract.ListSalesInput) (*domain.SaleListResult, *rest_err.RestErr) {
	if input.UserID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if input.Page <= 0 {
		return nil, rest_err.NewBadRequestError("Pagina invalida")
	}

	if input.Limit <= 0 || input.Limit > 100 {
		return nil, rest_err.NewBadRequestError("Limite invalido")
	}

	if input.Start != nil && input.Start.IsZero() {
		return nil, rest_err.NewBadRequestError("Data de inicio invalida")
	}

	if input.End != nil && input.End.IsZero() {
		return nil, rest_err.NewBadRequestError("Data de fim invalida")
	}

	if input.Start != nil && input.End != nil && input.End.Before(*input.Start) {
		return nil, rest_err.NewBadRequestError("Data de termino deve ser posterior a data de inicio")
	}

	if input.Status != nil && !isValidPaymentStatus(*input.Status) {
		return nil, rest_err.NewBadRequestError("Status de pagamento invalido")
	}

	if input.PaymentType != nil && !isValidPaymentType(*input.PaymentType) {
		return nil, rest_err.NewBadRequestError("Tipo de pagamento invalido")
	}

	if input.CustomerID != nil && *input.CustomerID <= 0 {
		return nil, rest_err.NewBadRequestError("Cliente invalido")
	}

	result, err := s.repository.List(domain.SaleListFilter{
		Page:          input.Page,
		UserID:        input.UserID,
		Limit:         input.Limit,
		Start:         input.Start,
		End:           input.End,
		PaymentStatus: input.Status,
		PaymentType:   input.PaymentType,
		CustomerID:    input.CustomerID,
		Search:        input.Search,
	})
	if err != nil {
		logger.Error("Erro ao listar vendas: ", err, zap.String("journey", "ListSales"))
		return nil, rest_err.NewInternalServerError("Erro ao listar vendas")
	}

	return result, nil
}

func (s *SaleService) ListByPeriod(userID uint, start time.Time, end time.Time) ([]domain.Sale, *rest_err.RestErr) {
	if userID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if start.IsZero() {
		return nil, rest_err.NewBadRequestError("Data de início não pode ser zero")
	}

	if end.IsZero() {
		return nil, rest_err.NewBadRequestError("Data de fim não pode ser zero")
	}

	if end.Before(start) {
		return nil, rest_err.NewBadRequestError("Data de término deve ser posterior à data de início")
	}

	sales, err := s.repository.ListByPeriod(userID, start, end)
	if err != nil {
		logger.Error("Erro ao listar vendas por período: ", err, zap.Time("start", start), zap.Time("end", end), zap.String("journey", "ListSalesByPeriod"))
		return nil, rest_err.NewInternalServerError("Erro ao listar vendas por período")
	}
	return sales, nil
}

func (s *SaleService) UpdateStatus(userID uint, id int, status domain.PaymentStatus) *rest_err.RestErr {
	if userID == 0 {
		return rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if id <= 0 {
		return rest_err.NewBadRequestError("ID inválido")
	}

	if !isValidPaymentStatus(status) {
		return rest_err.NewBadRequestError("Status de pagamento inválido")
	}

	sale, err := s.repository.FindByID(userID, id)
	if err != nil {
		logger.Error("Erro ao buscar venda: ", err, zap.String("journey", "UpdateStatus"))
		return rest_err.NewInternalServerError("Erro ao buscar venda")
	}
	if sale == nil {
		return rest_err.NewNotFoundError("Venda não encontrada")
	}

	err = s.repository.UpdateStatus(userID, id, status)
	if err != nil {
		logger.Error("Erro ao atualizar status da venda: ", err, zap.Int("sale_id", id), zap.String("payment_status", string(status)), zap.String("journey", "UpdateSaleStatus"))
		return rest_err.NewInternalServerError("Erro ao atualizar status da venda")
	}
	return nil
}

func (s *SaleService) Delete(userID uint, id int) *rest_err.RestErr {
	if userID == 0 {
		return rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if id <= 0 {
		return rest_err.NewBadRequestError("ID inválido")
	}

	sale, err := s.repository.FindByID(userID, id)
	if err != nil {
		logger.Error("Erro ao buscar venda: ", err, zap.Int("sale_id", id), zap.String("journey", "DeleteSale"))
		return rest_err.NewInternalServerError("Erro ao deletar venda")
	}
	if sale == nil {
		return rest_err.NewNotFoundError("Venda não encontrada")
	}

	err = s.repository.Delete(userID, id)
	if err != nil {
		logger.Error("Erro ao deletar venda: ", err, zap.Int("sale_id", id), zap.String("journey", "DeleteSale"))
		return rest_err.NewInternalServerError("Erro ao deletar venda")
	}
	return nil
}

func (s *SaleService) ListInstallmentAlerts(userID uint, now time.Time, windowDays int) ([]domain.SaleInstallment, *rest_err.RestErr) {
	if userID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if windowDays <= 0 {
		windowDays = 7
	}

	installments, err := s.repository.ListInstallmentAlerts(userID, now, windowDays)
	if err != nil {
		logger.Error("Erro ao listar parcelas para acompanhamento: ", err, zap.String("journey", "ListInstallmentAlerts"))
		return nil, rest_err.NewInternalServerError("Erro ao listar parcelas")
	}

	return installments, nil
}

func (s *SaleService) ListInstallmentsBySaleID(userID uint, saleID int) ([]domain.SaleInstallment, *rest_err.RestErr) {
	if userID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if saleID <= 0 {
		return nil, rest_err.NewBadRequestError("ID inválido")
	}

	installments, err := s.repository.ListInstallmentsBySaleID(userID, saleID)
	if err != nil {
		logger.Error("Erro ao listar parcelas da venda: ", err, zap.Int("sale_id", saleID), zap.String("journey", "ListInstallmentsBySaleID"))
		return nil, rest_err.NewInternalServerError("Erro ao listar parcelas")
	}

	return installments, nil
}

func (s *SaleService) UpdateInstallmentStatus(userID uint, id int, input contract.UpdateInstallmentStatusInput) (*domain.SaleInstallment, *rest_err.RestErr) {
	if userID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	if id <= 0 {
		return nil, rest_err.NewBadRequestError("ID inválido")
	}

	if !isValidInstallmentStatus(input.Status) || input.Status == domain.InstallmentPending {
		return nil, rest_err.NewBadRequestError("Status da parcela inválido")
	}

	installment, err := s.repository.UpdateInstallmentStatus(userID, id, input.Status, input.Notes, time.Now())
	if err != nil {
		logger.Error("Erro ao atualizar parcela: ", err, zap.Int("installment_id", id), zap.String("status", string(input.Status)), zap.String("journey", "UpdateInstallmentStatus"))
		return nil, rest_err.NewInternalServerError("Erro ao atualizar parcela")
	}
	if installment == nil {
		return nil, rest_err.NewNotFoundError("Parcela não encontrada")
	}

	return installment, nil
}

func isValidPaymentStatus(status domain.PaymentStatus) bool {
	switch status {
	case domain.PaymentPending, domain.PaymentApproved, domain.PaymentCanceled:
		return true
	default:
		return false
	}
}

func isValidPaymentType(paymentType domain.PaymentType) bool {
	switch paymentType {
	case domain.Pix, domain.Money, domain.CreditCard, domain.DebitCard:
		return true
	default:
		return false
	}
}

func isValidInstallmentStatus(status domain.SaleInstallmentStatus) bool {
	switch status {
	case domain.InstallmentPending, domain.InstallmentPaid, domain.InstallmentUnpaid:
		return true
	default:
		return false
	}
}

func normalizeCardFields(input *contract.CreateSaleInput) *rest_err.RestErr {
	hasInstallments := input.Installments != nil
	hasBillingDay := input.BillingDay != nil

	switch input.TipoPagamento {
	case domain.DebitCard:
		if hasInstallments || hasBillingDay {
			return rest_err.NewBadRequestError("Debito nao aceita parcelas ou dia de cobranca")
		}
		input.StatusPagamento = domain.PaymentApproved
	case domain.CreditCard:
		if !hasInstallments || *input.Installments < 1 || *input.Installments > 24 {
			return rest_err.NewBadRequestError("Credito exige parcelas entre 1 e 24")
		}
		if !hasBillingDay || *input.BillingDay < 1 || *input.BillingDay > 31 {
			return rest_err.NewBadRequestError("Credito exige dia de cobranca entre 1 e 31")
		}
	case domain.Pix, domain.Money:
		if hasInstallments || hasBillingDay {
			return rest_err.NewBadRequestError("PIX e dinheiro nao aceitam parcelas ou dia de cobranca")
		}
	}

	return nil
}

func creditCardDueDates(sale domain.Sale) []time.Time {
	if sale.Installments == nil || sale.BillingDay == nil || *sale.Installments <= 0 {
		return nil
	}

	first := firstBillingDate(sale.SaleDate, *sale.BillingDay)
	dates := make([]time.Time, *sale.Installments)
	for i := 0; i < *sale.Installments; i++ {
		monthIndex := int(first.Month()) - 1 + i
		year := first.Year() + monthIndex/12
		month := time.Month(monthIndex%12 + 1)
		dates[i] = dateWithClampedDay(year, month, *sale.BillingDay, first.Location())
	}

	return dates
}

func buildSaleInstallments(sale *domain.Sale) []domain.SaleInstallment {
	if sale == nil || sale.PaymentType != domain.CreditCard || sale.Installments == nil || *sale.Installments <= 0 {
		return nil
	}

	dueDates := creditCardDueDates(*sale)
	installments := make([]domain.SaleInstallment, len(dueDates))
	baseAmount := sale.TotalValue / len(dueDates)
	remainder := sale.TotalValue % len(dueDates)

	for i, dueDate := range dueDates {
		amount := baseAmount
		if i < remainder {
			amount++
		}

		installments[i] = domain.SaleInstallment{
			DueDate:           dueDate,
			InstallmentNumber: i + 1,
			TotalInstallments: len(dueDates),
			Amount:            amount,
			Status:            domain.InstallmentPending,
		}
	}

	return installments
}

func firstBillingDate(saleDate time.Time, billingDay int) time.Time {
	candidate := dateWithClampedDay(saleDate.Year(), saleDate.Month(), billingDay, saleDate.Location())
	if !candidate.Before(dateOnly(saleDate)) {
		return candidate
	}

	nextMonth := saleDate.AddDate(0, 1, 0)
	return dateWithClampedDay(nextMonth.Year(), nextMonth.Month(), billingDay, saleDate.Location())
}

func dateWithClampedDay(year int, month time.Month, day int, location *time.Location) time.Time {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, location).Day()
	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
