package result

import "net/http"

type Response struct {
	Success bool
	ErrCode int
	ErrMsg  string
	Data    interface{}
}

func Success(data any) Response {
	return Response{
		Success: true,
		ErrCode: http.StatusOK,
		ErrMsg:  "",
		Data:    data,
	}
}

func Failed(e LocalError) Response {
	if e.Err == nil {
		return Response{
			Success: false,
			ErrCode: e.ErrCode,
			ErrMsg:  e.ErrMsg,
			Data:    nil,
		}
	}
	return Response{
		Success: false,
		ErrCode: e.ErrCode,
		ErrMsg:  e.ErrMsg,
		Data:    e.Error(),
	}
}
