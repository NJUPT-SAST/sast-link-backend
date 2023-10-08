package model

import (
	"errors"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var profileLogger = log.Log

type Profile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    *uint          `json:"user_id" gorm:"not null"`
	Nickname  *string        `json:"nickname" gorm:"not null"`
	Email     *string        `json:"email" gorm:"not null"`
	IsDeleted bool           `json:"is_deleted" gorm:"not null"`
	Avatar    *string        `json:"avatar" gorm:"not null"`
	OrgId     int            `json:"org_id"`
	Bio       *string        `json:"bio"`
	Link      pq.StringArray `json:"link" gorm:"type:varchar[]"`
	Badge     *badge         `json:"badge,omitempty" gorm:"type:json"`
	Hide      pq.StringArray `json:"hide,omitempty" gorm:"type:varchar(30)[]"`
}

type Organize struct {
	Id  uint   `json:"id" gorm:"primaryKey"`
	Dep string `json:"dep" gorm:"not null"`
	Org string `json:"org"`
}

func SelectProfileByUid(uid string) (*Profile, error) {
	var profile Profile
	err := Db.Table("profile").
		Joins(`left join "user" u on profile.user_id=u.id AND profile.is_deleted = ?`, false).
		Where("u.uid = ? AND u.is_deleted = ?", uid, false).
		First(&profile).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			profileLogger.Infof("Profile [%s] Not Exist\n", uid)
			return nil, nil
		}
		profileLogger.Errorln("check profile by uid err", err)
		return nil, err
	}
	return &profile, nil
}

func UpdateProfile(oldProfile, newProfile *Profile) error {
	newProfile.ID = oldProfile.ID
	if err := Db.Table("profile").Model(&Profile{}).Where("profile.id = ?", oldProfile.ID).Updates(newProfile).Error; err != nil {
		return err
	}
	return nil
}

func GetDepAndOrgByOrgId(orgId int) (string, string, error) {
	var res Organize
	if err := Db.Table("organize").Select("dep,org").Where("id = ?", orgId).Find(&res).Error; err != nil {
		profileLogger.Errorln("select profile by org_id err", err)
		return "", "", nil
	}
	return res.Dep, res.Org, nil
}
