package handler

import (
	authDomain "istore/internal/auth/domain"
	authMiddleware "istore/internal/auth/middleware"
	authContracts "istore/internal/auth/service/contracts"
	"istore/internal/users/dto/request"
	"istore/internal/users/dto/response"
	"istore/internal/users/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service       contracts.UserService
	cookieManager authContracts.CookieManager
}

func NewUserHandler(service contracts.UserService, cookieManager authContracts.CookieManager) *UserHandler {
	return &UserHandler{service: service, cookieManager: cookieManager}
}

func (h *UserHandler) Create(ctx *gin.Context) {
	logger.Info("create user request received", zap.String("journey", "CreateUser"), zap.String("stage", "received"))

	var req request.UserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		logger.Error("create user request validation failed", err, zap.String("journey", "CreateUser"), zap.String("stage", "bind_json"), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	logger.Info("create user service starting", zap.String("journey", "CreateUser"), zap.String("stage", "service_start"))
	user, restErr := h.service.Create(contracts.CreateUserInput{
		Email:                req.Email,
		Password:             req.Password,
		AcceptPrivacyPolicy:  req.AcceptPrivacyPolicy,
		AcceptTerms:          req.AcceptTerms,
		PrivacyPolicyVersion: req.PrivacyPolicyVersion,
		TermsVersion:         req.TermsVersion,
	})
	if restErr != nil {
		logger.Error("create user service failed", restErr, zap.String("journey", "CreateUser"), zap.String("stage", "service_error"), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	logger.Info("create user completed", zap.String("journey", "CreateUser"), zap.String("stage", "completed"), zap.Uint("user_id", user.ID))
	ctx.JSON(http.StatusCreated, response.FromDomain(user))
}

func (h *UserHandler) Me(ctx *gin.Context) {
	logger.Info("me request received", zap.String("journey", "Me"), zap.String("stage", "received"))

	payload, ok := ctx.Get(authMiddleware.AuthPayloadContextKey)
	if !ok {
		restErr := rest_err.NewUnauthorizedRequestError("missing auth payload")
		logger.Error("me auth payload missing", restErr, zap.String("journey", "Me"), zap.String("stage", "auth_payload"))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	tokenPayload, ok := payload.(*authDomain.TokenPayload)
	if !ok {
		restErr := rest_err.NewUnauthorizedRequestError("invalid auth payload")
		logger.Error("me auth payload invalid", restErr, zap.String("journey", "Me"), zap.String("stage", "auth_payload"))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	logger.Info("me user lookup starting", zap.String("journey", "Me"), zap.String("stage", "find_user_by_id"), zap.Uint("user_id", tokenPayload.UserID))
	user, restErr := h.service.FindByID(tokenPayload.UserID)
	if restErr != nil {
		logger.Error("me user lookup failed", restErr, zap.String("journey", "Me"), zap.String("stage", "find_user_by_id"), zap.Uint("user_id", tokenPayload.UserID), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	logger.Info("me completed", zap.String("journey", "Me"), zap.String("stage", "completed"), zap.Uint("user_id", tokenPayload.UserID))
	ctx.JSON(http.StatusOK, response.FromDomain(user))
}

func (h *UserHandler) DeleteMe(ctx *gin.Context) {
	payload, restErr := authMiddleware.GetAuthPayload(ctx)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	var req request.DeleteOwnAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	if restErr := h.service.DeleteOwnAccount(payload.UserID, req.Password); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	if restErr := h.cookieManager.ClearAuthCookie(ctx); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := h.cookieManager.ClearCSRFCookie(ctx); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}
