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

func CheckPassword(username string, password string) (string, error) {
	//get uid from username by regexp
	var user User
	matched, regErr := regexp.MatchString("@", username)
	if regErr != nil {
		userLogger.Infof("regexp matchiong error")
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
			userLogger.Infof("User [%s] Not Exist\n", username)
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

func ChangePassword(username string, password string) error {
	pwdEncrypted := util.ShaHashing(password)
	err := Db.Model(&User{}).Where("uid = ?", username).Where("is_deleted = ?", false).Update("password", pwdEncrypted).Error
	if err != nil {
		return err
	}
	return nil
}

// GetUserByEmail find user by email
func GetUserByEmail(email string) (*User, error) {
	var user User
	err := Db.Where("email = ?", email).Where("is_deleted = ?", false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Infof("User [%s] Not Exist\n", email)
			return nil, nil
		}
		return nil, result.InternalErr
	}
	return &user, nil
}

// CheckUserByUid find user by uid
// return true if user exist
func GetUserByUid(uid string) (*User, error) {
	var user User
	err := Db.Where("uid = ?", uid).Where("is_deleted = ?", false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Infof("User [%s] Not Exist\n", uid)
			return nil, result.UserNotExist
		}
		return nil, err
	}
	return &user, nil
}

func UserInfo(username string) (*User, error) {
	var user = User{Uid: &username}
	matched, err2 := regexp.MatchString("@", username)
	if err2 != nil {
		userLogger.Infof("regexp matchiong error")
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
			userLogger.Infof("User [%s] Not Exist\n", username)
			return nil, err
		}
	}
	return &user, nil
}

// Add github id to user
func UpdateGithubId(username string, githubId string) error {
	matched, err := regexp.MatchString("@", username)
	if err != nil {
		userLogger.Error("regexp matchiong error")
		return err
	}

	//get user by email/uid
	if matched {
		err = Db.Model(&User{}).Where("email = ?", username).Where("is_deleted = ?", false).Update("github_id", githubId).Error
	} else {
		err = Db.Model(&User{}).Where("uid = ?", username).Where("is_deleted = ?", false).Update("github_id", githubId).Error
	}
	if err != nil {
		return fmt.Errorf("add github id error: %s", err.Error())
	}
	return nil
}

// Find user by github id
// Use it need to check if the user is nil
// Since the RecordNotFound error is nil
func FindUserByGithubId(githubId string) (*User, error) {
	var user User
	err := Db.Where("github_id = ?", githubId).Where("is_deleted = ?", false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userLogger.Infof("User [%s] Not Exist\n", githubId)
			return nil, nil
		}
		return nil, err
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
