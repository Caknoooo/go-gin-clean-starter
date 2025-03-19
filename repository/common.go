package repository

import (
	"math"

	"github.com/Caknoooo/go-gin-clean-starter/dto"
	"gorm.io/gorm"
)

func Paginate(req dto.PaginationRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (req.Page - 1) * req.PerPage
		return db.Offset(offset).Limit(req.PerPage)
	}
}

func TotalPage(count, perPage int64) int64 {
	totalPage := int64(math.Ceil(float64(count) / float64(perPage)))
	
	return totalPage
}
