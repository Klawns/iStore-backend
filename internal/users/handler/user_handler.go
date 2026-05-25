package handler

import (
	authDomain "istore/internal/auth/domain"
	authMiddleware "istore/internal/auth/middleware"
	authContracts "istore/internal/auth/service/contracts"
	"istore/internal/users/dto/request"
	"istore/internal/users/dto/response"
	"istore/internal/users/service/contracts"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service       contracts.UserService
	cookieManager authContracts.CookieManager
}

func NewUserHandler(service contracts.UserService, cookieManager authContracts.CookieManager) *UserHandler {
	return &UserHandler{service: service, cookieManager: cookieManager}
}

func (h *UserHandler) Create(ctx *gin.Context) {
	var req request.UserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	user, restErr := h.service.Create(contracts.CreateUserInput{
		Email:                req.Email,
		Password:             req.Password,
		AcceptPrivacyPolicy:  req.AcceptPrivacyPolicy,
		AcceptTerms:          req.AcceptTerms,
		PrivacyPolicyVersion: req.PrivacyPolicyVersion,
		TermsVersion:         req.TermsVersion,
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusCreated, response.FromDomain(user))
}

func (h *UserHandler) Me(ctx *gin.Context) {
	payload, ok := ctx.Get(authMiddleware.AuthPayloadContextKey)
	if !ok {
		restErr := rest_err.NewUnauthorizedRequestError("missing auth payload")
		ctx.JSON(restErr.Code, restErr)
		return
	}

	tokenPayload, ok := payload.(*authDomain.TokenPayload)
	if !ok {
		restErr := rest_err.NewUnauthorizedRequestError("invalid auth payload")
		ctx.JSON(restErr.Code, restErr)
		return
	}

	user, restErr := h.service.FindByID(tokenPayload.UserID)
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

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
