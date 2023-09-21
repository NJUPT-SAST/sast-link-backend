package model

import (
	"fmt"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	code := GenerateVerifyCode()
	fmt.Println(code)
}

func TestSendEmail(t *testing.T) {
	testEmail := "xunop@qq.com"
	code := GenerateVerifyCode()
	title := "发送邮件测试"
	SendEmail(testEmail, code, title)
}

func TestVerifyCode(t *testing.T) {
	data := "SAST"
	str := InsertCode(data)
	fmt.Println(str)
}
