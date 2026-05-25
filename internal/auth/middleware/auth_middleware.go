package middleware

import (
	"istore/internal/auth/domain"
	"istore/internal/auth/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const AuthPayloadContextKey = "auth_payload"

type AuthMiddleware struct {
	jwtProvider contracts.JwtProvider
	cookie      contracts.CookieManager
}

func (m *AuthMiddleware) CSRF() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !requiresCSRF(ctx.Request.Method) {
			ctx.Next()
			return
		}

		cookieToken, err := ctx.Cookie("csrf_token")
		headerToken := ctx.GetHeader("X-CSRF-Token")
		if err != nil || cookieToken == "" || headerToken == "" || cookieToken != headerToken {
			restErr := rest_err.NewForbiddenError("invalid csrf token")
			ctx.AbortWithStatusJSON(restErr.Code, restErr)
			return
		}

		ctx.Next()
	}
}

func requiresCSRF(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func GetAuthPayload(ctx *gin.Context) (*domain.TokenPayload, *rest_err.RestErr) {
	payload, ok := ctx.Get(AuthPayloadContextKey)
	if !ok {
		return nil, rest_err.NewUnauthorizedRequestError("missing auth payload")
	}

	tokenPayload, ok := payload.(*domain.TokenPayload)
	if !ok || tokenPayload.UserID == 0 {
		return nil, rest_err.NewUnauthorizedRequestError("invalid auth payload")
	}

	return tokenPayload, nil
}

func NewAuthMiddleware(jwtProvider contracts.JwtProvider, cookie contracts.CookieManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtProvider: jwtProvider,
		cookie:      cookie,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, restErr := m.cookie.GetAuthCookie(ctx)
		if restErr != nil {
			logger.Error("error getting auth cookie", restErr, zap.String("journey", "Authenticate"))
			ctx.AbortWithStatusJSON(restErr.Code, restErr)
			return
		}

		payload, restErr := m.jwtProvider.Validate(token)
		if restErr != nil {
			logger.Error("error validating jwt", restErr, zap.String("journey", "Authenticate"))
			ctx.AbortWithStatusJSON(restErr.Code, restErr)
			return
		}

		ctx.Set(AuthPayloadContextKey, payload)
		ctx.Next()
	}
}
