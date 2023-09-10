package util

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	mr "math/rand"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"time"

	"github.com/google/uuid"
)

func OutputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

// Generate UUID
func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

// Generate random string
func GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes to base64
	randomString := base64.URLEncoding.EncodeToString(randomBytes)

	// Remove any characters that might be problematic
	//randomString = cleanRandomString(randomString)

	// Trim to desired length
	if len(randomString) > length {
		randomString = randomString[:length]
	}

	return randomString, nil
}

// GenerateCode generate a random code
func GenerateCode() string {
	seed := time.Now().UnixNano() + int64(mr.Intn(4478))
	//rand.NewSource()
	rand := mr.New(mr.NewSource(seed))
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

// sendEmail send email to user
func SendEmail(sender string, secret string, recipient string, content string) error {
	// https://gist.github.com/chrisgillis/10888032
	from := mail.Address{"", sender}
	to := mail.Address{"", recipient}
	subj := "确认电子邮件注册SAST-Link账户（无需回复）"
	body := content

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP server
	servername := "smtp.feishu.cn:465"

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", sender, secret, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	return nil
}

// ShaHashing use sha512 to hash input.
func ShaHashing(in string) string {
	sha512Hash := sha512.Sum512([]byte(in))
	return hex.EncodeToString(sha512Hash[:])
}
