package config

import "testing"

func TestConfig(t *testing.T) {
	configs := []string{
		"database.host",
		"database.port",
	}

	for _, config := range configs {
		if Config.GetString(config) == "" {
			t.Errorf("Config: %s null\n", config)
		}
	}
}
