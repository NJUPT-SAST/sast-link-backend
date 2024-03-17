package model

type CheckRes struct {
	Code int `json:"code"`
	Data struct {
		ForbiddenStatus int    `json:"forbidden_status"`
		Event           string `json:"event"`
		Result          int    `json:"result"`
		TraceId         string `json:"trace_id"`
		Url             string `json:"url"`
	}
	Message string `json:"message"`
}
