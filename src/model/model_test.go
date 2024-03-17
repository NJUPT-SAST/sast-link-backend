package model

import "testing"

func TestDB(t *testing.T) {
	connectPostgreSQL()
	connectRedis()
}
