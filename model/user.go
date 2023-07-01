package model

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"gorm.io/gorm"
)

var ctx = context.Background()
var userLogger = log.Log

type User struct {
	ID        uint      `json:"id,omitempty" gorm:"primaryKey"`
	Uid       *string   `json:"uid,omitempty" gorm:"not null"`
	Email     *string   `json:"email,omitempty" gorm: "not null"`
	Password  *string   `json:"passowrd,omitempty" grom:"not null"`
	QQId      *string   `json:"qq_id,omitempty"`
	LarkId    *string   `json:"lark_id,omitempty"`
	GithubId  *string   `json:"github_id,omitempty"`
	WechatId  *string   `json:"wechat_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"not null"`
	IsDeleted bool      `json:"is_deleted,omitempty" gorm:"not null"`
}

func CreateUser(user *User) error {
	if res := Db.Create(user); res.Error != nil {
		return res.Error
	}
	return nil
}

func CheckPassword(username string, password string) (bool, error) {
	var user User
	matched, err2 := regexp.MatchString("@", username)
	if err2 != nil {
		userLogger.Infof("regexp matchiong error")
		return false, err2
	}
	exist := true
	var err error = nil
	if matched {
		err = Db.Where("email = ?", username).First(&user).Error
	} else {
		err = Db.Where("uid = ?", username).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Infof("User [%s] Not Exist\n", username)
			exist = false
		}
	}
	if *user.Password != password {
		exist = false
	}
	return exist, err
}

// CheckUserByEmail find user by email
// return true if user exist
func CheckUserByEmail(email string) (bool, error) {
	var user User
	err := Db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Infof("User [%s] Not Exist\n", email)
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CheckUserByUid find user by uid
// return true if user exist
func CheckUserByUid(uid string) (bool, error) {
	var user User
	err := Db.Where("uid = ?", uid).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Infof("User [%s] Not Exist\n", uid)
			return false, result.UserNotExist
		}
		return false, err
	}
	return true, nil
}

func UserInfo(username string) (*User, error) {
	var user = User{Uid: &username}
	if err := Db.First(&user).Error; err != nil {
		return nil, fmt.Errorf("%v: User [%s] Not Exist\n", err, username)
	}
	return &user, nil
}

func GenerateVerifyCode(username string) string {
	code := util.GenerateCode()
	return code
}

func SendEmail(recipient string, content string) error {
	emailInfo := conf.Sub("email")
	sender := emailInfo.GetString("sender")
	secret := emailInfo.GetString("secret")
	return util.SendEmail(sender, secret, recipient, content)
}
