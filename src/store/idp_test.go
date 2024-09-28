package store

import (
	"context"
	"testing"

	"github.com/NJUPT-SAST/sast-link-backend/config"
)

func TestListIdentityProviders(t *testing.T) {
	if config.SetupConfig() != nil {
		t.Fatal("config.SetupConfig() failed")
	}
	instanceConfig := config.NewConfig()
	store, _ := NewStore(context.Background(), instanceConfig)

	idps, err := store.ListIdentityProviders(context.Background())
	if err != nil {
		t.Error(err)
	}
	for _, idp := range idps {
		t.Log(idp.Name)
		if idp.GetOauth2Setting() == nil {
			t.Error("GetOauth2Setting() is nil")
			continue
		}
		t.Log(idp.GetOauth2Setting().AuthURL)
		t.Log(idp.GetOauth2Setting().TokenURL)
	}
}
