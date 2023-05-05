package util

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateCode
func GenerateCode() string {
	seed := time.Now().UnixNano() + int64(rand.Intn(4478))
	rand.Seed(seed)
	// 除去容易混淆的字符
	chars := "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

	// 生成6位随机数
	code := make([]byte, 5)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}

	// 将字符数组转换为字符串，并将第一个字符设置为"S"
	codeStr := fmt.Sprintf("S-%s", string(code))
	return codeStr
}
