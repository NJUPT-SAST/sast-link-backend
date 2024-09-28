package store

type CheckRes struct {
	Code int `json:"code"`
	Data struct {
		ForbiddenStatus int    `json:"forbidden_status"`
		Event           string `json:"event"`
		Result          int    `json:"result"`
		TraceID         string `json:"trace_id"`
		URL             string `json:"url"`
	}
	Message string `json:"message"`
}
