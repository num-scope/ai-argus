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
		"ms": func(value float64) string {
			if value == 0 {
				return "-"
			}
			return fmt.Sprintf("%.1f ms", value)
		},
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
