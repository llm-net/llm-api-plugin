# llm-api-plugin

把各家 LLM 服务商的慢速 API（图片生成、视频生成、数字人等）封装成命令行工具，再包装为 Claude Code Skills，让其他 Claude Code 项目可以直接调用，不需要自己对接 API。

## 安装

在 Claude Code 中执行：

```
/install-plugin https://github.com/llm-net/llm-api-plugin
```

安装过程会自动运行 `scripts/setup.sh`，检测你的操作系统和 CPU 架构，从 GitHub Release 下载对应的预编译二进制到插件的 `bin/` 目录。**不需要 Go 环境。**

## 配置 API Key

两种方式，**环境变量优先**：

### 方式一：环境变量（推荐）

适合 CI/CD、容器、或需要多项目隔离的场景。

```bash
# 加到 ~/.bashrc 或 ~/.zshrc

# Gemini
export GEMINI_API_KEY="AIza..."

# 火山方舟 Ark（Seedance 模型）
export ARK_API_KEY="..."

# 即梦 Jimeng（火山引擎 AccessKey）
export JIMENG_ACCESS_KEY_ID="AKL..."
export JIMENG_SECRET_ACCESS_KEY="..."

# TopView
export TOPVIEW_API_KEY="..."
export TOPVIEW_UID="..."
```

### 方式二：配置文件

适合个人开发，配一次就不用管了。

```bash
# Gemini
gemini-cli config set-key AIza...

# Ark
ark-cli config set-key ...

# Jimeng（AccessKey 对）
jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>
# ark-cli 中的 Jimeng 模型也用这组凭证
ark-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>

# TopView
topview-cli config set-key ...
topview-cli config set-uid ...
```

配置存储在 `~/.config/llm-api-plugin/config.json`，所有 CLI 工具共享同一份：

```json
{
  "gemini":  { "api_key": "AIza..." },
  "ark":     { "api_key": "..." },
  "jimeng":  { "access_key_id": "AKL...", "secret_access_key": "..." },
  "topview": { "api_key": "...", "uid": "..." }
}
```

查看当前配置（会显示 key 来源是环境变量还是配置文件）：

```bash
gemini-cli config show
# Gemini API Key: AIza...c3_s (source: env GEMINI_API_KEY)
```

> **优先级**：环境变量 > 配置文件。如果两处都设了，环境变量生效。

## 使用

安装配置完成后，在任意 Claude Code 项目中直接调用 skill：

```
/llm-api-plugin:gemini-image    # Gemini 图片生成
/llm-api-plugin:ark             # Seedance / 即梦视频生成
/llm-api-plugin:jimeng          # 即梦动作模仿 / OmniHuman 数字人
/llm-api-plugin:topview         # TopView 数字人口播视频
```

Agent 会自动读取 SKILL.md 中的说明，调用对应的 CLI 工具完成任务。

### 手动使用 CLI

每个 CLI 都支持 `models` 命令，查看可用的模型和参数：

```bash
# 查看所有可用模型
gemini-cli models

# 查看某个模型的详细参数
gemini-cli models gemini-3-pro-image-preview

# 生成图片
gemini-cli generate "A cat riding a bicycle" --ratio 16:9 --size 2K --output cat.png

# 生成视频
ark-cli generate "Ocean waves at sunset" --duration 10 --resolution 1080p --output ocean.mp4
```

## 当前支持的工具

| CLI | Skill 名称 | 能力 | 状态 |
|-----|-----------|------|------|
| gemini-cli | `/llm-api-plugin:gemini-image` | Gemini 图片生成 | 可用 |
| ark-cli | `/llm-api-plugin:ark` | Seedance / 即梦视频生成 | 可用 |
| jimeng-cli | `/llm-api-plugin:jimeng` | 即梦动作模仿 / OmniHuman | 可用 |
| topview-cli | `/llm-api-plugin:topview` | TopView 数字人口播 | 可用 |

## 升级

```
/plugin update llm-api-plugin
```

会 git pull 最新代码，如果 `scripts/version` 中的版本号有变化，自动下载新的二进制。

## 开发

### 环境要求

- Go 1.24.0+

### 编译

```bash
make build            # 编译所有 CLI 到 bin/
make gemini-cli       # 只编译 gemini-cli
make ark-cli          # 只编译 ark-cli
```

### 项目结构

```
cmd/xxx-cli/          各 CLI 的 main 包
internal/config/      统一配置管理（环境变量 + 配置文件）
internal/httpclient/  公共 HTTP client（120s 超时）
internal/models/      模型自描述结构（models 子命令的数据类型）
skills/xxx/SKILL.md   Claude Code Skill 定义
scripts/setup.sh      用户安装时自动下载二进制
scripts/version       当前版本号
```

### 添加新的 CLI 工具

1. 创建 `cmd/xxx-cli/main.go` — 参考 `cmd/gemini-cli/` 的结构
2. 在 `models.go` 中注册模型和参数 — 让 `xxx-cli models` 能输出 JSON
3. 用 `config.ResolveAPIKey("XXX_API_KEY", cfg.Xxx)` 读取 API key
4. 创建 `skills/xxx/SKILL.md` — 告诉 agent 怎么调用
5. 在 `Makefile` 的 `TOOLS` 列表和 `scripts/setup.sh` 的 `TOOLS` 数组中添加 `xxx-cli`

### 发布

打 tag 即可，GitHub Actions 自动交叉编译并上传到 Release：

```bash
# 更新 scripts/version 为新版本号
git tag v0.2.0
git push --tags
```
