package handler

import (
	"istore/internal/sale/dto/request"
	"istore/internal/sale/dto/response"
	"istore/internal/sale/service/contract"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SaleHandler struct {
	service contract.SaleService
}

func NewSaleHandler(service contract.SaleService) *SaleHandler {
	return &SaleHandler{service: service}
}

func (h *SaleHandler) Create(ctx *gin.Context) {
	var req request.SaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	sale, restErr := h.service.Create(&contract.CreateSaleInput{
		ClienteID:       req.ClienteID,
		TipoPagamento:   req.TipoPagamento,
		StatusPagamento: req.StatusPagamento,
		SaleDate:        req.SaleDate,
		Itens:           toCreateSaleItems(req.Itens),
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusCreated, response.FromDomain(sale))
}

func (h *SaleHandler) GetByID(ctx *gin.Context) {
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	sale, restErr := h.service.GetByID(id)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromDomain(sale))
}

func (h *SaleHandler) List(ctx *gin.Context) {
	sales, restErr := h.service.List()
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	responses := make([]response.SaleResponse, len(sales))
	for i, sale := range sales {
		responses[i] = *response.FromDomain(&sale)
	}

	ctx.JSON(http.StatusOK, responses)
}

func (h *SaleHandler) ListByPeriod(ctx *gin.Context) {
	start, restErr := parseTimeQuery(ctx.Query("start"))
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	end, restErr := parseTimeQuery(ctx.Query("end"))
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	sales, restErr := h.service.ListByPeriod(start, end)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	responses := make([]response.SaleResponse, len(sales))
	for i, sale := range sales {
		responses[i] = *response.FromDomain(&sale)
	}

	ctx.JSON(http.StatusOK, responses)
}

func (h *SaleHandler) UpdateStatus(ctx *gin.Context) {
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	var req request.SaleStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	restErr = h.service.UpdateStatus(id, req.Status)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *SaleHandler) Delete(ctx *gin.Context) {
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	restErr = h.service.Delete(id)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func getIDParam(ctx *gin.Context) (int, *rest_err.RestErr) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return 0, rest_err.NewBadRequestError("ID inválido")
	}

	return id, nil
}

func parseTimeQuery(value string) (time.Time, *rest_err.RestErr) {
	if value == "" {
		return time.Time{}, rest_err.NewBadRequestError("Parâmetro de data obrigatório")
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, rest_err.NewBadRequestError("Formato de data inválido")
	}

	return parsed, nil
}

func toCreateSaleItems(items []request.SaleItemRequest) []contract.CreateSaleItemInput {
	result := make([]contract.CreateSaleItemInput, len(items))
	for i, item := range items {
		result[i] = contract.CreateSaleItemInput{
			ProductName: item.ProductName,
			Specs:       item.Specs,
			Quantity:    item.Quantity,
			CostPrice:   item.CostPrice,
			SalePrice:   item.SalePrice,
		}
	}

	return result
}
