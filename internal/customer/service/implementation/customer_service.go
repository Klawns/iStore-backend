package implementation

import (
	"istore/internal/customer/domain"
	repoContracts "istore/internal/customer/repository/contracts"
	serviceContracts "istore/internal/customer/service/contracts"
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

func (s *CustomerService) Create(input serviceContracts.CreateCustomerInput) *rest_err.RestErr {
	// Primeiro, criamos o domínio de cliente com os dados recebidos.
	customer := &domain.Customer{
		Name:  input.Name,
		Phone: input.Phone,
	}

	// Segundo, chamamos o repository para persistir o cliente.
	err := s.CustomerRepository.Create(customer)
	if err != nil {
		// Terceiro, erros técnicos viram log e erro interno da API.
		logger.Error("Error trying to create an customer", err, zap.String("customer_name", input.Name), zap.String("journey", "CreateCustomer"))
		return rest_err.NewInternalServerError("Erro ao criar cliente")
	}

	return nil
}

func (s *CustomerService) Update(id int, input serviceContracts.UpdateCustomerInput) *rest_err.RestErr {
	// Primeiro, buscamos o cliente atual para validar existência e preservar campos não enviados.
	customer, err := s.CustomerRepository.FindByID(id)
	if err != nil {
		logger.Error("Error trying to find an customer", err, zap.Int("customer_id", id), zap.String("journey", "UpdateCustomer"))
		return rest_err.NewInternalServerError("Erro ao atualizar cliente")
	}
	// Segundo, se o repository não encontrou o cliente, retornamos 404.
	if customer == nil {
		return rest_err.NewNotFoundError("Cliente não encontrado")
	}

	// Terceiro, aplicamos update parcial: campo vazio significa não atualizar.
	if input.Name != "" {
		customer.Name = input.Name
	}
	if input.Phone != "" {
		customer.Phone = input.Phone
	}

	// Quarto, salvamos o cliente atualizado pelo repository.
	err = s.CustomerRepository.Update(customer)
	if err != nil {
		logger.Error("Error trying to update an customer", err, zap.Int("customer_id", id), zap.String("journey", "UpdateCustomer"))
		return rest_err.NewInternalServerError("Erro ao atualizar cliente")
	}

	return nil
}

func (s *CustomerService) Delete(id int) *rest_err.RestErr {
	// Primeiro, buscamos o cliente para garantir que ele existe.
	customer, err := s.CustomerRepository.FindByID(id)
	if err != nil {
		logger.Error("Error trying to find an customer", err, zap.Int("customer_id", id), zap.String("journey", "DeleteCustomer"))
		return rest_err.NewInternalServerError("Erro ao deletar cliente")
	}
	// Segundo, se não existir, retornamos 404.
	if customer == nil {
		return rest_err.NewNotFoundError("Cliente não encontrado")
	}

	// Terceiro, deletamos o cliente pelo repository.
	err = s.CustomerRepository.Delete(id)
	if err != nil {
		logger.Error("Error trying to delete an customer", err, zap.Int("customer_id", id), zap.String("journey", "DeleteCustomer"))
		return rest_err.NewInternalServerError("Erro ao deletar cliente")
	}

	return nil
}

func (s *CustomerService) GetByID(id int) (*domain.Customer, *rest_err.RestErr) {
	// Primeiro, buscamos o cliente pelo ID usando o repository.
	customer, err := s.CustomerRepository.FindByID(id)
	if err != nil {
		logger.Error("Error trying to find an customer", err, zap.Int("customer_id", id), zap.String("journey", "FindCustomerByID"))
		return nil, rest_err.NewInternalServerError("Erro ao buscar cliente")
	}
	// Segundo, se não encontrar, retornamos erro 404 para a API.
	if customer == nil {
		return nil, rest_err.NewNotFoundError("Cliente não encontrado")
	}

	return customer, nil
}

func (s *CustomerService) List() ([]domain.Customer, *rest_err.RestErr) {
	// Primeiro, buscamos todos os clientes usando o repository.
	customers, err := s.CustomerRepository.FindAll()
	if err != nil {
		// Segundo, erros técnicos viram log e erro interno da API.
		logger.Error("Error trying to list customers", err, zap.String("journey", "ListCustomers"))
		return nil, rest_err.NewInternalServerError("Erro ao listar clientes")
	}
	return customers, nil
}
