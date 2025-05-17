package main_test

import (
	"testing"

	"semyi"
)

func TestMigrate(t *testing.T) {
	for attempt := 1; attempt <= 3; attempt++ {
		err := main.Migrate(database, t.Context(), true)
		if err != nil {
			t.Errorf("failed to migrate database on attempt #%d: %v", attempt, err)
		} else {
			t.Log("database migrated successfully")
			break
		}
	}
}
