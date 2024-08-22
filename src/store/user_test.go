package store

import (
	"fmt"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	code := GenerateVerifyCode()
	fmt.Println(code)
}

// func TestSendEmail(t *testing.T) {
// 	testEmail := "user@example.org"
// 	code := GenerateVerifyCode()
// 	title := "发送邮件测试，请勿回复"
// 	content := InsertCode(code)
// 	SendEmail(testEmail, content, title)
// }

func TestVerifyCode(t *testing.T) {
	data := "SAST"
	str := InsertCode(data)
	fmt.Println(str)
}
