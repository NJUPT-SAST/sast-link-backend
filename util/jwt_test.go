package util

import (
	"fmt"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, _ := GenerateToken("xunop@qq.com")
	fmt.Println(token)
}

func TestGetUsername(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzYXN0Iiwic3ViIjoieHVub3BAcXEuY29tIiwiZXhwIjoxNjgzMjE5NTg1LCJuYmYiOjE2ODMyMDg3ODUsImlhdCI6MTY4MzIwODc4NX0.vKejtjKBLJu_CdIlvuF5zuS7VB51VHpHCntAa9yhLJY"
	claims, _ := ParseToken(token)
	fmt.Println(claims.Subject)
	username, _ := GetUsername(token)
	fmt.Println(username)
}
