package model

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"gorm.io/gorm"
)

var userLogger = log.Log

type User struct {
	ID        uint      `json:"id,omitempty" gorm:"primaryKey"`
	Uid       *string   `json:"uid,omitempty" gorm:"not null"`
	Email     *string   `json:"email,omitempty" gorm:"not null"`
	Password  *string   `json:"password,omitempty" gorm:"not null"`
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

func CheckPassword(username string, password string) error {
	//get uid from username by regexp
	var user User
	matched, err2 := regexp.MatchString("@", username)
	if err2 != nil {
		//print err log
		userLogger.Infof("regexp matchiong error")
		return err2
	}
	//get user by email/uid
	var err error = nil
	if matched {
		err = Db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	} else {
		err = Db.Where("uid = ?", username).Where("is_deleted = ?", false).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Infof("User [%s] Not Exist\n", username)
		}
		userLogger.Infof("gorm search user fail")
		return err
	}
	//encrypt and verify password
	pwdEncrypted := util.ShaHashing(password)
	if *user.Password != pwdEncrypted {
		//return err to front
		err = result.PasswordError
	}
	return err
}

func ChangePassword(username string, password string) error {
	pwdEncrypted := util.ShaHashing(password)
	matched, err := regexp.MatchString("@", username)
	if err != nil {
		userLogger.Infof("regexp matchiong error")
		return err
	}
	//get user by email/uid
	if matched {
		err = Db.Model(&User{}).Where("email = ?", username).Where("is_deleted = ?", false).Update("password", pwdEncrypted).Error
	} else {
		err = Db.Model(&User{}).Where("uid = ?", username).Where("is_deleted = ?", false).Update("password", pwdEncrypted).Error
	}
	if err != nil {
		return err
	}
	return nil
}

// CheckUserByEmail find user by email
// return true if user exist
func CheckUserByEmail(email string) (bool, error) {
	var user User
	err := Db.Where("email = ?", email).Where("is_deleted = ?", false).First(&user).Error
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
	err := Db.Where("uid = ?", uid).Where("is_deleted = ?", false).First(&user).Error
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
	if err := Db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error; err != nil {
		return nil, fmt.Errorf("%v: User [%s] Not Exist\n", err, username)
	}

	return &user, nil
}

func GenerateVerifyCode() string {
	code := util.GenerateCode()
	return code
}

func SendEmail(recipient, content string) error {
	emailInfo := conf.Sub("email")
	sender := emailInfo.GetString("sender")
	secret := emailInfo.GetString("secret")
	return util.SendEmail(sender, secret, recipient, content)
}
