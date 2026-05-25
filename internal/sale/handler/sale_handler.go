package handler

import (
	authMiddleware "istore/internal/auth/middleware"
	"istore/internal/sale/domain"
	"istore/internal/sale/dto/request"
	"istore/internal/sale/dto/response"
	"istore/internal/sale/service/contract"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SaleHandler struct {
	service contract.SaleService
}

func NewSaleHandler(service contract.SaleService) *SaleHandler {
	return &SaleHandler{service: service}
}

func (h *SaleHandler) Create(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	var req request.SaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		logSaleRestErr(ctx, "CreateSale", restErr, err, saleRequestFields(req)...)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	sale, restErr := h.service.Create(&contract.CreateSaleInput{
		UserID:          payload.UserID,
		ClienteID:       req.ClienteID,
		TipoPagamento:   req.TipoPagamento,
		StatusPagamento: req.StatusPagamento,
		SaleDate:        req.SaleDate,
		Installments:    req.Installments,
		BillingDay:      req.BillingDay,
		Itens:           toCreateSaleItems(req.Itens),
	})
	if restErr != nil {
		logSaleRestErr(ctx, "CreateSale", restErr, restErr, saleRequestFields(req)...)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusCreated, response.FromDomain(sale))
}

func (h *SaleHandler) GetByID(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	id, restErr := getIDParam(ctx)
	if restErr != nil {
		logSaleRestErr(ctx, "GetSaleByID", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	sale, restErr := h.service.GetByID(payload.UserID, id)
	if restErr != nil {
		logSaleRestErr(ctx, "GetSaleByID", restErr, restErr, zap.Int("sale_id", id))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromDomain(sale))
}

func (h *SaleHandler) List(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	input, restErr := parseListSalesQuery(ctx)
	if restErr != nil {
		logSaleRestErr(ctx, "ListSales", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	input.UserID = payload.UserID
	result, restErr := h.service.List(input)
	if restErr != nil {
		logSaleRestErr(ctx, "ListSales", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.ListFromDomain(result))
}

func (h *SaleHandler) ListByPeriod(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	start, restErr := parseTimeQuery(ctx.Query("start"))
	if restErr != nil {
		logSaleRestErr(ctx, "ListSalesByPeriod", restErr, restErr, zap.String("start", ctx.Query("start")), zap.String("end", ctx.Query("end")))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	end, restErr := parseTimeQuery(ctx.Query("end"))
	if restErr != nil {
		logSaleRestErr(ctx, "ListSalesByPeriod", restErr, restErr, zap.Time("start", start), zap.String("end", ctx.Query("end")))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	sales, restErr := h.service.ListByPeriod(payload.UserID, start, end)
	if restErr != nil {
		logSaleRestErr(ctx, "ListSalesByPeriod", restErr, restErr, zap.Time("start", start), zap.Time("end", end))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	responses := make([]response.SaleResponse, len(sales))
	for i, sale := range sales {
		responses[i] = *response.FromDomain(&sale)
	}

	ctx.JSON(http.StatusOK, responses)
}

func (h *SaleHandler) ListInstallmentAlerts(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	installments, restErr := h.service.ListInstallmentAlerts(payload.UserID, time.Now(), 7)
	if restErr != nil {
		logSaleRestErr(ctx, "ListInstallmentAlerts", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	responses := make([]response.SaleInstallmentResponse, len(installments))
	for i, installment := range installments {
		responses[i] = *response.SaleInstallmentFromDomain(&installment)
	}

	ctx.JSON(http.StatusOK, responses)
}

func (h *SaleHandler) ListInstallmentsBySaleID(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	id, restErr := getIDParam(ctx)
	if restErr != nil {
		logSaleRestErr(ctx, "ListInstallmentsBySaleID", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	installments, restErr := h.service.ListInstallmentsBySaleID(payload.UserID, id)
	if restErr != nil {
		logSaleRestErr(ctx, "ListInstallmentsBySaleID", restErr, restErr, zap.Int("sale_id", id))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	responses := make([]response.SaleInstallmentResponse, len(installments))
	for i, installment := range installments {
		responses[i] = *response.SaleInstallmentFromDomain(&installment)
	}

	ctx.JSON(http.StatusOK, responses)
}

func (h *SaleHandler) UpdateInstallmentStatus(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	id, restErr := getIDParam(ctx)
	if restErr != nil {
		logSaleRestErr(ctx, "UpdateInstallmentStatus", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	var req request.SaleInstallmentStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		logSaleRestErr(ctx, "UpdateInstallmentStatus", restErr, err, zap.Int("installment_id", id))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	installment, restErr := h.service.UpdateInstallmentStatus(payload.UserID, id, contract.UpdateInstallmentStatusInput{
		Status: req.Status,
		Notes:  req.Notes,
	})
	if restErr != nil {
		logSaleRestErr(ctx, "UpdateInstallmentStatus", restErr, restErr, zap.Int("installment_id", id), zap.String("status", string(req.Status)))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.SaleInstallmentFromDomain(installment))
}

func (h *SaleHandler) UpdateStatus(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	id, restErr := getIDParam(ctx)
	if restErr != nil {
		logSaleRestErr(ctx, "UpdateSaleStatus", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	var req request.SaleStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		logSaleRestErr(ctx, "UpdateSaleStatus", restErr, err, zap.Int("sale_id", id))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	restErr = h.service.UpdateStatus(payload.UserID, id, req.Status)
	if restErr != nil {
		logSaleRestErr(ctx, "UpdateSaleStatus", restErr, restErr, zap.Int("sale_id", id), zap.String("payment_status", string(req.Status)))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *SaleHandler) Delete(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	id, restErr := getIDParam(ctx)
	if restErr != nil {
		logSaleRestErr(ctx, "DeleteSale", restErr, restErr)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	restErr = h.service.Delete(payload.UserID, id)
	if restErr != nil {
		logSaleRestErr(ctx, "DeleteSale", restErr, restErr, zap.Int("sale_id", id))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func parseListSalesQuery(ctx *gin.Context) (contract.ListSalesInput, *rest_err.RestErr) {
	page, restErr := parsePositiveIntQuery(ctx.Query("page"), 1, "Pagina invalida")
	if restErr != nil {
		return contract.ListSalesInput{}, restErr
	}

	limit, restErr := parsePositiveIntQuery(ctx.Query("limit"), 10, "Limite invalido")
	if restErr != nil {
		return contract.ListSalesInput{}, restErr
	}

	start, restErr := parseOptionalDateQuery(ctx.Query("start"), false)
	if restErr != nil {
		return contract.ListSalesInput{}, restErr
	}

	end, restErr := parseOptionalDateQuery(ctx.Query("end"), true)
	if restErr != nil {
		return contract.ListSalesInput{}, restErr
	}

	var status *domain.PaymentStatus
	if rawStatus := strings.TrimSpace(ctx.Query("status")); rawStatus != "" {
		value := domain.PaymentStatus(rawStatus)
		status = &value
	}

	var paymentType *domain.PaymentType
	if rawPaymentType := strings.TrimSpace(ctx.Query("paymentType")); rawPaymentType != "" {
		value := domain.PaymentType(rawPaymentType)
		paymentType = &value
	}

	var customerID *int
	if rawCustomerID := strings.TrimSpace(ctx.Query("customerId")); rawCustomerID != "" {
		value, err := strconv.Atoi(rawCustomerID)
		if err != nil {
			return contract.ListSalesInput{}, rest_err.NewBadRequestError("Cliente invalido")
		}
		customerID = &value
	}

	return contract.ListSalesInput{
		Page:        page,
		Limit:       limit,
		Start:       start,
		End:         end,
		Status:      status,
		PaymentType: paymentType,
		CustomerID:  customerID,
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

func logSaleRestErr(ctx *gin.Context, journey string, restErr *rest_err.RestErr, err error, fields ...zap.Field) {
	if restErr == nil {
		return
	}

	baseFields := []zap.Field{
		zap.String("journey", journey),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.Request.URL.Path),
		zap.String("client_ip", ctx.ClientIP()),
		zap.Int("status_code", restErr.Code),
		zap.Any("causes", normalizedCauses(restErr, err)),
	}

	logger.Error(restErr.Message, err, append(baseFields, fields...)...)
}

func normalizedCauses(restErr *rest_err.RestErr, err error) []rest_err.Causes {
	if len(restErr.Causes) > 0 {
		return restErr.Causes
	}

	if err != nil {
		return []rest_err.Causes{{Message: err.Error()}}
	}

	return []rest_err.Causes{{Message: restErr.Message}}
}

func saleRequestFields(req request.SaleRequest) []zap.Field {
	return []zap.Field{
		zap.Int64("customer_id", req.ClienteID),
		zap.String("payment_type", string(req.TipoPagamento)),
		zap.String("payment_status", string(req.StatusPagamento)),
		zap.Int("items_count", len(req.Itens)),
		zap.Any("installments", req.Installments),
		zap.Any("billing_day", req.BillingDay),
	}
}
