package handler

import (
	"istore/internal/analytics/domain"
	"istore/internal/analytics/dto/response"
	serviceContracts "istore/internal/analytics/service/contracts"
	authMiddleware "istore/internal/auth/middleware"
	saleDomain "istore/internal/sale/domain"
	"istore/pkg/rest_err"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service serviceContracts.AnalyticsService
}

func NewAnalyticsHandler(service serviceContracts.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) Dashboard(ctx *gin.Context) {
	filter, restErr := filterFromQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := addAuthenticatedUser(ctx, &filter); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	metrics, restErr := h.service.GetDashboard(filter)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromDashboardDomain(metrics))
}

func (h *AnalyticsHandler) Revenue(ctx *gin.Context) {
	filter, restErr := filterFromQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := addAuthenticatedUser(ctx, &filter); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	metrics, restErr := h.service.GetRevenue(filter)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromFinancialDomainSlice(metrics))
}

func (h *AnalyticsHandler) Profit(ctx *gin.Context) {
	filter, restErr := filterFromQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := addAuthenticatedUser(ctx, &filter); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	metrics, restErr := h.service.GetProfit(filter)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromFinancialDomainSlice(metrics))
}

func (h *AnalyticsHandler) TopProducts(ctx *gin.Context) {
	filter, restErr := filterFromQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := addAuthenticatedUser(ctx, &filter); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	metrics, restErr := h.service.GetTopProducts(filter)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromProductDomainSlice(metrics))
}

func (h *AnalyticsHandler) PaymentMethods(ctx *gin.Context) {
	filter, restErr := filterFromQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := addAuthenticatedUser(ctx, &filter); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	metrics, restErr := h.service.GetPaymentMethods(filter)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromPaymentDomainSlice(metrics))
}

func (h *AnalyticsHandler) TopCustomers(ctx *gin.Context) {
	filter, restErr := filterFromQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := addAuthenticatedUser(ctx, &filter); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	metrics, restErr := h.service.GetTopCustomers(filter)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromCustomerDomainSlice(metrics))
}

func (h *AnalyticsHandler) Statuses(ctx *gin.Context) {
	filter, restErr := filterFromQuery(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := addAuthenticatedUser(ctx, &filter); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	metrics, restErr := h.service.GetStatuses(filter)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.FromStatusDomainSlice(metrics))
}

func addAuthenticatedUser(ctx *gin.Context, filter *domain.AnalyticsFilter) *rest_err.RestErr {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		return restErr
	}

	filter.UserID = payload.UserID
	return nil
}

func filterFromQuery(ctx *gin.Context) (domain.AnalyticsFilter, *rest_err.RestErr) {
	start, restErr := parseOptionalTimeQuery(ctx.Query("start"), false)
	if restErr != nil {
		return domain.AnalyticsFilter{}, restErr
	}

	end, restErr := parseOptionalTimeQuery(ctx.Query("end"), true)
	if restErr != nil {
		return domain.AnalyticsFilter{}, restErr
	}

	limit, restErr := parseOptionalIntQuery(ctx.Query("limit"))
	if restErr != nil {
		return domain.AnalyticsFilter{}, restErr
	}

	status := saleDomain.PaymentStatus(strings.ToUpper(strings.TrimSpace(ctx.Query("status"))))
	paymentType := saleDomain.PaymentType(strings.ToUpper(strings.TrimSpace(ctx.Query("paymentType"))))
	groupBy := strings.ToLower(strings.TrimSpace(ctx.Query("groupBy")))
	if groupBy == "" {
		groupBy = strings.ToLower(strings.TrimSpace(ctx.Query("group")))
	}

	return domain.AnalyticsFilter{
		StartDate:   start,
		EndDate:     end,
		Limit:       limit,
		Status:      status,
		PaymentType: paymentType,
		GroupBy:     groupBy,
	}, nil
}

func parseOptionalTimeQuery(value string, endOfDay bool) (time.Time, *rest_err.RestErr) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}

	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		if endOfDay {
			return parsed.AddDate(0, 0, 1).Add(-time.Nanosecond), nil
		}

		return parsed, nil
	}

	return time.Time{}, rest_err.NewBadRequestError("Formato de data invalido")
}

func parseOptionalIntQuery(value string) (int, *rest_err.RestErr) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, rest_err.NewBadRequestError("Limite invalido")
	}

	return parsed, nil
}
