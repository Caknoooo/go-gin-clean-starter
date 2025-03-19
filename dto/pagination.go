package dto

type (
	PaginationRequest struct {
		Search  string `form:"search"`
		Page    int    `form:"page"`
		PerPage int    `form:"per_page"`
	}

	PaginationResponse struct {
		Page    int   `json:"page"`
		PerPage int   `json:"per_page"`
		MaxPage int64 `json:"max_page"`
		Count   int64 `json:"count"`
	}
)

func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

func (p *PaginationRequest) GetLimit() int {
	return p.PerPage
}

func (p *PaginationRequest) GetPage() int {
	return p.Page
}

func (p *PaginationRequest) Default() {
	if p.Page == 0 {
		p.Page = 1
	}

	if p.PerPage == 0 {
		p.PerPage = 10
	}
}