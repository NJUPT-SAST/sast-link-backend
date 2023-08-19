package util

import (
	"fmt"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	for i := 0; i < 10; i++ {
		code := GenerateCode()
		fmt.Println(code)
	}
}

func TestGenerateRandomString(t *testing.T) {
	randomString, err := GenerateRandomString(32)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(randomString)
}
