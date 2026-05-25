package contracts

import (
	"istore/pkg/rest_err"

	"github.com/gin-gonic/gin"
)

type CookieManager interface {
	SetAuthCookie(c *gin.Context, token string, maxAge int) *rest_err.RestErr
	SetCSRFCookie(c *gin.Context, token string, maxAge int) *rest_err.RestErr
	ClearAuthCookie(c *gin.Context) *rest_err.RestErr
	ClearCSRFCookie(c *gin.Context) *rest_err.RestErr
	GetAuthCookie(c *gin.Context) (string, *rest_err.RestErr)
}
