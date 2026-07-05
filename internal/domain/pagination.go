package domain

type Pagination struct {
	Page     uint64
	PageSize uint64
}

func (p Pagination) Offset() uint64 {
	if p.Page <= 1 {
		return 0
	}

	return (p.Page - 1) * p.PageSize
}