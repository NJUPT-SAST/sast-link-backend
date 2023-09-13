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
	SendEmail(testEmail, code)
}

func TestVerifyCode(t *testing.T) {
	data := "SAST"
	str := InsertCode(data)
	fmt.Println(str)
}
