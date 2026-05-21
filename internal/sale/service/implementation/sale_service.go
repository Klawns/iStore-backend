package implementation

import (
	"istore/internal/sale/domain"
	"istore/internal/sale/repository/contracts"
	"istore/internal/sale/service/contract"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"strconv"
	"time"

	"go.uber.org/zap"
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

	if input.ClienteID <= 0 {
		return nil, rest_err.NewBadRequestError("Cliente inválido")
	}

	if !isValidPaymentType(input.TipoPagamento) {
		return nil, rest_err.NewBadRequestError("Tipo de pagamento inválido")
	}

	if !isValidPaymentStatus(input.StatusPagamento) {
		return nil, rest_err.NewBadRequestError("Status de pagamento inválido")
	}

	if len(input.Itens) == 0 {
		return nil, rest_err.NewBadRequestError("Venda deve possuir ao menos um item")
	}

	sale := &domain.Sale{
		CustomerID:    int(input.ClienteID),
		PaymentType:   input.TipoPagamento,
		PaymentStatus: input.StatusPagamento,
		SaleDate:      input.SaleDate,
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

	// Agora, podemos usar o repositório para salvar a venda no banco de dados
	err := s.repository.Create(sale)
	if err != nil {
		logger.Error("Erro ao criar venda: ", err, zap.String("customer_id", strconv.Itoa(int(input.ClienteID))), zap.String("payment_type", string(input.TipoPagamento)), zap.String("payment_status", string(input.StatusPagamento)), zap.String("journey", "CreateSale"))
		return nil, rest_err.NewInternalServerError("Erro ao criar venda")
	}
	return sale, nil

}

func (s *SaleService) GetByID(id int) (*domain.Sale, *rest_err.RestErr) {
	if id <= 0 {
		return nil, rest_err.NewBadRequestError("ID inválido")
	}

	// Primeiro, precisamos usar o repositório para buscar a venda pelo ID
	sale, err := s.repository.FindByID(id)
	if err != nil {
		logger.Error("Erro ao buscar venda por ID: ", err, zap.Int("sale_id", id), zap.String("journey", "GetSaleByID"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar venda por ID")
	}
	if sale == nil {
		return nil, rest_err.NewNotFoundError("Venda não encontrada")
	}
	return sale, nil
}

func (s *SaleService) List() ([]domain.Sale, *rest_err.RestErr) {
	// Primeiro, precisamos usar o repositório para buscar todas as vendas
	sales, err := s.repository.FindAll()
	if err != nil {
		logger.Error("Erro ao listar vendas: ", err, zap.String("journey", "ListSales"))
		return nil, rest_err.NewInternalServerError("Erro ao listar vendas")
	}

	return sales, nil
}

func (s *SaleService) ListByPeriod(start time.Time, end time.Time) ([]domain.Sale, *rest_err.RestErr) {
	if start.IsZero() {
		return nil, rest_err.NewBadRequestError("Data de início não pode ser zero")
	}

	if end.IsZero() {
		return nil, rest_err.NewBadRequestError("Data de fim não pode ser zero")
	}

	if end.Before(start) {
		return nil, rest_err.NewBadRequestError("Data de término deve ser posterior à data de início")
	}

	sales, err := s.repository.ListByPeriod(start, end)
	if err != nil {
		logger.Error("Erro ao listar vendas por período: ", err, zap.Time("start", start), zap.Time("end", end), zap.String("journey", "ListSalesByPeriod"))
		return nil, rest_err.NewInternalServerError("Erro ao listar vendas por período")
	}
	return sales, nil
}

func (s *SaleService) UpdateStatus(id int, status domain.PaymentStatus) *rest_err.RestErr {
	if id <= 0 {
		return rest_err.NewBadRequestError("ID inválido")
	}

	if !isValidPaymentStatus(status) {
		return rest_err.NewBadRequestError("Status de pagamento inválido")
	}

	sale, err := s.repository.FindByID(id)
	if err != nil {
		logger.Error("Erro ao buscar venda: ", err, zap.String("journey", "UpdateStatus"))
		return rest_err.NewInternalServerError("Erro ao buscar venda")
	}
	if sale == nil {
		return rest_err.NewNotFoundError("Venda não encontrada")
	}

	err = s.repository.UpdateStatus(id, status)
	if err != nil {
		logger.Error("Erro ao atualizar status da venda: ", err, zap.Int("sale_id", id), zap.String("payment_status", string(status)), zap.String("journey", "UpdateSaleStatus"))
		return rest_err.NewInternalServerError("Erro ao atualizar status da venda")
	}
	return nil
}

func (s *SaleService) Delete(id int) *rest_err.RestErr {
	if id <= 0 {
		return rest_err.NewBadRequestError("ID inválido")
	}

	sale, err := s.repository.FindByID(id)
	if err != nil {
		logger.Error("Erro ao buscar venda: ", err, zap.Int("sale_id", id), zap.String("journey", "DeleteSale"))
		return rest_err.NewInternalServerError("Erro ao deletar venda")
	}
	if sale == nil {
		return rest_err.NewNotFoundError("Venda não encontrada")
	}

	err = s.repository.Delete(id)
	if err != nil {
		logger.Error("Erro ao deletar venda: ", err, zap.Int("sale_id", id), zap.String("journey", "DeleteSale"))
		return rest_err.NewInternalServerError("Erro ao deletar venda")
	}
	return nil
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
	case domain.Pix, domain.Money, domain.Card:
		return true
	default:
		return false
	}
}
