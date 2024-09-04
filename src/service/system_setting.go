package service

import (
	"context"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/store"
)

type SysSettingService struct {
	*BaseService
}

func NewSysSettingService(store *BaseService) *SysSettingService {
	return &SysSettingService{store}
}

func (s *SysSettingService) GetSysSetting(ctx context.Context, settingName string) (interface{}, error) {
	systemSetting, err := s.Store.GetSystemSetting(ctx, settingName)
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

func (s *SysSettingService) ListIDPName(ctx context.Context) ([]string, error) {
	idps, err := s.Store.ListIdentityProviders(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, idp := range idps {
		names = append(names, idp.Name)
	}

	return names, nil
}
