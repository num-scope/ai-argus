package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"time"

	"github.com/xtj/ai-argus/internal/dto"
)

//go:embed templates/*.html static/*
var files embed.FS

func Templates() (*template.Template, error) {
	functions := template.FuncMap{
		"formatTime": func(value time.Time) string {
			if value.IsZero() {
				return "-"
			}
			return value.Local().Format("01-02 15:04:05")
		},
		"statusLabel": statusLabel,
		"statusClass": statusClass,
		"percent": func(value float64) string {
			return fmt.Sprintf("%.1f%%", value)
		},
		"decimal": func(value float64) string {
			return fmt.Sprintf("%.2f", value)
		},
		"formatDuration": formatDuration,
		"ms":             formatLatencyMS,
		"derefFloat": func(value *float64) float64 {
			if value == nil {
				return 0
			}
			return *value
		},
		"derefInt": func(value *int) int {
			if value == nil {
				return 0
			}
			return *value
		},
		"maxFloat": func(values ...float64) float64 {
			maximum := 1.0
			for _, value := range values {
				if value > maximum {
					maximum = value
				}
			}
			return maximum
		},
		"maxInt": func(values ...int) int {
			maximum := 1
			for _, value := range values {
				if value > maximum {
					maximum = value
				}
			}
			return maximum
		},
		"ratio": func(value, total float64) float64 {
			if total <= 0 {
				return 0
			}
			return min(max(value/total*100, 0), 100)
		},
		"ratioInt": func(value, total int) float64 {
			if total <= 0 {
				return 0
			}
			return min(max(float64(value)/float64(total)*100, 0), 100)
		},
		"resultMaxMS": func(results []dto.RequestResultResponse) float64 {
			maximum := 1.0
			for _, result := range results {
				if result.ElapsedMS > maximum {
					maximum = result.ElapsedMS
				}
			}
			return maximum
		},
		// recentResults returns the last n results (for charts when full log is loaded ASC).
		"recentResults": func(results []dto.RequestResultResponse, n int) []dto.RequestResultResponse {
			if n <= 0 || len(results) <= n {
				return results
			}
			return results[len(results)-n:]
		},
		"joinLines": func(values []string) string {
			result := ""
			for index, value := range values {
				if index > 0 {
					result += "\n"
				}
				result += value
			}
			return result
		},
	}
	return template.New("pages").Funcs(functions).ParseFS(files, "templates/*.html")
}

func StaticFS() (fs.FS, error) {
	return fs.Sub(files, "static")
}

// formatDuration renders a duration in seconds with adaptive units:
// sub-second -> ms, under 1 min -> s, under 1 hour -> min, else h / d.
func formatDuration(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	switch {
	case seconds == 0:
		return "0 s"
	case seconds < 1:
		return fmt.Sprintf("%.0f ms", seconds*1000)
	case seconds < 60:
		if seconds < 10 {
			return fmt.Sprintf("%.2f s", seconds)
		}
		return fmt.Sprintf("%.1f s", seconds)
	case seconds < 3600:
		minutes := seconds / 60
		if minutes < 10 {
			return fmt.Sprintf("%.1f min", minutes)
		}
		return fmt.Sprintf("%.0f min", minutes)
	case seconds < 86400:
		hours := seconds / 3600
		if hours < 10 {
			return fmt.Sprintf("%.1f h", hours)
		}
		return fmt.Sprintf("%.0f h", hours)
	default:
		days := seconds / 86400
		if days < 10 {
			return fmt.Sprintf("%.1f d", days)
		}
		return fmt.Sprintf("%.0f d", days)
	}
}

// formatLatencyMS formats a latency value stored in milliseconds with adaptive units.
// <1s stays in ms; ≥1s uses s; ≥1min uses min; ≥1h uses h.
func formatLatencyMS(ms float64) string {
	if ms <= 0 {
		return "-"
	}
	switch {
	case ms < 1000:
		return fmt.Sprintf("%.1f ms", ms)
	case ms < 60_000:
		seconds := ms / 1000
		if seconds < 10 {
			return fmt.Sprintf("%.2f s", seconds)
		}
		return fmt.Sprintf("%.1f s", seconds)
	case ms < 3_600_000:
		minutes := ms / 60_000
		if minutes < 10 {
			return fmt.Sprintf("%.1f min", minutes)
		}
		return fmt.Sprintf("%.0f min", minutes)
	default:
		hours := ms / 3_600_000
		if hours < 10 {
			return fmt.Sprintf("%.1f h", hours)
		}
		return fmt.Sprintf("%.0f h", hours)
	}
}

func statusLabel(status string) string {
	switch status {
	case "queued":
		return "排队中"
	case "warming":
		return "预热中"
	case "running":
		return "运行中"
	case "completed":
		return "已完成"
	case "failed":
		return "失败"
	case "cancelled":
		return "已停止"
	default:
		return status
	}
}

func statusClass(status string) string {
	switch status {
	case "completed":
		return "good"
	case "failed":
		return "bad"
	case "cancelled":
		return "muted"
	default:
		return "live"
	}
}
