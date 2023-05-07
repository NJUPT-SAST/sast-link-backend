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

// 创建多个错误变量
var ParamError = LocalError{ErrCode: 10001, ErrMsg: "请求参数错误", Err: nil}
var UsernameOrPasswordError = LocalError{ErrCode: 10002, ErrMsg: "用户名或密码错误", Err: nil}
var PasswordError = LocalError{ErrCode: 10003, ErrMsg: "密码错误", Err: nil}
var Password_NOTFOUND = LocalError{ErrCode: 10004, ErrMsg: "密码为空", Err: nil}
var UserNotExist = LocalError{ErrCode: 10011, ErrMsg: "用户不存在", Err: nil}
var CheckExistUserfail = LocalError{ErrCode: 10012, ErrMsg: "检查用户是否存在失败", Err: nil}
var ADD_USER_FAIL = LocalError{ErrCode: 10013, ErrMsg: "添加用户失败", Err: nil}
var DELETE_USER_FAIL = LocalError{ErrCode: 10014, ErrMsg: "删除用户失败", Err: nil}
var GET_USERINFO_FAIL = LocalError{ErrCode: 10015, ErrMsg: "获取用户信息失败", Err: nil}
var UserIsExist = LocalError{ErrCode: 10016, ErrMsg: "用户已存在", Err: nil}
var AUTH_CHECK_TOKEN_FAIL = LocalError{ErrCode: 20001, ErrMsg: "Token鉴权失败", Err: nil}
var AUTH_CHECK_TOKEN_TIMEOUT = LocalError{ErrCode: 20002, ErrMsg: "Token已超时", Err: nil}
var GENERATE_TOKEN = LocalError{ErrCode: 20003, ErrMsg: "Token生成失败", Err: nil}
var AUTH_ERROR = LocalError{ErrCode: 20004, ErrMsg: "Token错误", Err: nil}
var TOKEN_NOT_EXIST = LocalError{ErrCode: 20009, ErrMsg: "Token不存在", Err: nil}
var AUTH_INCOMING_TOKEN_FAIL = LocalError{ErrCode: 20005, ErrMsg: "Token 为空", Err: nil}
var AUTH_PARSE_TOKEN_FAIL = LocalError{ErrCode: 20006, ErrMsg: "Token解析失败", Err: nil}
var TICKET_NOT_CORRECT = LocalError{ErrCode: 20007, ErrMsg: "Ticket不正确", Err: nil}
var CHECK_TICKET_NOTFOUND = LocalError{ErrCode: 20008, ErrMsg: "Ticket不存在", Err: nil}
var AUTH_INCOMING_TICKET_FAIL = LocalError{ErrCode: 20009, ErrMsg: "Ticket 为空", Err: nil}
var SendEmailError = LocalError{ErrCode: 30001, ErrMsg: "发送邮件失败", Err: nil}
var VerifyCodeError = LocalError{ErrCode: 30002, ErrMsg: "验证码错误", Err: nil}
var VerifyAccountError = LocalError{ErrCode: 40001, ErrMsg: "验证账户失败", Err: nil}
var VerifyPasswordError = LocalError{ErrCode: 40002, ErrMsg: "验证账户密码失败", Err: nil}

// warp error
func (e *LocalError) Wrap(err error) LocalError {
	e.Err = err
	return *e
}

func (e *LocalError) Is(err error) bool {
	if err, ok := err.(LocalError); ok {
		return err.ErrCode == e.ErrCode
	}
	return false
}
