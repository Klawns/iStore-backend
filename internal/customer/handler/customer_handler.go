package handler

import (
	"istore/internal/customer/dto/request"
	"istore/internal/customer/dto/response"
	serviceContracts "istore/internal/customer/service/contracts"
	saleDomain "istore/internal/sale/domain"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	service serviceContracts.CustomerService
}

func NewCustomerHandler(service serviceContracts.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) Create(ctx *gin.Context) {
	var req request.CustomerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	customer, restErr := h.service.Create(serviceContracts.CreateCustomerInput{
		Name:  req.Name,
		Phone: req.Phone,
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusCreated, response.FromDomain(customer))
}

func (h *CustomerHandler) Update(ctx *gin.Context) {
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	var req request.CustomerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	customer, restErr := h.service.Update(id, serviceContracts.UpdateCustomerInput{
		Name:  req.Name,
		Phone: req.Phone,
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromDomain(customer))
}

func (h *CustomerHandler) Delete(ctx *gin.Context) {
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

func (h *CustomerHandler) GetByID(ctx *gin.Context) {
	id, restErr := getIDParam(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	customer, restErr := h.service.GetByID(id)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromDomain(customer))
}

func (h *CustomerHandler) List(ctx *gin.Context) {
	input, restErr := parseListCustomersQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	result, restErr := h.service.List(input)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.ListFromDomain(result))
}

func getIDParam(ctx *gin.Context) (int, *rest_err.RestErr) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return 0, rest_err.NewBadRequestError("ID inválido")
	}

	return id, nil
}

func parseListCustomersQuery(ctx *gin.Context) (serviceContracts.ListCustomersInput, *rest_err.RestErr) {
	page, restErr := parsePositiveIntQuery(ctx.Query("page"), 1, "Pagina invalida")
	if restErr != nil {
		return serviceContracts.ListCustomersInput{}, restErr
	}

	limit, restErr := parsePositiveIntQuery(ctx.Query("limit"), 10, "Limite invalido")
	if restErr != nil {
		return serviceContracts.ListCustomersInput{}, restErr
	}

	start, restErr := parseOptionalDateQuery(ctx.Query("start"), false)
	if restErr != nil {
		return serviceContracts.ListCustomersInput{}, restErr
	}

	end, restErr := parseOptionalDateQuery(ctx.Query("end"), true)
	if restErr != nil {
		return serviceContracts.ListCustomersInput{}, restErr
	}

	var status *saleDomain.PaymentStatus
	if rawStatus := strings.TrimSpace(ctx.Query("status")); rawStatus != "" {
		value := saleDomain.PaymentStatus(rawStatus)
		status = &value
	}

	var paymentType *saleDomain.PaymentType
	if rawPaymentType := strings.TrimSpace(ctx.Query("paymentType")); rawPaymentType != "" {
		value := saleDomain.PaymentType(rawPaymentType)
		paymentType = &value
	}

	return serviceContracts.ListCustomersInput{
		Page:        page,
		Limit:       limit,
		Start:       start,
		End:         end,
		Status:      status,
		PaymentType: paymentType,
		Search:      strings.TrimSpace(ctx.Query("search")),
	}, nil
}

func parsePositiveIntQuery(value string, defaultValue int, message string) (int, *rest_err.RestErr) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, rest_err.NewBadRequestError(message)
	}

	return parsed, nil
}

func parseOptionalDateQuery(value string, endOfDay bool) (*time.Time, *rest_err.RestErr) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return &parsed, nil
	}

	parsed, err = time.Parse("2006-01-02", value)
	if err != nil {
		return nil, rest_err.NewBadRequestError("Formato de data invalido")
	}

	if endOfDay {
		parsed = parsed.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	return &parsed, nil
}
