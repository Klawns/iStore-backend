package implementation

import (
	"istore/internal/auth/service/contracts"
	"istore/pkg/logger"
	"istore/pkg/rest_err"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type cookieService struct {
}

func NewCookieService() contracts.CookieManager {
	return &cookieService{}
}

func (c *cookieService) SetAuthCookie(ctx *gin.Context, token string, maxAge int) *rest_err.RestErr {
	ctx.SetCookie("auth_token", token, maxAge, "/", "", false, true)
	return nil
}

func (c *cookieService) ClearAuthCookie(ctx *gin.Context) *rest_err.RestErr {
	ctx.SetCookie("auth_token", "", -1, "/", "", false, true)
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
