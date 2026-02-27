# llm-api-plugin

Claude Code Plugin，封装多个 LLM 服务商的 API 为独立 CLI 工具，作为 skills 供其他 Claude Code 项目使用。

## 项目结构

```
llm-api-plugin/
├── .claude-plugin/
│   └── plugin.json                  # 插件清单（name、version、description）
├── scripts/
│   ├── setup.sh                     # 安装/升级时自动执行，从 GitHub Release 下载二进制
│   └── version                      # 当前需要的二进制版本号（如 v1.0.0）
├── skills/                          # Claude Code Skills（SKILL.md）
│   ├── gemini-image/
│   │   └── SKILL.md                 # /llm-api-plugin:gemini-image
│   ├── ark/
│   │   └── SKILL.md                 # /llm-api-plugin:ark
│   ├── jimeng/
│   │   └── SKILL.md                 # /llm-api-plugin:jimeng
│   └── topview/
│       └── SKILL.md                 # /llm-api-plugin:topview
├── cmd/                             # 各 CLI 源码入口（每个子目录一个 main 包）
│   ├── gemini-cli/
│   │   └── main.go
│   ├── ark-cli/
│   │   └── main.go
│   ├── jimeng-cli/
│   │   └── main.go
│   └── topview-cli/
│       └── main.go
├── internal/                        # 公共内部库（所有 CLI 共享）
│   ├── config/                      # 统一配置管理（~/.config/llm-api-plugin/config.json）
│   │   └── config.go
│   ├── httpclient/                  # 公共 HTTP client 封装
│   │   └── client.go
│   └── models/                      # 模型自描述结构（models 子命令的数据类型）
│       └── models.go
├── bin/                             # 下载的二进制存放目录（gitignore）
│   └── .version                     # 本地已安装的版本号
├── .github/
│   └── workflows/
│       └── release.yml              # tag 触发 → 交叉编译 → 上传 GitHub Release
├── go.mod
├── Makefile                         # 本地开发：make build / make gemini-cli
└── CLAUDE.md
```

## 架构决策

### 二进制分发（不依赖用户编译环境）
- 用户安装 plugin 时不需要 Go 环境
- GitHub Actions 交叉编译 linux/darwin/windows × amd64/arm64
- `scripts/setup.sh` 检测平台，从 Release 下载对应二进制到 `bin/`

### 升级机制
- `scripts/version` 文件记录期望的二进制版本
- `bin/.version` 记录本地已安装的版本
- 用户执行 `/plugin update llm-api-plugin` 时：
  1. git pull 拉到新的 `scripts/version`
  2. 触发 `setup.sh`
  3. 对比版本，不一致则下载新二进制

### 配置管理
- 优先级：**环境变量 > 配置文件**
- 环境变量：
  - `GEMINI_API_KEY` — Gemini
  - `ARK_API_KEY` — 火山方舟 Ark
  - `JIMENG_ACCESS_KEY_ID` + `JIMENG_SECRET_ACCESS_KEY` — 即梦（火山引擎凭证）
  - `TOPVIEW_API_KEY` + `TOPVIEW_UID` — TopView
- 配置文件：`~/.config/llm-api-plugin/config.json`
- 解析入口：`config.ResolveAPIKey(envVar, fromConfig)` / `config.ResolveAccessKeys(akEnv, skEnv, fromConfig)` — 所有 CLI 统一调用

### API 调用
- 图片生成（gemini-cli）为同步 REST 调用
- 视频生成（ark-cli、jimeng-cli、topview-cli）为异步任务：提交 → 轮询 → 下载
- 公共 HTTP client 统一超时 120s（用于同步请求）
- 异步任务轮询超时：ark/jimeng 300s，topview 600s
- 认证方式按服务商不同：Gemini 用 `x-goog-api-key` header，Ark 用 Bearer token，Jimeng 用火山引擎 SDK，TopView 用 Bearer token + UID header

## 开发约定

- Go 版本：1.24.0
- 一个 `go.mod`，整个仓库作为一个 Go module
- `cmd/xxx-cli/` 每个 CLI 独立的 main 包
- `internal/` 放共享代码，不对外暴露
- 二进制命名：`gemini-cli`、`ark-cli`、`jimeng-cli`、`topview-cli`
- Skill 命名：`/llm-api-plugin:gemini-image`、`/llm-api-plugin:ark`、`/llm-api-plugin:jimeng`、`/llm-api-plugin:topview`

## 当前支持的模型

### gemini-cli
- `gemini-3-pro-image-preview` — 图片生成（同步，返回 base64 JPEG）
- `gemini-3.1-flash-image-preview` — 更快更便宜的图片生成，支持更多比例

### ark-cli
- `doubao-seedance-1-5-pro-251215` — Seedance 1.5 Pro 视频生成（文生视频 / 图生视频）
- `jimeng-video-gen-3-pro` — 即梦 3.0 Pro 文生视频 / 图生视频（首帧）
- `jimeng-i2v-3-pro` — 即梦 3.0 Pro 图生视频（首帧）
- `jimeng-i2v-startend-3-pro` — 即梦 3.0 Pro 图生视频（首尾帧）

### jimeng-cli
- `jimeng-action-imitation-v2` — 动作模仿 2.0（模板视频 + 人物图片 → 视频）
- `jimeng-omnihuman` — OmniHuman 1.5（肖像 + 音频 → 数字人视频）

### topview-cli
- `topview-video-avatar` — TopView 数字人（肖像 + 音频 → 口播视频）
