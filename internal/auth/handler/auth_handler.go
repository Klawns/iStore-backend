package handler

import (
	"crypto/rand"
	"encoding/base64"
	"istore/internal/auth/dto/request"
	"istore/internal/auth/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const authCookieMaxAge = 24 * 60 * 60

type AuthHandler struct {
	authService   contracts.AuthService
	cookieManager contracts.CookieManager
}

func NewAuthHandler(authService contracts.AuthService, cookieManager contracts.CookieManager) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		cookieManager: cookieManager,
	}
}

func (h *AuthHandler) SignIn(ctx *gin.Context) {
	logger.Info("sign in request received", zap.String("journey", "SignIn"), zap.String("stage", "received"))

	var req request.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		logger.Error("sign in request validation failed", err, zap.String("journey", "SignIn"), zap.String("stage", "bind_json"), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	logger.Info("sign in service starting", zap.String("journey", "SignIn"), zap.String("stage", "service_start"))
	token, restErr := h.authService.SignIn(contracts.SignInInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if restErr != nil {
		logger.Error("sign in service failed", restErr, zap.String("journey", "SignIn"), zap.String("stage", "service_error"), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	logger.Info("sign in auth cookie starting", zap.String("journey", "SignIn"), zap.String("stage", "set_auth_cookie"))
	if restErr := h.cookieManager.SetAuthCookie(ctx, token, authCookieMaxAge); restErr != nil {
		logger.Error("sign in auth cookie failed", restErr, zap.String("journey", "SignIn"), zap.String("stage", "set_auth_cookie"), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	csrfToken, restErr := generateCSRFToken()
	if restErr != nil {
		logger.Error("sign in csrf token failed", restErr, zap.String("journey", "SignIn"), zap.String("stage", "generate_csrf"), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := h.cookieManager.SetCSRFCookie(ctx, csrfToken, authCookieMaxAge); restErr != nil {
		logger.Error("sign in csrf cookie failed", restErr, zap.String("journey", "SignIn"), zap.String("stage", "set_csrf_cookie"), zap.Int("status", restErr.Code))
		ctx.JSON(restErr.Code, restErr)
		return
	}

	logger.Info("sign in completed", zap.String("journey", "SignIn"), zap.String("stage", "completed"))
	ctx.JSON(http.StatusOK, gin.H{"message": "signed in"})
}

func (h *AuthHandler) SignOut(ctx *gin.Context) {
	if restErr := h.cookieManager.ClearAuthCookie(ctx); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := h.cookieManager.ClearCSRFCookie(ctx); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "signed out"})
}

func generateCSRFToken() (string, *rest_err.RestErr) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", rest_err.NewInternalServerError("error creating session")
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
