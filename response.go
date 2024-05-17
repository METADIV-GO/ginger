package ginger

import "github.com/METADIV-GO/gorm"

type Response struct {
	Success    bool             `json:"success"`
	Time       string           `json:"time"`
	Duration   int64            `json:"duration"`
	Pagination *gorm.Pagination `json:"pagination,omitempty"`
	ErrMessage string           `json:"err_message,omitempty"`
	Data       any              `json:"data,omitempty"`
}

type Trace struct {
	TraceId   string     `json:"trace_id"`
	Responses []Response `json:"responses"`
}
