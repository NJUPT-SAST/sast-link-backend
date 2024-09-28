package config

import "testing"

func TestLoadSystemSetting(t *testing.T) {
	if SetupConfig() != nil {
		t.Fatalf("SetupConfig() failed")
	}
	instanceConfig := NewConfig()
	instanceConfig.LoadSystemSettings()
}

// func TestConfig(t *testing.T) {
// 	configs := []string{
// 		"postgres.host",
// 		"postgres.port",
// 		"redis.password",
// 	}
//
// 	for _, config := range configs {
// 		if Config.GetString(config) == "" {
// 			t.Errorf("Config: %s null\n", config)
// 		}
// 	}
// }
