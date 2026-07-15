package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xtj/ai-argus/internal/dao"
	"github.com/xtj/ai-argus/internal/dto"
)

func GetDashboard(ctx context.Context) (*dto.Dashboard, error) {
	targetCount, err := dao.CountTargets(ctx)
	if err != nil {
		return nil, err
	}
	scenarioCount, err := dao.CountScenarios(ctx)
	if err != nil {
		return nil, err
	}
	runCount, err := dao.CountRuns(ctx)
	if err != nil {
		return nil, err
	}
	activeCount, err := dao.CountActiveRuns(ctx)
	if err != nil {
		return nil, err
	}
	recent, err := dao.ListRuns(ctx, 8)
	if err != nil {
		return nil, err
	}
	dashboard := &dto.Dashboard{
		TargetCount:    targetCount,
		ScenarioCount:  scenarioCount,
		RunCount:       runCount,
		ActiveRunCount: activeCount,
		RecentRuns:     make([]dto.RunResponse, 0, len(recent)),
		UpdatedAt:      time.Now(),
	}
	for _, run := range recent {
		dashboard.RecentRuns = append(dashboard.RecentRuns, toRunResponse(run))
	}
	for _, run := range recent {
		if run.SummaryJSON == "" || run.SummaryJSON == "{}" {
			continue
		}
		var summary dto.RunSummary
		if err := json.Unmarshal([]byte(run.SummaryJSON), &summary); err != nil {
			return nil, fmt.Errorf("解析运行 %d 汇总: %w", run.ID, err)
		}
		dashboard.LastSuccessRate = summary.SuccessRate
		dashboard.LastRPS = summary.RPS
		break
	}
	return dashboard, nil
}
