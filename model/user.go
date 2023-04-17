package model

import (
	"time"
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
