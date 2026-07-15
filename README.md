# AI Argus

AI Argus 是一个面向大语言模型推理接口的开源性能观测平台。它将目标接口、负载场景、运行记录和分析报告集中到一个 Web 工作台中，帮助开发者持续评估接口延迟、吞吐、稳定性与 Token 效率。

项目支持 OpenAI Chat Completions 兼容接口，也支持通过自定义请求字段、请求头、请求体和 SSE 事件适配其他流式推理服务。页面由 Go 服务端渲染，HTMX 负责局部更新，Flowbite 提供交互组件，ECharts 负责数据图表。构建后的静态资源会嵌入 Go 二进制文件，生产环境不需要 Node.js。

## 核心能力

- **目标管理**：配置接口地址、模型、Bearer API Key、协议类型以及额外请求头和请求体。
- **场景管理**：设置并发数、请求总量、预热、渐进升压、生成参数、超时、重试和提示词集合。
- **协议适配**：解析标准 `data:` SSE、命名 SSE、非流式 JSON、`[DONE]` 和 `finish_reason`。
- **性能分析**：统计 RPS、成功率、E2E、Queue、HTTP、TTFT、TPOT、P50/P95/P99 和 Token 吞吐。
- **运行诊断**：记录逐请求状态、尝试次数、流完整性、响应预览和错误原因。
- **可视化报告**：展示延迟分布、关键阶段对比、Token 构成和吞吐指标，支持浏览器打印或导出 PDF。
- **数据持久化**：使用 SQLite WAL 保存目标、场景、运行快照和请求结果，并识别服务重启时中断的任务。

## 快速开始

直接运行已构建的静态资源只需要 Go 1.25 或更高版本。默认配置无需额外文件：

```bash
go run ./cmd/server
```

服务默认监听 [http://127.0.0.1:8080](http://127.0.0.1:8080)。首次启动时会创建 SQLite 数据库并自动执行迁移。

常用开发命令：

```bash
make run
make test
make build
```

`make build` 会先构建前端资源，需要 Node.js 20 或更高版本，并在 Linux、macOS 和 Windows 上将可执行文件输出到 `bin/`。如果环境中没有 Make，也可以基于已生成的静态资源直接使用 Go 命令：

```bash
go test ./...
go build -o ai-argus ./cmd/server
```

修改前端源码需要 Node.js 20 或更高版本：

```bash
npm ci
npm run build
```

前端构建使用 Tailwind CSS、Flowbite、HTMX、模块化 ECharts 和 esbuild，产物写入 `web/static/dist/`。

### Docker

```bash
docker build -f deploy/Dockerfile -t ai-argus .
docker run --rm -p 8080:8080 -v ai-argus-data:/app/data ai-argus
```

Docker 示例使用命名卷保存数据库，不依赖宿主机的固定目录。

## 配置

应用通过环境变量读取配置；`.env.example` 仅作为变量清单，不会被程序自动加载。可以由 Shell、容器平台或进程管理器注入这些变量。

Linux/macOS 示例：

```bash
ARGUS_ADDRESS=:8080 ARGUS_LOG_FORMAT=json go run ./cmd/server
```

Windows PowerShell 示例：

```powershell
$env:ARGUS_ADDRESS = ":8080"
$env:ARGUS_LOG_FORMAT = "json"
go run ./cmd/server
```

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `ARGUS_ADDRESS` | `127.0.0.1:8080` | HTTP 监听地址 |
| `ARGUS_DATABASE_PATH` | `data/ai-argus.db` | SQLite 数据库文件，相对路径基于启动目录解析 |
| `ARGUS_LOG_LEVEL` | `info` | `debug`、`info`、`warn` 或 `error` |
| `ARGUS_LOG_FORMAT` | `console` | `console` 可读文本或 `json` 结构化日志 |
| `ARGUS_GORM_LOG_LEVEL` | `warn` | `silent`、`error`、`warn` 或 `info` |
| `ARGUS_MAX_CONCURRENCY` | `1000` | 单个场景允许的最大并发数 |

## 使用流程

1. 在“推理目标”中添加待测试接口及鉴权信息。
2. 在“评测场景”中配置负载、提示词和生成参数。
3. 在“运行记录”中选择目标与场景并启动任务。
4. 查看实时指标、逐请求日志和最终分析报告。

运行开始后会保存目标与场景快照，因此后续修改配置不会改变历史报告的上下文。

## 协议兼容

OpenAI 模式发送 `model`、`messages`、`temperature`、`top_p`、`max_tokens` 和 `stream` 等 Chat Completions 字段，并解析标准流式或非流式响应。

自定义模式保留标准字段，同时可将用户提示词写入指定的顶层字段，默认字段名为 `content`。还可以添加供应商要求的请求头和请求体字段，并解析 `reasoning`、`delta`、`done`、`error` 等命名 SSE 事件。

## 指标口径

| 指标 | 含义 |
| --- | --- |
| `RPS` | 统计周期内每秒完成的请求数 |
| `E2E` | 请求从开始执行到最终结束的总耗时，包含重试和退避等待 |
| `Queue` | 请求进入调度器后等待执行的时间 |
| `HTTP` | 最后一次 HTTP 尝试从发送到响应体读取完成的耗时 |
| `TTFT` | 从发起请求到收到首个输出 Token 的时间 |
| `TPOT` | 首个 Token 之后，每个输出 Token 的平均生成时间 |
| `Usage Coverage` | 成功请求中包含 Token Usage 数据的比例 |

延迟指标按有效样本计算 P50、P95 和 P99。Token 相关指标依赖上游接口返回 Usage；缺失 Usage 不会影响请求成功状态，但会降低覆盖率。

## 安全说明

- API Key 不会通过页面或 JSON API 回显，但会保存在 SQLite 数据库中，请妥善保护数据库和备份。
- 项目默认仅监听环回地址，且当前版本不包含用户登录与权限系统。
- 对公网部署时，应在服务前增加身份认证、TLS、CSRF 防护、访问控制和请求限流。
- 回答预览可能包含敏感内容，可在场景配置中关闭保存。

## 项目结构

项目采用单应用分层结构：

```text
api/                 Gin 路由、页面处理器和 v1 JSON API
cmd/server/          服务入口与优雅关闭
config/              环境配置加载、归一化和校验
database/            GORM、SQLite 和 SQL 日志适配
internal/benchmark/  并发调度与指标聚合
internal/protocol/   推理接口客户端与响应解析
internal/service/    业务流程编排
internal/dao/        数据访问
internal/model/      持久化模型
internal/dto/        请求与响应结构
migrations/          数据库迁移
pkg/                 日志和统一 API 响应
web/frontend/        Tailwind、Flowbite、HTMX 和 ECharts 前端源码
web/static/          原有样式与构建后的嵌入式静态资源
web/templates/       Go 服务端模板和 HTMX 局部区域
```

主要依赖方向为 `cmd -> api -> service -> dao -> database`。协议调用和并发调度由独立包负责，HTTP 处理器不直接访问数据库。

## JSON API

JSON API 使用统一响应结构：`code` 表示状态码，`message` 表示结果信息，成功响应通过 `data` 返回业务数据。

- `GET /health`
- `GET /api/v1/targets?page=1&page_size=20`
- `GET /api/v1/scenarios?page=1&page_size=20`
- `GET /api/v1/runs/:id`
- `POST /api/v1/runs/:id/cancel`

Web 页面负责目标、场景和运行管理；JSON API 可用于健康检查、状态查询和自动化集成。

## 开发约定

- HTTP 参数解析和响应写入保留在 `api/`，业务流程放在 `internal/service/`。
- 数据库操作集中在 `internal/dao/`，并沿调用链传递 `context.Context`。
- 协议适配放在 `internal/protocol/`，负载调度和指标计算放在 `internal/benchmark/`。
- 提交变更前运行 `gofmt`、`go test ./...` 和 `make build`。
- 新增配置时同步更新 `.env.example` 和本文档的配置表。

## License

本项目基于 [Apache License 2.0](LICENSE) 开源。
