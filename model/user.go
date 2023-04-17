package model

import (
	"time"
)

type User struct {
	ID        uint      `json:"id,omitempty" gorm:"primaryKey`
	Uid       *string   `json:"uid,omitempty"`
	Email     *string   `json:"email,omitempty"`
	QQId      *string   `json:"qq_id,omitempty"`
	LarkId    *string   `json:"lark_id,omitempty"`
	GithubId  *string   `json:"github_id,omitempty"`
	WechatId  *string   `json:"wechat_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	IsDeleted bool      `json:"is_deleted,omitempty"`
}

func Create()
