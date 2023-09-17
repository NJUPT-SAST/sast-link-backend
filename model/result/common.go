package result

import "fmt"

type Result struct {
	Success bool
	ErrCode int
	ErrMsg  string
	Data    any
}

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
	RequestParamError     = LocalError{ErrCode: 10001, ErrMsg: "请求参数错误"}
	UsernameError         = LocalError{ErrCode: 10002, ErrMsg: "用户名错误"}
	PasswordError         = LocalError{ErrCode: 10003, ErrMsg: "密码错误"}
	PasswordEmpty         = LocalError{ErrCode: 10004, ErrMsg: "密码为空"}
	PasswordIllegal       = LocalError{ErrCode: 10006, ErrMsg: "密码不合法"}
	UserAlreadyExist      = LocalError{ErrCode: 10007, ErrMsg: "重复注册"}
	LoginError            = LocalError{ErrCode: 10005, ErrMsg: "登录失败"}
	UserNotExist          = LocalError{ErrCode: 10011, ErrMsg: "用户不存在"}
	CheckExistUserFail    = LocalError{ErrCode: 10012, ErrMsg: "检查用户是否存在失败"}
	AddUserFail           = LocalError{ErrCode: 10013, ErrMsg: "添加用户失败"}
	DeleteUserFail        = LocalError{ErrCode: 10014, ErrMsg: "删除用户失败"}
	GetUserinfoFail       = LocalError{ErrCode: 10015, ErrMsg: "获取用户信息失败"}
	UserIsExist           = LocalError{ErrCode: 10016, ErrMsg: "用户已存在"}
	UserNameEmpty         = LocalError{ErrCode: 10016, ErrMsg: "用户邮箱获取失败"}
	AuthCheckTokenFail    = LocalError{ErrCode: 20001, ErrMsg: "Token鉴权失败"}
	AuthCheckTokenTimeout = LocalError{ErrCode: 20002, ErrMsg: "Token已超时"}
	GenerateToken         = LocalError{ErrCode: 20003, ErrMsg: "Token生成失败"}
	AuthError             = LocalError{ErrCode: 20004, ErrMsg: "Token错误"}
	AuthIncomingTokenFail = LocalError{ErrCode: 20005, ErrMsg: "Token为空"}
	AuthParseTokenFail    = LocalError{ErrCode: 20006, ErrMsg: "Token解析失败"}
	AuthTokenTypeError    = LocalError{ErrCode: 20010, ErrMsg: "Token类型错误"}
	TicketNotCorrect      = LocalError{ErrCode: 20007, ErrMsg: "Ticket不正确"}
	CheckTicketNotfound   = LocalError{ErrCode: 20008, ErrMsg: "Ticket不存在"}
	InvalidAccToken       = LocalError{ErrCode: 20009, ErrMsg: "无效的token"}
	SendEmailError        = LocalError{ErrCode: 30001, ErrMsg: "发送邮件失败"}
	CaptchaError          = LocalError{ErrCode: 30002, ErrMsg: "验证码错误"}
	VerifyAccountError    = LocalError{ErrCode: 40001, ErrMsg: "验证账户失败"}
	VerifyPasswordError   = LocalError{ErrCode: 40002, ErrMsg: "验证账户密码失败"}
	// this is default error
	InternalErr     = LocalError{ErrCode: 50000, ErrMsg: "未知错误"}
	ClientErr       = LocalError{ErrCode: 60001, ErrMsg: "客户端错误"}
	AccessTokenErr  = LocalError{ErrCode: 60002, ErrMsg: "access_token错误"}
	RefreshTokenErr = LocalError{ErrCode: 60003, ErrMsg: "refresh_token错误"}

	RegisterPhaseError = LocalError{ErrCode: 70003, ErrMsg: "注册步骤错误 （！！！！hack？？？？）"}
)

var errorMap = map[int]LocalError{
	10001: RequestParamError,
	10002: UsernameError,
	10003: PasswordError,
	10004: PasswordEmpty,
	10005: LoginError,
	10006: PasswordIllegal,
	10011: UserNotExist,
	10012: CheckExistUserFail,
	10013: AddUserFail,
	10014: DeleteUserFail,
	10015: GetUserinfoFail,
	10016: UserIsExist,
	20001: AuthCheckTokenFail,
	20002: AuthCheckTokenTimeout,
	20003: GenerateToken,
	20004: AuthError,
	20005: AuthIncomingTokenFail,
	20006: AuthParseTokenFail,
	20007: TicketNotCorrect,
	20008: CheckTicketNotfound,
	20009: InvalidAccToken,
	20010: AuthTokenTypeError,
	30001: SendEmailError,
	30002: CaptchaError,
	40001: VerifyAccountError,
	40002: VerifyPasswordError,
	60001: ClientErr,
	60002: AccessTokenErr,
	60003: RefreshTokenErr,
	50000: InternalErr,
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
	return InternalErr.Wrap(err)
}
