package result

type Response struct {
	Success bool
	ErrCode StatusCode
	ErrMsg  string
	Data    interface{}
}

func Success(data any) Response {
	return Response{
		Success: true,
		ErrCode: StatusOK,
		ErrMsg:  "",
		Data:    data,
	}
}

func Failed(errCode StatusCode, errMsg string) Response {
	return Response{
		Success: false,
		ErrCode: errCode,
		ErrMsg:  errMsg,
		Data:    nil,
	}
}
