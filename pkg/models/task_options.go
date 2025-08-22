package models

// TaskSortBy is a struct to hold sorting options
type TaskSortBy struct {
	SortBy []string `query:"sort_by" json:"sort_by"`
	OrderBy []string `query:"order_by" json:"order_by"`
}

// TaskFilterBy is a struct to hold filtering options
type TaskFilterBy struct {
	Filter string `query:"filter" json:"filter"`
	FilterTimezone string `query:"filter_timezone" json:"-"`
	FilterIncludeNulls bool `query:"filter_include_nulls" json:"filter_include_nulls"`
}

// TaskPagination is a struct to hold pagination options
type TaskPagination struct {
	Page int `query:"page" json:"page"`
	PerPage int `query:"per_page" json:"per_page"`
}
