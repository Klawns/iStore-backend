package middleware

import (
	"istore/internal/auth/service/contracts"
	"istore/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const AuthPayloadContextKey = "auth_payload"

type AuthMiddleware struct {
	jwtProvider contracts.JwtProvider
	cookie      contracts.CookieManager
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
