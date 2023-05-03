package result

// StatusCode represent server-side self defined code
type StatusCode int

const (
	SUCCESS        = 200
	ERROR          = 500
	INVALID_PARAMS = 400

	ERROR_NOT_EXIST_USER       = 10011
	ERROR_CHECK_EXIST_USERFAIL = 10012
	ERROR_ADD_USER_FAIL        = 10013
	ERROR_DELETE_USER_FAIL     = 10014
	ERROR_GET_USERINFO_FAIL    = 10015

	ERROR_AUTH_CHECK_TOKEN_FAIL      = 20001
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT   = 20002
	ERROR_GENERATE_TOKEN             = 20003
	ERROR_AUTH                       = 20004
	ERROR_AUTH_CHECK_TICKET_NOTFOUND = 20005
	ERROR_AUTH_INCOMING_TOKEN_FAIL   = 20006
	ERROR_AUTH_PARSE_TOKEN_FAIL      = 20007
)

// MsgFlags represent the error msg by "errCode":"errMsg" form
var MsgFlags = map[int]string{
	SUCCESS:        "ok",
	ERROR:          "fail",
	INVALID_PARAMS: "请求参数错误",

	ERROR_NOT_EXIST_USER:       "该用户不存在",
	ERROR_ADD_USER_FAIL:        "新增用户失败",
	ERROR_DELETE_USER_FAIL:     "删除用户失败",
	ERROR_CHECK_EXIST_USERFAIL: "检查用户是否存在失败",
	ERROR_GET_USERINFO_FAIL:    "查询用户信息失败",

	ERROR_AUTH_INCOMING_TOKEN_FAIL:   "TOKEN传入错误",
	ERROR_AUTH_CHECK_TOKEN_FAIL:      "Token鉴权失败",
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT:   "Token已超时",
	ERROR_GENERATE_TOKEN:             "Token生成失败",
	ERROR_AUTH:                       "Token错误",
	ERROR_AUTH_CHECK_TICKET_NOTFOUND: "TICKET不存在",
	ERROR_AUTH_PARSE_TOKEN_FAIL:      "Token解析失败",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
