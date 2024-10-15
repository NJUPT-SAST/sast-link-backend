package service

import (
	"context"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/store"
)

type SysSettingService struct {
	*BaseService
}

func NewSysSettingService(base *BaseService) *SysSettingService {
	return &SysSettingService{base}
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

func (s *SysSettingService) IDPInfo(ctx context.Context, idp string) (map[string]interface{}, error) {
	idpInfo, err := s.Store.GetIdentityProviderByName(ctx, idp)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"auth_url":  idpInfo.GetOauth2Setting().AuthURL,
		"client_id": idpInfo.GetOauth2Setting().ClientID,
		"scopes":    idpInfo.GetOauth2Setting().Scopes,
	}, nil
}
