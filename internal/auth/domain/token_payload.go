package domain

type TokenPayload struct {
	UserID uint
	Email  string
	Exp    int64
}
