package util

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/mail"
	"net/smtp"
	"time"
)

// GenerateCode generate a random code
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

// sendEmail send email to user
func SendEmail(sender string, secret string, recipient string, content string) error {
	// https://gist.github.com/chrisgillis/10888032
	from := mail.Address{"", sender}
	to := mail.Address{"", recipient}
	subj := "确认电子邮件注册SAST-Link账户"
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
