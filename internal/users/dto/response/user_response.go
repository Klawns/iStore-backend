package response

import "istore/internal/users/domain"

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func FromDomain(user *domain.User) *UserResponse {
	if user == nil {
		return nil
	}

	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}
}
