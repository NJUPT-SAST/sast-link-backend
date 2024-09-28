package store

import (
	"errors"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Profile struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    *uint          `json:"user_id" gorm:"not null"`
	Nickname  *string        `json:"nickname" gorm:"not null"`
	Email     *string        `json:"email" gorm:"not null"`
	IsDeleted bool           `json:"is_deleted" gorm:"not null"`
	Avatar    *string        `json:"avatar" gorm:"not null"`
	OrgID     int            `json:"org_id"`
	Bio       *string        `json:"bio"`
	Link      pq.StringArray `json:"link" gorm:"type:varchar[]"`
	Badge     *badge         `json:"badge,omitempty" gorm:"type:json"`
	Hide      pq.StringArray `json:"hide,omitempty" gorm:"type:varchar(30)[]"`
}

type Organize struct {
	ID  uint   `json:"id" gorm:"primaryKey"`
	Dep string `json:"dep" gorm:"not null"`
	Org string `json:"org"`
}

func (s *Store) UpdateAvatar(avatar string, userID uint) error {
	if err := s.db.Table("profile").Where("profile.user_id = ?", userID).Update("avatar", avatar).Error; err != nil {
		return err
	}
	return nil
}

func (s *Store) SelectProfileByUID(uid string) (*Profile, error) {
	var profile Profile
	err := s.db.Table("profile").
		Joins(`left join "user" u on profile.user_id=u.id AND profile.is_deleted = ?`, false).
		Where("u.uid = ? AND u.is_deleted = ?", uid, false).
		First(&profile).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (s *Store) UpdateProfile(oldProfile, newProfile *Profile) error {
	newProfile.ID = oldProfile.ID
	if err := s.db.Debug().Table("profile").Model(&Profile{}).Where("profile.id = ?", oldProfile.ID).Updates(newProfile).Error; err != nil {
		return err
	}
	return nil
}

func (s *Store) GetDepAndOrgByOrgID(orgID int) (string, string, error) {
	var res Organize
	if err := s.db.Table("organize").Select("dep,org").Where("id = ?", orgID).Find(&res).Error; err != nil {
		return "", "", err
	}
	return res.Dep, res.Org, nil
}
