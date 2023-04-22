package config

import "testing"

func TestConfig(t *testing.T) {
	configs := []string{
		"postgres.host",
		"postgres.port",
		"redis.localhost.addr",
	}

	for _, config := range configs {
		if Config.GetString(config) == "" {
			t.Errorf("Config: %s null\n", config)
		}
	}
}
