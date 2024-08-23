package service

import (
	"context"

	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/NJUPT-SAST/sast-link-backend/config"
)

type SysSettingService struct {
	*BaseService
}

func NewSysSettingService(store *BaseService) *SysSettingService {
	return &SysSettingService{store}
}

func (s *SysSettingService) GetSysSetting(ctx context.Context, settingType config.SystemSettingType) (interface{}, error) {
	systemSetting, err := s.Store.GetSystemSetting(ctx, settingType)
	if err != nil {
		return nil, err
	}
	return systemSetting, nil
}

func (s *SysSettingService) UpSetSysSetting(ctx context.Context, settingType config.SystemSettingType, settingValue interface{}) error {
	systemSetting := &store.SystemSetting{
		Name:  settingType.String(),
		Value: settingValue.(string),
	}
	if err := s.Store.UpsetSystemSetting(ctx, systemSetting); err != nil {
		return err
	}
	return nil
}
