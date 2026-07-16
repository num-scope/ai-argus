package benchmark

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xtj/ai-argus/internal/prompt"
	"github.com/xtj/ai-argus/internal/protocol"
)

type Config struct {
	Protocol                protocol.Config `json:"target"`
	Prompts                 []string        `json:"prompts"`
	Concurrency             int             `json:"concurrency"`
	TotalRequests           int             `json:"total_requests"`
	WarmupRequests          int             `json:"warmup_requests"`
	RampUpSeconds           float64         `json:"ramp_up_seconds"`
	RandomPromptMode        bool            `json:"random_prompt_mode"`
	RandomPromptTargetChars int             `json:"random_prompt_target_chars"`
	RandomPromptMaxChars    int             `json:"random_prompt_max_chars"`
}

type PhaseCallback func(status string) error
type ResultCallback func(result protocol.Result) error

type scheduledJob struct {
	index    int
	prompt   string
	queuedAt time.Time
}

func Run(ctx context.Context, cfg Config, onPhase PhaseCallback, onResult ResultCallback) (Summary, error) {
	if err := validateConfig(cfg); err != nil {
		return Summary{}, err
	}
	client := protocol.NewClient(cfg.Protocol, cfg.Concurrency)
	defer client.Close()

	if cfg.WarmupRequests > 0 {
		if err := onPhase("warming"); err != nil {
			return Summary{}, err
		}
		failed := 0
		firstError := ""
		err := execute(ctx, client, cfg, min(cfg.Concurrency, cfg.WarmupRequests), cfg.WarmupRequests, 0, func(result protocol.Result) error {
			if !result.OK {
				failed++
				if firstError == "" {
					firstError = result.Error
				}
			}
			return nil
		})
		if err != nil {
			return Summary{}, err
		}
		if failed > 0 {
			return Summary{}, fmt.Errorf("预热失败 %d/%d，首个错误：%s", failed, cfg.WarmupRequests, firstError)
		}
	}

	if err := onPhase("running"); err != nil {
		return Summary{}, err
	}
	startedAt := time.Now()
	collector := NewCollector(startedAt)
	err := execute(ctx, client, cfg, cfg.Concurrency, cfg.TotalRequests, cfg.RampUpSeconds, func(result protocol.Result) error {
		collector.Add(result)
		return onResult(result)
	})
	summary := collector.Summary(time.Now())
	if err != nil {
		return summary, err
	}
	return summary, nil
}

func execute(ctx context.Context, client *protocol.Client, cfg Config, concurrency, total int, rampUpSeconds float64, callback ResultCallback) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	workerCount := concurrency
	if total > 0 {
		workerCount = min(workerCount, total)
	}
	results := make(chan protocol.Result, max(workerCount, 1))
	jobs := make(chan scheduledJob, max(workerCount, 1))
	var workers sync.WaitGroup

	go func() {
		defer close(results)
		interval := time.Duration(0)
		if rampUpSeconds > 0 && workerCount > 1 {
			interval = time.Duration(rampUpSeconds * float64(time.Second) / float64(workerCount-1))
		}
		for worker := 0; worker < workerCount; worker++ {
			if worker > 0 && interval > 0 {
				if err := wait(runCtx, interval); err != nil {
					break
				}
			}
			if runCtx.Err() != nil {
				break
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				for job := range jobs {
					select {
					case <-runCtx.Done():
						return
					default:
					}
					result := client.Execute(runCtx, job.index, job.prompt, time.Since(job.queuedAt))
					if runCtx.Err() != nil {
						return
					}
					select {
					case results <- result:
					case <-runCtx.Done():
						return
					}
				}
			}()
		}
		workers.Wait()
	}()

	go func() {
		defer close(jobs)
		for index := 1; total <= 0 || index <= total; index++ {
			job := scheduledJob{
				index:    index,
				prompt:   pickPrompt(cfg, index),
				queuedAt: time.Now(),
			}
			select {
			case jobs <- job:
			case <-runCtx.Done():
				return
			}
		}
	}()

	for result := range results {
		if err := callback(result); err != nil {
			cancel()
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func pickPrompt(cfg Config, index int) string {
	if cfg.RandomPromptMode {
		return prompt.Generate(cfg.RandomPromptTargetChars, cfg.RandomPromptMaxChars)
	}
	if len(cfg.Prompts) == 0 {
		return ""
	}
	return cfg.Prompts[(index-1)%len(cfg.Prompts)]
}

func validateConfig(cfg Config) error {
	if cfg.Concurrency < 1 {
		return errors.New("并发数必须大于 0")
	}
	if cfg.TotalRequests < 0 || cfg.WarmupRequests < 0 || cfg.RampUpSeconds < 0 {
		return errors.New("请求数、预热数和升压时间不能小于 0")
	}
	if !cfg.RandomPromptMode && len(cfg.Prompts) == 0 {
		return errors.New("至少需要一个提示词")
	}
	if cfg.RandomPromptMode {
		if cfg.RandomPromptTargetChars < 1 || cfg.RandomPromptMaxChars < cfg.RandomPromptTargetChars {
			return errors.New("随机提示词长度配置无效")
		}
	}
	return nil
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
