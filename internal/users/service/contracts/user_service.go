package contracts

import (
	"istore/internal/users/dto/request"
	"istore/internal/users/dto/response"
	"istore/pkg/rest_err"
)

type UserService interface {
	Create(req request.UserRequest) (*response.UserResponse, *rest_err.RestErr)
	FindByID(id uint) (*response.UserResponse, *rest_err.RestErr)
}
