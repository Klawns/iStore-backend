package handler

import (
	authDomain "istore/internal/auth/domain"
	authMiddleware "istore/internal/auth/middleware"
	"istore/internal/users/dto/request"
	"istore/internal/users/dto/response"
	"istore/internal/users/service/contracts"
	"istore/pkg/rest_err"
	"istore/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service contracts.UserService
}

func NewUserHandler(service contracts.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(ctx *gin.Context) {
	var req request.UserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restErr := validation.ValidateUserError(err)
		ctx.JSON(restErr.Code, restErr)
		return
	}

	user, restErr := h.service.Create(contracts.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
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
