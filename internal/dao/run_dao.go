package dao

import (
	"context"
	"errors"
	"time"

	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/model"
	"gorm.io/gorm"
)

func CreateRun(ctx context.Context, run *model.Run) error {
	return database.DB.WithContext(ctx).Create(run).Error
}

func UpdateRun(ctx context.Context, run *model.Run) error {
	return updateRun(database.DB.WithContext(ctx), run)
}

func CancelRunIfActive(ctx context.Context, id int64, finishedAt time.Time, message string) (bool, error) {
	result := database.DB.WithContext(ctx).Model(&model.Run{}).
		Where("id = ? AND status IN ?", id, activeRunStatuses()).
		Updates(map[string]any{
			"status":        model.RunStatusCancelled,
			"error_message": message,
			"finished_at":   finishedAt,
		})
	return result.RowsAffected > 0, result.Error
}

func GetRunByID(ctx context.Context, id int64) (*model.Run, error) {
	var run model.Run
	if err := database.DB.WithContext(ctx).First(&run, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func ListRuns(ctx context.Context, limit int) ([]model.Run, error) {
	var runs []model.Run
	tx := database.DB.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	err := tx.Find(&runs).Error
	return runs, err
}

func CountRuns(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.WithContext(ctx).Model(&model.Run{}).Count(&count).Error
	return count, err
}

func CountActiveRuns(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.WithContext(ctx).Model(&model.Run{}).
		Where("status IN ?", activeRunStatuses()).
		Count(&count).Error
	return count, err
}

func FailInterruptedRuns(ctx context.Context) error {
	now := time.Now()
	return database.DB.WithContext(ctx).Model(&model.Run{}).
		Where("status IN ?", activeRunStatuses()).
		Updates(map[string]any{
			"status":        model.RunStatusFailed,
			"error_message": "服务重启导致任务中断",
			"finished_at":   now,
		}).Error
}

func AppendRunResults(ctx context.Context, run *model.Run, results []model.RequestResult) error {
	if len(results) == 0 {
		return UpdateRun(ctx, run)
	}
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(results, 50).Error; err != nil {
			return err
		}
		return updateRun(tx, run)
	})
}

func ListRequestResults(ctx context.Context, runID int64, limit int) ([]model.RequestResult, error) {
	var results []model.RequestResult
	tx := database.DB.WithContext(ctx).Where("run_id = ?", runID).Order("request_index DESC")
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	err := tx.Find(&results).Error
	return results, err
}

func ListRequestResultsASC(ctx context.Context, runID int64) ([]model.RequestResult, error) {
	var results []model.RequestResult
	err := database.DB.WithContext(ctx).Where("run_id = ?", runID).Order("request_index ASC").Find(&results).Error
	return results, err
}

func activeRunStatuses() []string {
	return []string{model.RunStatusQueued, model.RunStatusWarming, model.RunStatusRunning}
}

func runUpdateFields(run *model.Run) map[string]any {
	return map[string]any{
		"status":        run.Status,
		"summary_json":  run.SummaryJSON,
		"planned":       run.Planned,
		"completed":     run.Completed,
		"succeeded":     run.Succeeded,
		"failed":        run.Failed,
		"error_message": run.ErrorMessage,
		"started_at":    run.StartedAt,
		"finished_at":   run.FinishedAt,
	}
}

func updateRun(tx *gorm.DB, run *model.Run) error {
	result := tx.Model(&model.Run{}).
		Where("id = ?", run.ID).
		Updates(runUpdateFields(run))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}
