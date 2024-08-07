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
	RequestParamError  = LocalError{ErrCode: 10001, ErrMsg: "请求参数错误"}
	UsernameError      = LocalError{ErrCode: 10002, ErrMsg: "用户名错误"}
	PasswordError      = LocalError{ErrCode: 10003, ErrMsg: "密码错误"}
	PasswordEmpty      = LocalError{ErrCode: 10004, ErrMsg: "密码为空"}
	LoginError         = LocalError{ErrCode: 10005, ErrMsg: "登录失败"}
	PasswordIllegal    = LocalError{ErrCode: 10006, ErrMsg: "密码不合法"}
	UserAlreadyExist   = LocalError{ErrCode: 10007, ErrMsg: "重复注册"}
	OauthUserUnbounded = LocalError{ErrCode: 10010, ErrMsg: "Oauth用户未注册或未绑定"}
	OauthTokenError    = LocalError{ErrCode: 20004, ErrMsg: "Oauth Token错误"}
	UserNotExist       = LocalError{ErrCode: 10011, ErrMsg: "用户不存在"}
	CheckExistUserFail = LocalError{ErrCode: 10012, ErrMsg: "检查用户是否存在失败"}
	AddUserFail        = LocalError{ErrCode: 10013, ErrMsg: "添加用户失败"}
	DeleteUserFail     = LocalError{ErrCode: 10014, ErrMsg: "删除用户失败"}
	GetUserinfoFail    = LocalError{ErrCode: 10015, ErrMsg: "获取用户信息失败"}
	UserIsExist        = LocalError{ErrCode: 10016, ErrMsg: "用户已存在"}

	AuthCheckTokenTimeout = LocalError{ErrCode: 20002, ErrMsg: "Token已超时"}
	GenerateToken         = LocalError{ErrCode: 20003, ErrMsg: "Token生成失败"}
	TokenError            = LocalError{ErrCode: 20004, ErrMsg: "Token错误"}
	AuthParseTokenFail    = LocalError{ErrCode: 20006, ErrMsg: "Token解析失败"}
	TicketNotCorrect      = LocalError{ErrCode: 20007, ErrMsg: "Ticket不正确"}
	CheckTicketNotfound   = LocalError{ErrCode: 20008, ErrMsg: "Ticket不存在"}
	InvalidAccToken       = LocalError{ErrCode: 20009, ErrMsg: "无效的token"}

	SendEmailError      = LocalError{ErrCode: 30001, ErrMsg: "发送邮件失败"}
	CaptchaError        = LocalError{ErrCode: 30002, ErrMsg: "验证码错误"}
	UserEmailError      = LocalError{ErrCode: 30003, ErrMsg: "邮箱格式错误"}
	VerifyAccountError  = LocalError{ErrCode: 40001, ErrMsg: "验证账户失败"}
	VerifyPasswordError = LocalError{ErrCode: 40002, ErrMsg: "验证账户密码失败"}
	// this is default error
	InternalErr           = LocalError{ErrCode: 50000, ErrMsg: "未知错误"}
	ClientErr             = LocalError{ErrCode: 60001, ErrMsg: "客户端错误"}
	AccessTokenErr        = LocalError{ErrCode: 60002, ErrMsg: "access_token错误"}
	RefreshTokenErr       = LocalError{ErrCode: 60003, ErrMsg: "refresh_token错误"}
	RegisterPhaseError    = LocalError{ErrCode: 70003, ErrMsg: "注册失败 （！！！！hack？？？？）"}
	ResetPasswordEror     = LocalError{ErrCode: 70004, ErrMsg: "重置密码失败 （！！！！hack？？？？）"}
	AlreadySetPasswordErr = LocalError{ErrCode: 70004, ErrMsg: "重复设置密码"}

	ProfileNotExist  = LocalError{ErrCode: 80000, ErrMsg: "用户profile不存在"}
	OrgIdError       = LocalError{ErrCode: 80001, ErrMsg: "组织填写错误"}
	CheckHideIllegal = LocalError{ErrCode: 80002, ErrMsg: "填写隐藏信息不合法"}

	SentMsgToBotErr  = LocalError{ErrCode: 90000, ErrMsg: "发送审核通知信息失败"}
	DealFrozenImgErr = LocalError{ErrCode: 90001, ErrMsg: "处理冻结图片失败"}
	PicURLErr        = LocalError{ErrCode: 90002, ErrMsg: "图片URL地址错误"}
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
	20002: AuthCheckTokenTimeout,
	20003: GenerateToken,
	20004: TokenError,
	20006: AuthParseTokenFail,
	20007: TicketNotCorrect,
	20008: CheckTicketNotfound,
	20009: InvalidAccToken,
	30001: SendEmailError,
	30002: CaptchaError,
	30003: UserEmailError,
	40001: VerifyAccountError,
	40002: VerifyPasswordError,
	60001: ClientErr,
	60002: AccessTokenErr,
	60003: RefreshTokenErr,
	50000: InternalErr,
	80000: ProfileNotExist,
	80001: OrgIdError,
	80002: CheckHideIllegal,
	90000: SentMsgToBotErr,
	90001: DealFrozenImgErr,
	90002: PicURLErr,
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
		// determine whether the error is existed in errorMap
		if _, ok := errorMap[err.ErrCode]; ok {
			return err
		}
	}
	// if not exist, return default error
	return InternalErr.Wrap(err)
}
func HandleErrorWithArgu(err error, localError LocalError) LocalError {
	if err, ok := err.(LocalError); ok {
		// determine whether the error is existed in errorMap
		if _, ok := errorMap[err.ErrCode]; ok {
			return err
		}
	}
	// if not exist, return default error warped with specified localError
	return localError.Wrap(err)
}
