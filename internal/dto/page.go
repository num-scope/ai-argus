package dto

type PageQuery struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func NewPageResponse[T any](list []T, total int64, page, pageSize int) PageResponse[T] {
	return PageResponse[T]{List: list, Total: total, Page: page, PageSize: pageSize}
}
