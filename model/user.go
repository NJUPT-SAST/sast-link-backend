package model

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/util"
	"gorm.io/gorm"
)

var ctx = context.Background()

type User struct {
	ID        uint      `json:"id,omitempty" gorm:"primaryKey"`
	Uid       *string   `json:"uid,omitempty" gorm:"not null"`
	Email     *string   `json:"email,omitempty"`
	Password  *string   `json:"passowrd,omitempty" grom:"not null"`
	QQId      *string   `json:"qq_id,omitempty"`
	LarkId    *string   `json:"lark_id,omitempty"`
	GithubId  *string   `json:"github_id,omitempty"`
	WechatId  *string   `json:"wechat_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"not null"`
	IsDeleted bool      `json:"is_deleted,omitempty" gorm:"not null"`
}

func CreateUser(user *User) error {
	if res := db.Create(user); res.Error != nil {
		return res.Error
	}
	return nil
}

func VerifyAccount(username string) (bool, string, error) {
	isExist := false
	ticket := ""
	var user User
	// select user by username
	err := db.Select("email").Where("email = ?", username).First(&user).Error
	// if user not exist
	if err != nil && gorm.ErrRecordNotFound != err {
		return isExist, ticket, err
	}
	// if user != null
	if user != (User{}) {
		isExist = true
	}

	if isExist {
		ticket, err = util.GenerateToken(username)
		// 5min expire
		Rdb.Set(ctx, "TICKET:"+username, ticket, time.Minute*5)
	}
	return isExist, ticket, nil
}

func UserInfo(username string) (*User, error) {
	var user = User{Uid: &username}
	if err := db.First(&user).Error; err != nil {
		return nil, fmt.Errorf("%v: User [%s] Not Exist\n", err, username)
	}
	return &user, nil
}

func GenerateVerifyCode(username string) string {
	code := util.GenerateCode()
	// 5min expire
	Rdb.Set(ctx, "VERIFY_CODE:"+username, code, time.Minute*5)
	return code
}

func SendEmail(recipient string, content string) error {
	// https://gist.github.com/chrisgillis/10888032
	emailInfo := conf.Sub("email")
	sender := emailInfo.GetString("sender")
	secret := emailInfo.GetString("secret")
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
