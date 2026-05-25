package implementation

import (
	"istore/internal/auth/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type cookieService struct {
	secure bool
}

func NewCookieService(secure bool) contracts.CookieManager {
	return &cookieService{secure: secure}
}

func (c *cookieService) SetAuthCookie(ctx *gin.Context, token string, maxAge int) *rest_err.RestErr {
	c.setCookie(ctx, "auth_token", token, maxAge, true)
	return nil
}

func (c *cookieService) SetCSRFCookie(ctx *gin.Context, token string, maxAge int) *rest_err.RestErr {
	c.setCookie(ctx, "csrf_token", token, maxAge, false)
	return nil
}

func (c *cookieService) ClearAuthCookie(ctx *gin.Context) *rest_err.RestErr {
	c.setCookie(ctx, "auth_token", "", -1, true)
	return nil
}

func (c *cookieService) ClearCSRFCookie(ctx *gin.Context) *rest_err.RestErr {
	c.setCookie(ctx, "csrf_token", "", -1, false)
	return nil
}

func (c *cookieService) GetAuthCookie(ctx *gin.Context) (string, *rest_err.RestErr) {
	token, err := ctx.Cookie("auth_token")
	if err != nil {
		logger.Error("error getting auth cookie", err, zap.String("journey", "GetAuthCookie"))
		return "", rest_err.NewUnauthorizedRequestError("error getting auth cookie")
	}
	return token, nil
}

func (c *cookieService) setCookie(ctx *gin.Context, name string, value string, maxAge int, httpOnly bool) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}
