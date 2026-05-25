package request

type CustomerRequest struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone" validate:"required"`
}

type BulkDeleteCustomersRequest struct {
	IDs []int `json:"ids" validate:"required,min=1,dive,gt=0"`
}
