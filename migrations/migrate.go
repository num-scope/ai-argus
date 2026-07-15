package migrations

import (
	"fmt"

	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/internal/model"
)

func Run() error {
	if err := database.DB.AutoMigrate(
		&model.Target{},
		&model.Scenario{},
		&model.Run{},
		&model.RequestResult{},
	); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}
