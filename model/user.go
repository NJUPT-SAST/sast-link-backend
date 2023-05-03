package model

import (
	"fmt"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/util"
	"gorm.io/gorm"
)

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
	err := db.Select("email").Where("email = ?", username).First(&user).Error
	if err != nil && gorm.ErrRecordNotFound != err {
		return isExist, ticket, err
	}
	// if user != null
	if user != (User{}) {
		isExist = true
	}

	if isExist {
		ticket, err = util.GenerateToken(username)
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
