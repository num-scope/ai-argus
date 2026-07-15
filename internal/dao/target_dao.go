package dao

import (
	"context"
	"errors"

	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/model"
	"gorm.io/gorm"
)

func CreateTarget(ctx context.Context, target *model.Target) error {
	return normalizeCreateError(database.DB.WithContext(ctx).Create(target).Error)
}

func GetTargetByID(ctx context.Context, id int64) (*model.Target, error) {
	var target model.Target
	if err := database.DB.WithContext(ctx).First(&target, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &target, nil
}

func ListTargets(ctx context.Context) ([]model.Target, error) {
	var targets []model.Target
	err := database.DB.WithContext(ctx).Order("updated_at DESC").Find(&targets).Error
	return targets, err
}

func ListTargetsPage(ctx context.Context, offset, limit int) ([]model.Target, int64, error) {
	var targets []model.Target
	tx := database.DB.WithContext(ctx).Model(&model.Target{})
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&targets).Error; err != nil {
		return nil, 0, err
	}
	return targets, total, nil
}

func DeleteTarget(ctx context.Context, id int64) error {
	result := database.DB.WithContext(ctx).Delete(&model.Target{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func CountTargets(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.WithContext(ctx).Model(&model.Target{}).Count(&count).Error
	return count, err
}
