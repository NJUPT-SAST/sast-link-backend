package result

import "fmt"

type LocalError struct {
	ErrCode int
	ErrMsg  string
	Err     error
}

// Error implement error interface
func (e LocalError) Error() string {
	return fmt.Sprintf("ErrCode: %d, ErrMsg: %s, Err: %v", e.ErrCode, e.ErrMsg, e.Err)
}

// create common error
var (
	ParamError             = LocalError{ErrCode: 10001, ErrMsg: "请求参数错误"}
	UsernameOrPasswordError = LocalError{ErrCode: 10002, ErrMsg: "用户名或密码错误"}
	PasswordError          = LocalError{ErrCode: 10003, ErrMsg: "密码错误"}
	Password_NOTFOUND      = LocalError{ErrCode: 10004, ErrMsg: "密码为空"}
	LoginError						 = LocalError{ErrCode: 10005, ErrMsg: "登录失败"}
	UserNotExist           = LocalError{ErrCode: 10011, ErrMsg: "用户不存在"}
	CheckExistUserfail     = LocalError{ErrCode: 10012, ErrMsg: "检查用户是否存在失败"}
	ADD_USER_FAIL          = LocalError{ErrCode: 10013, ErrMsg: "添加用户失败"}
	DELETE_USER_FAIL       = LocalError{ErrCode: 10014, ErrMsg: "删除用户失败"}
	GET_USERINFO_FAIL      = LocalError{ErrCode: 10015, ErrMsg: "获取用户信息失败"}
	UserIsExist            = LocalError{ErrCode: 10016, ErrMsg: "用户已存在"}
	AUTH_CHECK_TOKEN_FAIL  = LocalError{ErrCode: 20001, ErrMsg: "Token鉴权失败"}
	AUTH_CHECK_TOKEN_TIMEOUT = LocalError{ErrCode: 20002, ErrMsg: "Token已超时"}
	GENERATE_TOKEN = LocalError{ErrCode: 20003, ErrMsg: "Token生成失败"}
	AUTH_ERROR = LocalError{ErrCode: 20004, ErrMsg: "Token错误"}
	AUTH_INCOMING_TOKEN_FAIL = LocalError{ErrCode: 20005, ErrMsg: "Token 为空"}
	AUTH_PARSE_TOKEN_FAIL = LocalError{ErrCode: 20006, ErrMsg: "Token解析失败"}
	TICKET_NOT_CORRECT = LocalError{ErrCode: 20007, ErrMsg: "Ticket不正确"}
	CHECK_TICKET_NOTFOUND = LocalError{ErrCode: 20008, ErrMsg: "Ticket不存在"}
	AUTH_INCOMING_TICKET_FAIL = LocalError{ErrCode: 20009, ErrMsg: "Ticket 为空"}
	SendEmailError = LocalError{ErrCode: 30001, ErrMsg: "发送邮件失败"}
	VerifyCodeError = LocalError{ErrCode: 30002, ErrMsg: "验证码错误"}
	VerifyAccountError = LocalError{ErrCode: 40001, ErrMsg: "验证账户失败"}
	VerifyPasswordError = LocalError{ErrCode: 40002, ErrMsg: "验证账户密码失败"}
	// this is default error 
	UnknownError = LocalError{ErrCode: 50000, ErrMsg: "未知错误"}
)

var errorMap = map[int]LocalError{
	10001: ParamError,
	10002: UsernameOrPasswordError,
	10003: PasswordError,
	10004: Password_NOTFOUND,
	10005: LoginError,
	10011: UserNotExist,
	10012: CheckExistUserfail,
	10013: ADD_USER_FAIL,
	10014: DELETE_USER_FAIL,
	10015: GET_USERINFO_FAIL,
	10016: UserIsExist,
	20001: AUTH_CHECK_TOKEN_FAIL,
	20002: AUTH_CHECK_TOKEN_TIMEOUT,
	20003: GENERATE_TOKEN,
	20004: AUTH_ERROR,
	20005: AUTH_INCOMING_TOKEN_FAIL,
	20006: AUTH_PARSE_TOKEN_FAIL,
	20007: TICKET_NOT_CORRECT,
	20008: CHECK_TICKET_NOTFOUND,
	20009: AUTH_INCOMING_TICKET_FAIL,
	30001: SendEmailError,
	30002: VerifyCodeError,
	40001: VerifyAccountError,
	40002: VerifyPasswordError,
	50000: UnknownError,
}

// warp error
func (e *LocalError) Wrap(err error) LocalError {
	e.Err = err
	return *e
}

// determine whether the error is equal
func (e *LocalError) Is(err error) bool {
	if err, ok := err.(LocalError); ok {
		return err.ErrCode == e.ErrCode
	}
	return false
}

// this func is used to handle error
// When a function has multiple errors, 
// instead of using if-else to determine them one by one, 
// use this function to get the errors.
func HandleError(err error) LocalError {
	if err, ok := err.(LocalError); ok {
		// determine whether the error is exist in errorMap
		if _, ok := errorMap[err.ErrCode]; ok {
			return err
		}
	}
	// if not exist, return default error
	return UnknownError.Wrap(err)
}
