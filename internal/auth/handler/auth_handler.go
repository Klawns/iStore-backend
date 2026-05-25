package handler

import (
	"crypto/rand"
	"encoding/base64"
	"istore/internal/auth/dto/request"
	"istore/internal/auth/service/contracts"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
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
	var req request.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	token, restErr := h.authService.SignIn(contracts.SignInInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	if restErr := h.cookieManager.SetAuthCookie(ctx, token, authCookieMaxAge); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

	csrfToken, restErr := generateCSRFToken()
	if restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}
	if restErr := h.cookieManager.SetCSRFCookie(ctx, csrfToken, authCookieMaxAge); restErr != nil {
		ctx.JSON(restErr.Code, restErr)
		return
	}

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
