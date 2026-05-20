package handler

import (
	"istore/internal/customer/domain"
	"istore/internal/customer/dto/request"
	"istore/internal/customer/dto/response"
	serviceContracts "istore/internal/customer/service/contracts"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Essa struct representa o handler do customer e depende do service para aplicar as regras da aplicação.
type CustomerHandler struct {
	service serviceContracts.CustomerService
}

// Essa função é o construtor do handler, recebendo o service como dependência.
func NewCustomerHandler(service serviceContracts.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) Create(ctx *gin.Context) {
	// Primeiro, fazemos o bind do JSON recebido para a struct de request.
	var req request.CustomerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	// Segundo, chamamos o service para criar o cliente e tratar regras/erros.
	restErr := h.service.Create(serviceContracts.CreateCustomerInput{
		Name:  req.Name,
		Phone: req.Phone,
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	// Terceiro, devolvemos os dados criados no formato de response.
	ctx.JSON(http.StatusCreated, response.FromDomain(&domain.Customer{
		Name:  req.Name,
		Phone: req.Phone,
	}))
}

func (h *CustomerHandler) Update(ctx *gin.Context) {
	// Primeiro, extraímos o ID da URL.
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	// Segundo, fazemos o bind do JSON recebido para a struct de request.
	var req request.CustomerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	// Terceiro, chamamos o service para atualizar com regra de update parcial.
	restErr = h.service.Update(id, serviceContracts.UpdateCustomerInput{
		Name:  req.Name,
		Phone: req.Phone,
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *CustomerHandler) Delete(ctx *gin.Context) {
	// Primeiro, extraímos o ID da URL.
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	// Segundo, chamamos o service para validar existência e deletar.
	restErr = h.service.Delete(id)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *CustomerHandler) GetByID(ctx *gin.Context) {
	// Primeiro, extraímos o ID da URL.
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	// Segundo, buscamos o cliente pelo service.
	customer, restErr := h.service.GetByID(id)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromDomain(customer))
}

func (h *CustomerHandler) List(ctx *gin.Context) {
	// Primeiro, chamamos o service para listar os clientes.
	customers, restErr := h.service.List()
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	// Segundo, convertemos a lista para o formato de response.
	responses := make([]response.CustomerResponse, len(customers))
	for i, customer := range customers {
		responses[i] = *response.FromDomain(&customer)
	}

	ctx.JSON(http.StatusOK, responses)
}

func getIDParam(ctx *gin.Context) (int, *rest_err.RestErr) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return 0, rest_err.NewBadRequestError("ID inválido")
	}

	return id, nil
}
