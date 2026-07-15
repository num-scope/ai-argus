package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xtj/ai-argus/internal/benchmark"
	"github.com/xtj/ai-argus/internal/dao"
	"github.com/xtj/ai-argus/internal/model"
	"github.com/xtj/ai-argus/internal/protocol"
	"github.com/xtj/ai-argus/pkg/logger"
	"go.uber.org/zap"
)

const resultBatchSize = 25
const persistenceTimeout = 5 * time.Second
const finalPersistenceAttempts = 3

var runs = struct {
	sync.Mutex
	cancels        map[int64]context.CancelFunc
	wait           sync.WaitGroup
	maxConcurrency int
	rootContext    context.Context
}{cancels: make(map[int64]context.CancelFunc), maxConcurrency: 1000, rootContext: context.Background()}

func ConfigureRuns(ctx context.Context, maxConcurrency int) {
	runs.Lock()
	defer runs.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	runs.rootContext = ctx
	runs.maxConcurrency = maxConcurrency
}

func RunConcurrencyLimit() int {
	runs.Lock()
	defer runs.Unlock()
	return runs.maxConcurrency
}

func ShutdownRuns(ctx context.Context) error {
	runs.Lock()
	for _, cancel := range runs.cancels {
		cancel()
	}
	runs.Unlock()
	done := make(chan struct{})
	go func() {
		runs.wait.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func executeRun(ctx context.Context, run *model.Run, cfg benchmark.Config) {
	defer func() {
		runs.Lock()
		delete(runs.cancels, run.ID)
		runs.Unlock()
		runs.wait.Done()
	}()
	defer func() {
		if recovered := recover(); recovered != nil {
			finishRunWithError(ctx, run, fmt.Errorf("任务异常: %v", recovered))
		}
	}()

	pending := make([]model.RequestResult, 0, resultBatchSize)
	lastFlush := time.Now()
	var liveCollector *benchmark.Collector
	flush := func() error {
		if len(pending) == 0 {
			return nil
		}
		batch := append([]model.RequestResult(nil), pending...)
		persistCtx, cancel := persistenceContext(ctx)
		defer cancel()
		if err := dao.AppendRunResults(persistCtx, run, batch); err != nil {
			return err
		}
		pending = pending[:0]
		lastFlush = time.Now()
		return nil
	}

	summary, err := benchmark.Run(ctx, cfg, func(status string) error {
		now := time.Now()
		if run.StartedAt == nil {
			run.StartedAt = &now
		}
		run.Status = status
		if status == model.RunStatusRunning {
			liveCollector = benchmark.NewCollector(now)
		}
		persistCtx, cancel := persistenceContext(ctx)
		defer cancel()
		return dao.UpdateRun(persistCtx, run)
	}, func(result protocol.Result) error {
		if liveCollector != nil {
			liveCollector.Add(result)
			if liveJSON, liveErr := json.Marshal(liveCollector.Summary(time.Now())); liveErr == nil {
				run.SummaryJSON = string(liveJSON)
			}
		}
		pending = append(pending, toRequestResult(run.ID, result))
		run.Completed++
		if result.OK {
			run.Succeeded++
		} else {
			run.Failed++
		}
		if len(pending) >= resultBatchSize || time.Since(lastFlush) >= time.Second {
			return flush()
		}
		return nil
	})
	if flushErr := flush(); flushErr != nil && err == nil {
		err = flushErr
	}

	finishedAt := time.Now()
	summaryJSON, marshalErr := json.Marshal(summary)
	if marshalErr == nil {
		run.SummaryJSON = string(summaryJSON)
	}
	run.FinishedAt = &finishedAt
	switch {
	case errors.Is(err, context.Canceled):
		run.Status = model.RunStatusCancelled
		run.ErrorMessage = "任务已手动停止"
	case err != nil:
		run.Status = model.RunStatusFailed
		run.ErrorMessage = err.Error()
	default:
		run.Status = model.RunStatusCompleted
		run.ErrorMessage = ""
	}
	if updateErr := persistRunWithRetry(ctx, run); updateErr != nil {
		logger.L().Error("finish run", zap.Int64("run_id", run.ID), zap.Error(updateErr))
	}
}

func finishRunWithError(ctx context.Context, run *model.Run, err error) {
	now := time.Now()
	run.Status = model.RunStatusFailed
	run.ErrorMessage = err.Error()
	run.FinishedAt = &now
	if updateErr := persistRunWithRetry(ctx, run); updateErr != nil {
		logger.L().Error("persist crashed run", zap.Int64("run_id", run.ID), zap.Error(updateErr))
	}
}

func persistenceContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), persistenceTimeout)
}

func persistRunWithRetry(ctx context.Context, run *model.Run) error {
	var lastErr error
	for attempt := 0; attempt < finalPersistenceAttempts; attempt++ {
		persistCtx, cancel := persistenceContext(ctx)
		lastErr = dao.UpdateRun(persistCtx, run)
		cancel()
		if lastErr == nil {
			return nil
		}
		if attempt+1 < finalPersistenceAttempts {
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}
	return lastErr
}
