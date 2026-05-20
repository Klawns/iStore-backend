package implementation

import (
	"errors"
	"istore/internal/customer/domain"
	"istore/internal/customer/repository/contracts"
	"istore/internal/customer/repository/entity"

	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) contracts.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(customer *domain.Customer) error {
	// Primeiro, convertemos o domínio para a entidade do banco de dados.
	customerEntity := entity.FromDomain(customer)

	// Segundo, salvamos a entidade usando o GORM e retornamos o erro bruto, se houver.
	if err := r.db.Create(customerEntity).Error; err != nil {
		return err
	}

	// Terceiro, devolvemos o ID gerado para o domínio recebido.
	if customerEntity != nil {
		customer.ID = customerEntity.ID
	}

	return nil
}

func (r *customerRepository) Update(customer *domain.Customer) error {
	// Primeiro, convertemos o domínio já validado pelo service para entidade.
	customerEntity := entity.FromDomain(customer)

	// Segundo, salvamos a entidade no banco e retornamos apenas o erro do GORM.
	return r.db.Save(customerEntity).Error
}

func (r *customerRepository) Delete(id int) error {
	// Deletamos pelo ID e retornamos apenas o erro do GORM.
	return r.db.Delete(&entity.CustomerEntity{}, id).Error
}

func (r *customerRepository) FindByID(id int) (*domain.Customer, error) {
	// Primeiro, buscamos o cliente no banco usando o ID recebido.
	var customerEntity entity.CustomerEntity
	if err := r.db.First(&customerEntity, id).Error; err != nil {
		// Quando não encontrar registro, retornamos nil sem erro para o service decidir o 404.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	// Segundo, convertemos a entidade encontrada para domínio.
	return customerEntity.ToDomain(), nil
}

func (r *customerRepository) FindAll() ([]domain.Customer, error) {
	// Primeiro, buscamos todos os clientes cadastrados no banco.
	var customerEntities []entity.CustomerEntity
	if err := r.db.Find(&customerEntities).Error; err != nil {
		return nil, err
	}

	// Segundo, convertemos a lista de entidades para domínios.
	customers := make([]domain.Customer, len(customerEntities))
	for i, customerEntity := range customerEntities {
		customers[i] = *customerEntity.ToDomain()
	}

	return customers, nil
}
