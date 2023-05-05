package model

import (
	"fmt"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	code := GenerateVerifyCode("B21030518@njupt.edu.cn")
	fmt.Println(code)
}

func TestSendEmail(t *testing.T) {
	testEmail := "xunop@qq.com"
	code := GenerateVerifyCode(testEmail)
	SendEmail(testEmail, code)
}

func TestVerifyCode(t *testing.T) {
	data := "SAST"
	str := InsertCode(data)
	fmt.Println(str)
}
