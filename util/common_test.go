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
