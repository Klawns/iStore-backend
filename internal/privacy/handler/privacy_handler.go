package handler

import (
	authMiddleware "istore/internal/auth/middleware"
	"istore/internal/privacy/dto/request"
	"istore/internal/privacy/dto/response"
	serviceContracts "istore/internal/privacy/service/contracts"
	"istore/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PrivacyHandler struct {
	service serviceContracts.PrivacyService
}

func NewPrivacyHandler(service serviceContracts.PrivacyService) *PrivacyHandler {
	return &PrivacyHandler{service: service}
}

func (h *PrivacyHandler) ListRequests(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	requests, restErr := h.service.ListRequests(payload.UserID)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, response.ListFromDomain(requests))
}

func (h *PrivacyHandler) CreateRequest(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	var req request.PrivacyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	privacyRequest, restErr := h.service.CreateRequest(serviceContracts.CreatePrivacyRequestInput{
		UserID:  payload.UserID,
		Type:    req.Type,
		Message: req.Message,
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusCreated, response.FromDomain(privacyRequest))
}
