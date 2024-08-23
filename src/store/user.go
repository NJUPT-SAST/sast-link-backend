package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/http/response"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"gorm.io/gorm"
)

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

func (store *Store) CreateUserAndProfile(user *User, profile *Profile) error {
	err := store.db.Transaction(func(tx *gorm.DB) error {
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

// CheckPassword check user password, return uid if matched
func (s *Store) CheckPassword(username string, password string) (string, error) {
	var user User

	var err error
	// Get user by email/uid
	// If matched, get user by email
	if strings.Contains(username, "@") {
		err = s.db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	} else {
		err = s.db.Where("uid = ?", username).Where("is_deleted = ?", false).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("User [%s] Not Exist\n", username)
			return "", response.UserNotExist
		}
	}

	// Encrypt and verify password
	pwdEncrypted := util.ShaHashing(password)
	if *user.Password != pwdEncrypted {
		return "", response.PasswordError
	}
	return *user.Uid, err
}

func (s *Store) ChangePassword(uid string, password string) error {
	pwdEncrypted := util.ShaHashing(password)
	err := s.db.Model(&User{}).Where("uid = ?", uid).Where("is_deleted = ?", false).Update("password", pwdEncrypted).Error
	if err != nil {
		return err
	}
	return nil
}

// UserByField find user by specific database table field name
func (s *Store) UserByField(field, value string) (*User, error) {
	var user User
	err := s.db.Where(fmt.Sprintf("%s = ?", field), value).Where("is_deleted = ?", false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// userLogger.Errorf("User with [%s: %s] Not Exist\n", field, value)
			log.Errorf("User with [%s: %s] Not Exist\n", field, value)
			return nil, nil
		}
		return nil, response.InternalErr
	}
	return &user, nil
}

func (s *Store) UserInfo(username string) (*User, error) {
	var user = User{Uid: &username}
	var err error
	if strings.Contains(username, "@") {
		err = s.db.Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	} else {
		err = s.db.Where("uid = ?", username).Where("is_deleted = ?", false).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("User [%s] Not Exist\n", username)
			return nil, err
		}
	}
	return &user, nil
}

func GenerateVerifyCode() string {
	code := util.GenerateCode()
	return code
}

// TODO: Send email
func (s *Store) SendEmail(ctx context.Context, recipient, content, title string) error {
	// FIXME: Get email sender and secret from system settings
	settings, err := s.GetSystemSetting(ctx, config.EmailSettingType)
	if err != nil {
		return err
	}

	email, err := settings.GetSettings()
	if err != nil {
		return err
	}
	emailInfo := email.(EmailSetting)

	sender := emailInfo.Sender
	secret := emailInfo.Secret
	return util.SendEmail(sender, secret, recipient, content, title)
}
