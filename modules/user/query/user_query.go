package query

import (
	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	TelpNumber string `json:"telp_number"`
	Role       string `json:"role"`
	ImageUrl   string `json:"image_url"`
	IsVerified bool   `json:"is_verified"`
}

type UserFilter struct {
	pagination.BaseFilter
}

func (f *UserFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	// Apply your filters here
	return query
}

func (f *UserFilter) GetTableName() string {
	return "users"
}

func (f *UserFilter) GetSearchFields() []string {
	return []string{"name"}
}

func (f *UserFilter) GetDefaultSort() string {
	return "id asc"
}

func (f *UserFilter) GetIncludes() []string {
	return f.Includes
}

func (f *UserFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *UserFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *UserFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{}
}
