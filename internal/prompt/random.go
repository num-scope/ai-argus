package prompt

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	topics = []string{
		"限流", "连接池", "超时重试", "缓存击穿", "网关鉴权", "队列积压",
		"SSE 断流", "Token 计费", "灰度发布", "熔断降级", "日志采样", "指标上报",
	}
	roles = []string{"运维", "后端", "测试", "架构师", "值班同学"}
	tones = []string{"简洁", "分点", "偏实操", "带排查顺序"}
	formats = []string{"三点总结", "检查清单", "简短步骤", "风险与动作"}
	suffixes = []string{
		"控制在 80 字内", "给出优先动作", "避免空话", "适合值班场景",
	}
	templates = []string{
		"用{tone}方式说明{topic}的常见原因和{format}",
		"从{role}视角解释{topic}，并给{format}",
		"围绕{topic}写一段{tone}说明，要求{format}",
		"请用{tone}口吻讲清{topic}，输出{format}",
	}
)

// Generate builds a short Chinese prompt near the target length, capped by maxChars.
func Generate(targetChars, maxChars int) string {
	if targetChars < 1 {
		targetChars = 20
	}
	if maxChars < targetChars {
		maxChars = targetChars
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	best := candidate(rng)
	bestDelta := abs(len([]rune(best)) - targetChars)
	for i := 0; i < 4; i++ {
		next := candidate(rng)
		delta := abs(len([]rune(next)) - targetChars)
		if delta < bestDelta {
			best, bestDelta = next, delta
		}
	}
	runes := []rune(best)
	if len(runes) > maxChars {
		return string(runes[:maxChars])
	}
	return best
}

func candidate(rng *rand.Rand) string {
	template := templates[rng.Intn(len(templates))]
	replacer := strings.NewReplacer(
		"{topic}", topics[rng.Intn(len(topics))],
		"{role}", roles[rng.Intn(len(roles))],
		"{tone}", tones[rng.Intn(len(tones))],
		"{format}", formats[rng.Intn(len(formats))],
	)
	return fmt.Sprintf("%s，%s", replacer.Replace(template), suffixes[rng.Intn(len(suffixes))])
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
