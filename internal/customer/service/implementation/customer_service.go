package implementation

import (
	"istore/internal/customer/domain"
	repoContracts "istore/internal/customer/repository/contracts"
	serviceContracts "istore/internal/customer/service/contracts"
	saleDomain "istore/internal/sale/domain"
	"istore/pkg/logger"
	"istore/pkg/rest_err"

	"go.uber.org/zap"
)

type CustomerService struct {
	CustomerRepository repoContracts.CustomerRepository
}

func NewCustomerService(customerRepository repoContracts.CustomerRepository) *CustomerService {
	return &CustomerService{
		CustomerRepository: customerRepository,
	}
}

func (s *CustomerService) Create(input serviceContracts.CreateCustomerInput) (*domain.Customer, *rest_err.RestErr) {
	if input.UserID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	customer := &domain.Customer{
		UserID: input.UserID,
		Name:   input.Name,
		Phone:  input.Phone,
	}

	err := s.CustomerRepository.Create(customer)
	if err != nil {
		logger.Error("Error trying to create an customer", err, zap.Uint("user_id", input.UserID), zap.String("journey", "CreateCustomer"))
		return nil, rest_err.NewInternalServerError("Erro ao criar cliente")
	}

	return customer, nil
}

func (s *CustomerService) Update(id int, input serviceContracts.UpdateCustomerInput) (*domain.Customer, *rest_err.RestErr) {
	if input.UserID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	customer, err := s.CustomerRepository.FindByID(input.UserID, id)
	if err != nil {
		logger.Error("Error trying to find an customer", err, zap.Int("customer_id", id), zap.String("journey", "UpdateCustomer"))
		return nil, rest_err.NewInternalServerError("Erro ao atualizar cliente")
	}
	if customer == nil {
		return nil, rest_err.NewNotFoundError("Cliente não encontrado")
	}

	if input.Name != "" {
		customer.Name = input.Name
	}
	if input.Phone != "" {
		customer.Phone = input.Phone
	}

	err = s.CustomerRepository.Update(customer)
	if err != nil {
		logger.Error("Error trying to update an customer", err, zap.Int("customer_id", id), zap.String("journey", "UpdateCustomer"))
		return nil, rest_err.NewInternalServerError("Erro ao atualizar cliente")
	}

	return customer, nil
}

func (s *CustomerService) Delete(userID uint, id int) *rest_err.RestErr {
	if userID == 0 {
		return rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	customer, err := s.CustomerRepository.FindByID(userID, id)
	if err != nil {
		logger.Error("Error trying to find an customer", err, zap.Int("customer_id", id), zap.String("journey", "DeleteCustomer"))
		return rest_err.NewInternalServerError("Erro ao deletar cliente")
	}
	if customer == nil {
		return rest_err.NewNotFoundError("Cliente não encontrado")
	}

	salesCount, err := s.CustomerRepository.CountSalesByCustomerIDs(userID, []int{id})
	if err != nil {
		logger.Error("Error trying to count customer sales", err, zap.Int("customer_id", id), zap.String("journey", "DeleteCustomer"))
		return rest_err.NewInternalServerError("Erro ao deletar cliente")
	}
	if salesCount > 0 {
		return rest_err.NewConflictError("Clientes com historico de vendas nao podem ser excluidos")
	}

	err = s.CustomerRepository.Delete(userID, id)
	if err != nil {
		logger.Error("Error trying to delete an customer", err, zap.Int("customer_id", id), zap.String("journey", "DeleteCustomer"))
		return rest_err.NewInternalServerError("Erro ao deletar cliente")
	}

	return nil
}

func (s *CustomerService) DeleteMany(userID uint, ids []int) (int, *rest_err.RestErr) {
	if userID == 0 {
		return 0, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	uniqueIDs := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return 0, rest_err.NewBadRequestError("IDs invalidos")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return 0, rest_err.NewBadRequestError("Informe ao menos um cliente")
	}

	customers, err := s.CustomerRepository.FindByIDs(userID, uniqueIDs)
	if err != nil {
		logger.Error("Error trying to find customers", err, zap.Uint("user_id", userID), zap.String("journey", "DeleteManyCustomers"))
		return 0, rest_err.NewInternalServerError("Erro ao deletar clientes")
	}
	if len(customers) != len(uniqueIDs) {
		return 0, rest_err.NewNotFoundError("Cliente não encontrado")
	}

	salesCount, err := s.CustomerRepository.CountSalesByCustomerIDs(userID, uniqueIDs)
	if err != nil {
		logger.Error("Error trying to count customer sales", err, zap.Uint("user_id", userID), zap.String("journey", "DeleteManyCustomers"))
		return 0, rest_err.NewInternalServerError("Erro ao deletar clientes")
	}
	if salesCount > 0 {
		return 0, rest_err.NewConflictError("Clientes com historico de vendas nao podem ser excluidos")
	}

	if err := s.CustomerRepository.DeleteMany(userID, uniqueIDs); err != nil {
		logger.Error("Error trying to delete customers", err, zap.Uint("user_id", userID), zap.String("journey", "DeleteManyCustomers"))
		return 0, rest_err.NewInternalServerError("Erro ao deletar clientes")
	}

	return len(uniqueIDs), nil
}

func (s *CustomerService) GetByID(userID uint, id int) (*domain.Customer, *rest_err.RestErr) {
	if userID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	customer, err := s.CustomerRepository.FindByID(userID, id)
	if err != nil {
		logger.Error("Error trying to find an customer", err, zap.Int("customer_id", id), zap.String("journey", "FindCustomerByID"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar cliente")
	}
	if customer == nil {
		return nil, rest_err.NewNotFoundError("Cliente não encontrado")
	}

	return customer, nil
}

func (s *CustomerService) List(input serviceContracts.ListCustomersInput) (*domain.CustomerListResult, *rest_err.RestErr) {
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

	result, err := s.CustomerRepository.List(domain.CustomerListFilter{
		UserID:        input.UserID,
		Page:          input.Page,
		Limit:         input.Limit,
		Start:         input.Start,
		End:           input.End,
		PaymentStatus: input.Status,
		PaymentType:   input.PaymentType,
		Search:        input.Search,
	})
	if err != nil {
		logger.Error("Error trying to list customers", err, zap.String("journey", "ListCustomers"))
		return nil, rest_err.NewInternalServerError("Erro ao listar clientes")
	}
	return result, nil
}

func isValidPaymentStatus(status saleDomain.PaymentStatus) bool {
	switch status {
	case saleDomain.PaymentPending, saleDomain.PaymentApproved, saleDomain.PaymentCanceled:
		return true
	default:
		return false
	}
}

func isValidPaymentType(paymentType saleDomain.PaymentType) bool {
	switch paymentType {
	case saleDomain.Pix, saleDomain.Money, saleDomain.CreditCard, saleDomain.DebitCard:
		return true
	default:
		return false
	}
}
