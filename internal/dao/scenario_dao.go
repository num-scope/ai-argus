package dao

import (
	"context"
	"errors"

	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/model"
	"gorm.io/gorm"
)

func CreateScenario(ctx context.Context, scenario *model.Scenario) error {
	return normalizeCreateError(database.DB.WithContext(ctx).Create(scenario).Error)
}

func GetScenarioByID(ctx context.Context, id int64) (*model.Scenario, error) {
	var scenario model.Scenario
	if err := database.DB.WithContext(ctx).First(&scenario, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &scenario, nil
}

func ListScenarios(ctx context.Context) ([]model.Scenario, error) {
	var scenarios []model.Scenario
	err := database.DB.WithContext(ctx).Order("updated_at DESC").Find(&scenarios).Error
	return scenarios, err
}

func ListScenariosPage(ctx context.Context, offset, limit int) ([]model.Scenario, int64, error) {
	var scenarios []model.Scenario
	tx := database.DB.WithContext(ctx).Model(&model.Scenario{})
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&scenarios).Error; err != nil {
		return nil, 0, err
	}
	return scenarios, total, nil
}

func DeleteScenario(ctx context.Context, id int64) error {
	result := database.DB.WithContext(ctx).Delete(&model.Scenario{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func CountScenarios(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.WithContext(ctx).Model(&model.Scenario{}).Count(&count).Error
	return count, err
}
