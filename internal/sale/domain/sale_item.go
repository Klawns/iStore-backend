package domain

type SaleItem struct {
	ID     int
	SaleID int

	ProductName string
	Specs       string

	Quantity  int
	CostPrice int // centavos
	SalePrice int // centavos
}

func (s *SaleItem) Total() int {
	return s.Quantity * s.SalePrice
}

func (s *SaleItem) Profit() int {
	return (s.SalePrice - s.CostPrice) *
		s.Quantity
}
