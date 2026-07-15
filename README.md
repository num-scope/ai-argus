# AI Argus

AI Argus 是一个面向大模型推理接口的性能观测平台。它把一次性并发脚本升级为可配置、可复用、可追溯的 Web 工作台，覆盖从请求建立到最后一个 Token 的关键性能信号。

平台基于 `/Users/c/Documents/xtj/ai-scripts` 的并发压测能力设计，首版支持 OpenAI Chat Completions 兼容协议与 Xin Buddy 风格的自定义兼容协议。所有页面由 Go `html/template` 服务端渲染，不需要 Node.js 或前端构建链。

## 已实现能力

- 推理目标：接口地址、模型、Bearer API Key、OpenAI / 自定义协议、额外请求头与请求体
- 自定义兼容：可配置顶层内容字段，识别 `reasoning`、`delta`、`done`、`error` 命名 SSE 事件
- 评测场景：并发、请求总量、无限持续运行、预热、渐进升压、生成参数、Seed、多提示词循环
- 网络策略：连接超时、完整流超时、可重试状态码、指数退避、连接池复用
- 响应处理：OpenAI `data:` SSE、自定义命名 SSE、非流式 JSON、`[DONE]`、`finish_reason`、业务错误
- 性能指标：RPS、成功率、E2E、Queue、HTTP、TTFT、TPOT、P50/P95/P99、Token 吞吐与 Usage 覆盖
- 实施日志：逐请求 HTTP 状态、尝试次数、流完整性、提示词、回答预览和错误原因
- 报告：运行上下文快照、延迟分位、Token 画像，并支持浏览器打印或导出 PDF
- 持久化：SQLite WAL、任务重启中断标记、目标/场景/运行/请求结果历史

## 项目结构

仓库采用 `go-web` Structure A 单应用分层：

```text
api/                 Gin 路由、模板处理器与 v1 JSON API
cmd/server/          服务启动和优雅关闭
config/              Viper 环境配置加载、归一化与校验
database/            GORM、SQLite 与 SQL 日志适配
internal/benchmark/  并发调度和指标聚合
internal/protocol/   OpenAI / 自定义协议客户端与 SSE 解析
internal/service/    业务编排
internal/dao/        数据访问
internal/model/      持久化模型
internal/dto/        请求与响应结构
migrations/          数据库迁移
pkg/                 Zap 日志与统一 API 响应
web/                 嵌入式模板、CSS 与原生 JavaScript
```

依赖方向保持为 `cmd -> api -> service -> dao -> database`，协议调用和并发执行位于独立的基础能力包中，HTTP 处理器不直接访问数据库。

## 本地启动

需要 Go 1.25 或更高版本：

```bash
cp .env.example .env
set -a && source .env && set +a
go run ./cmd/server
```

打开 [http://127.0.0.1:8080](http://127.0.0.1:8080)。首次启动会自动创建 `data/ai-argus.db` 并执行迁移。

也可以直接运行：

```bash
make run
make test
make build
```

## 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `ARGUS_ADDRESS` | `127.0.0.1:8080` | HTTP 监听地址 |
| `ARGUS_DATABASE_PATH` | `data/ai-argus.db` | SQLite 文件路径 |
| `ARGUS_LOG_LEVEL` | `info` | `debug`、`info`、`warn` 或 `error` |
| `ARGUS_LOG_FORMAT` | `console` | `console` 可读文本或 `json` 结构化日志 |
| `ARGUS_GORM_LOG_LEVEL` | `warn` | `silent`、`error`、`warn` 或 `info` |
| `ARGUS_MAX_CONCURRENCY` | `1000` | 单场景允许的最大并发 |

## 协议说明

OpenAI 模式发送标准 `model`、`messages`、`temperature`、`top_p`、`max_tokens` 与 `stream` 字段，并解析标准 Chat Completions 流。

自定义模式在保留 OpenAI 字段的同时，将用户提示词写入可配置的顶层字段，默认是 `content`。这与原脚本的 Xin Buddy 请求结构兼容，也允许通过额外请求头和请求体适配同类接口。

敏感信息不会由页面或 JSON API 返回，SQLite 文件会被设置为当前用户可读写。当前版本没有内置登录系统，默认只监听本机；如果修改为公网监听，请放在具备身份认证、TLS、CSRF 防护和访问控制的反向代理之后。

## JSON API

- `GET /healthz`
- `GET /api/v1/targets?page=1&page_size=20`
- `GET /api/v1/scenarios?page=1&page_size=20`
- `GET /api/v1/runs/:id`
- `POST /api/v1/runs/:id/cancel`

Web 表单负责目标、场景和任务创建，运行详情 API 用于状态查询和后续自动化集成。
