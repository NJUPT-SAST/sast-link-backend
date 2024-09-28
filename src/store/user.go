package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/util"

	"github.com/pkg/errors"
)

type User struct {
	ID        uint      `json:"id,omitempty" gorm:"primaryKey"`
	UID       *string   `json:"uid,omitempty" gorm:"not null"`
	Email     *string   `json:"email,omitempty" gorm:"not null"`
	Password  *string   `json:"password,omitempty" gorm:"not null"`
	QQID      *string   `json:"qq_id,omitempty"`
	LarkID    *string   `json:"lark_id,omitempty"`
	GithubID  *string   `json:"github_id,omitempty"`
	WechatID  *string   `json:"wechat_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"not null"`
	IsDeleted bool      `json:"is_deleted,omitempty" gorm:"not null"`
}

func (s *Store) CreateUserAndProfile(user *User, profile *Profile) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// create user and get user_id
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
	}
	return nil
}

// CheckPassword check user password, return uid if matched
//
// error:
// 1. User not exist
// 2. Password is incorrect.
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
			return "", errors.New("User not exist")
		}
	}

	// Encrypt and verify password
	pwdEncrypted := util.ShaHashing(password)
	if *user.Password != pwdEncrypted {
		return "", errors.New("Password is incorrect")
	}

	return *user.UID, nil
}

func (s *Store) ChangePassword(ctx context.Context, uid string, password string) error {
	pwdEncrypted := util.ShaHashing(password)
	err := s.db.WithContext(ctx).Model(&User{}).Where("uid = ?", uid).Where("is_deleted = ?", false).Update("password", pwdEncrypted).Error
	if err != nil {
		return err
	}
	return nil
}

// UserByField find user by specific database table field name
//
// If user not found, return nil.
func (s *Store) UserByField(ctx context.Context, field, value string) (*User, error) {
	var user User
	err := s.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", field), value).Where("is_deleted = ?", false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// userLogger.Errorf("User with [%s: %s] Not Exist\n", field, value)
			log.Debugf("User with [%s: %s] Not Exist\n", field, value)
			return nil, nil
		}
		log.Errorf("Failed to query user by field: %s\n", field)
		return nil, errors.Wrap(err, "failed to query user by field")
	}
	return &user, nil
}

// UserInfo returns the user information of the current user
//
// If user not found, return nil.
func (s *Store) UserInfo(ctx context.Context, username string) (*User, error) {
	var user = User{}
	var err error
	if strings.Contains(username, "@") {
		err = s.db.WithContext(ctx).Where("email = ?", username).Where("is_deleted = ?", false).First(&user).Error
	} else {
		err = s.db.WithContext(ctx).Where("uid = ?", username).Where("is_deleted = ?", false).First(&user).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("User [%s] Not Exist\n", username)
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

// TODO: Send email.
func (s *Store) SendEmail(ctx context.Context, recipient, content, title string) error {
	// FIXME: Get email sender and secret from system settings
	settings, err := s.GetSystemSetting(ctx, config.EmailSettingType.String())
	if err != nil {
		return err
	}

	email, err := settings.GetSettings()
	if err != nil {
		return err
	}
	emailInfo, ok := email.(EmailSetting)
	if !ok {
		return errors.New("failed to convert email settings to EmailSetting")
	}

	sender := emailInfo.Sender
	secret := emailInfo.Secret
	return util.SendEmail(sender, secret, recipient, content, title)
}
