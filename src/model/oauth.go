package model

import (
	// "encoding/json"
	"errors"

	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OAuth2Info struct
type OAuth2Info struct {
	ID      uint           `gorm:"primaryKey"`
	Client  string         `gorm:"not null"`
	Info    datatypes.JSON `gorm:"default:'[]'"`
	OauthID string         `gorm:"not null"`
	UserID  string         `gorm:"not null"`
}

func UpdateLarkUserInfo(info OAuth2Info) error {
	return Db.Table("oauth2_info").
		Where("user_id = ?", info.UserID).
		Where("client = ?", info.Client).
		Update("oauth_user_id = ?", info.OauthID).
		Update("info = ?", info.Info).Error
}

// OauthInfoByUID find user by specific client id in oauth2_info table
func OauthInfoByUID(clientType, oauthUID string) (*OAuth2Info, error) {
	var client OAuth2Info
	err := Db.Table("oauth2_info").
		Where("oauth_user_id = ?", oauthUID).
		Where("client = ?", clientType).
		First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("model.OauthInfoByUID ::: %s", err.Error())
		return nil, result.InternalErr
	}
	return &client, nil
}

// UpsetOauthInfo insert or update oauth2_info table
func UpsetOauthInfo(oauthInfo OAuth2Info) {
	// return Db.Table("oauth2_info").Save(oauthInfo).Error
	stmt := `
	       	INSERT INTO oauth2_info (client, info, oauth_user_id, user_id)
	       	VALUES (?, ?, ?, ?)
	       	ON CONFLICT (client, user_id) DO UPDATE
	       	SET info = EXCLUDED.info, oauth_user_id = EXCLUDED.oauth_user_id, client = EXCLUDED.client
	`

	Db.Exec(stmt, oauthInfo.Client, oauthInfo.Info, oauthInfo.OauthID, oauthInfo.UserID)
}
