package store

import (
	"context"
	"testing"

	"github.com/NJUPT-SAST/sast-link-backend/config"
)

func TestListSystemSetting(t *testing.T) {
	if config.SetupConfig() != nil {
		t.Fatal("config setup failed")
	}
	instanceConfig := config.NewConfig()
	store, _ := NewStore(context.Background(), instanceConfig)
	settings, err := store.ListSystemSetting(context.Background())
	if err != nil {
		t.Error(err)
	}
	t.Log(settings)
}
