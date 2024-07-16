package model

func UpdateLarkUserInfo(username, clientType, oauthID, larkUserInfo string) error {
	return Db.Table("oauth2_info").
		Where("user_id = ?", username).
		Where("client = ?", clientType).
		Update("oauth_user_id = ?", oauthID).
		Update("info = ?", larkUserInfo).Error
}
