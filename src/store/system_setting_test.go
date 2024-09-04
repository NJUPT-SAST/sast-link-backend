package store

import (
	"context"
	"testing"

	"github.com/NJUPT-SAST/sast-link-backend/config"
)

func TestListSystemSetting(t *testing.T) {
	config.SetupConfig()
	instanceConfig := config.NewConfig()
	store, _ := NewStore(context.Background(), instanceConfig)
	settings, err := store.ListSystemSetting(context.Background())
	if err != nil {
		t.Error(err)
	}
	t.Log(settings)
}
