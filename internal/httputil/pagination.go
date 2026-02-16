package httputil

import (
	"invento-service/internal/domain"
	"math"
)

type PaginationParams struct {
	Page  int
	Limit int
}

func NormalizePaginationParams(page, limit int) PaginationParams {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

func CalculatePagination(page, limit, totalItems int) domain.PaginationData {
	params := NormalizePaginationParams(page, limit)

	totalPages := int(math.Ceil(float64(totalItems) / float64(params.Limit)))

	return domain.PaginationData{
		Page:       params.Page,
		Limit:      params.Limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func CalculateOffset(page, limit int) int {
	params := NormalizePaginationParams(page, limit)
	return (params.Page - 1) * params.Limit
}
