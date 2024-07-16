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

func CreateUserAndProfile(user *User, profile *Profile) error {
	err := Db.Transaction(func(tx *gorm.DB) error {
		//create user and get user_id
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UserID = &user.ID

		tx = tx.Table("profile")
		if err := tx.Create(profile).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	} else {
		return nil
	}
}

func CheckPassword(username string, password string) (string, error) {
	//get uid from username by regexp
	var user User
	matched, regErr := regexp.MatchString("@", username)
	if regErr != nil {
		userLogger.Errorf("regexp matchiong error")
		return "", regErr
	}

	var err error = nil
	// Get user by email/uid
	// If matched, get user by email
	if matched {
		err = Db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	} else {
		err = Db.Where("uid = ?", username).Where("is_deleted = ?", false).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Errorf("User [%s] Not Exist\n", username)
			return "", result.UserNotExist
		}
	}

	//encrypt and verify password
	pwdEncrypted := util.ShaHashing(password)
	if *user.Password != pwdEncrypted {
		return "", result.PasswordError
	}
	return *user.Uid, err
}

func ChangePassword(uid string, password string) error {
	pwdEncrypted := util.ShaHashing(password)
	err := Db.Debug().Model(&User{}).Where("uid = ?", uid).Where("is_deleted = ?", false).Update("password", pwdEncrypted).Error
	if err != nil {
		return err
	}
	return nil
}

// UserByField find user by specific database table field name
func UserByField(field, value string) (*User, error) {
	var user User
	err := Db.Where(fmt.Sprintf("%s = ?", field), value).Where("is_deleted = ?", false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Errorf("User with [%s: %s] Not Exist\n", field, value)
			return nil, nil
		}
		return nil, result.InternalErr
	}
	return &user, nil
}

func UserInfo(username string) (*User, error) {
	var user = User{Uid: &username}
	matched, err2 := regexp.MatchString("@", username)
	if err2 != nil {
		userLogger.Errorf("regexp matchiong error")
		return nil, err2
	}
	var err error = nil
	if matched {
		err = Db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	} else {
		err = Db.Where("uid = ?", username).Where("is_deleted = ?", false).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Errorf("User [%s] Not Exist\n", username)
			return nil, err
		}
	}
	return &user, nil
}

func GenerateVerifyCode() string {
	code := util.GenerateCode()
	return code
}

func SendEmail(recipient, content, title string) error {
	emailInfo := conf.Sub("email")
	sender := emailInfo.GetString("sender")
	secret := emailInfo.GetString("secret")
	return util.SendEmail(sender, secret, recipient, content, title)
}
