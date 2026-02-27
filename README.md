# llm-api-plugin

把各家 LLM 服务商的慢速 API（图片生成、视频生成、数字人等）封装成命令行工具，再包装为 Claude Code Skills，让其他 Claude Code 项目可以直接调用，不需要自己对接 API。

## 安装

在 Claude Code 中执行：

```bash
# 1. 添加插件市场
/plugin marketplace add llm-net/llm-api-plugin

# 2. 安装插件（在 /plugin 界面的 Discover 标签中选择，或直接执行）
/plugin install llm-api-plugin@llm-net-llm-api-plugin
```

首次启动会话时，`SessionStart` hook 会自动运行 `scripts/setup.sh`，检测你的操作系统和 CPU 架构，从 GitHub Release 下载对应的预编译二进制到插件的 `bin/` 目录。**不需要 Go 环境。**

## 当前支持的工具

| CLI | Skill 名称 | 能力 |
|-----|-----------|------|
| gemini-cli | `/llm-api-plugin:gemini-image` | Gemini 图片生成 |
| ark-cli | `/llm-api-plugin:ark` | Seedance / 即梦视频生成 |
| jimeng-cli | `/llm-api-plugin:jimeng` | 即梦动作模仿 / OmniHuman 数字人 |
| topview-cli | `/llm-api-plugin:topview` | TopView 数字人口播视频 |

## 配置

两种方式，**环境变量优先于配置文件**。

### 环境变量

```bash
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

### 配置文件

```bash
gemini-cli config set-key <KEY>                          # Gemini
ark-cli config set-key <KEY>                             # Ark
ark-cli config set-keys <ACCESS_KEY_ID> <SECRET>         # Ark 中的 Jimeng 模型
jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET>      # Jimeng
topview-cli config set-key <KEY>                         # TopView
topview-cli config set-uid <UID>                         # TopView UID
```

配置存储在 `~/.config/llm-api-plugin/config.json`，所有 CLI 共享。用 `<cli> config show` 查看当前配置和来源。

## 使用

安装配置完成后，在任意 Claude Code 项目中直接调用 skill：

```
/llm-api-plugin:gemini-image    # Gemini 图片生成
/llm-api-plugin:ark             # Seedance / 即梦视频生成
/llm-api-plugin:jimeng          # 即梦动作模仿 / OmniHuman 数字人
/llm-api-plugin:topview         # TopView 数字人口播视频
```

Agent 会自动运行 `<cli> models` 获取可用模型和参数，然后构造正确的命令执行。

## 升级

```bash
# 刷新市场
/plugin marketplace update llm-net-llm-api-plugin

# 更新插件
/plugin update llm-api-plugin@llm-net-llm-api-plugin
```

下次启动会话时，hook 会自动检查并下载新版本二进制。

## 开发

### 环境要求

- Go 1.24.0+

### 编译

```bash
make build            # 编译所有 CLI 到 bin/
make gemini-cli       # 只编译单个
```

### 项目结构

```
cmd/xxx-cli/          各 CLI 的 main 包
internal/config/      统一配置管理（环境变量 + 配置文件）
internal/httpclient/  公共 HTTP client（120s 超时）
internal/models/      模型自描述结构（models 子命令的数据类型）
skills/xxx/SKILL.md   Claude Code Skill 定义
hooks/hooks.json      SessionStart hook（自动下载二进制）
scripts/setup.sh      二进制下载脚本
scripts/version       当前版本号
```

### 添加新的 CLI 工具

1. 创建 `cmd/xxx-cli/main.go` — 参考 `cmd/gemini-cli/` 的结构
2. 在 `models.go` 中注册模型和参数 — 让 `xxx-cli models` 能输出 JSON
3. 用 `config.ResolveAPIKey("XXX_API_KEY", cfg.Xxx)` 读取 API key
4. 创建 `skills/xxx/SKILL.md` — 告诉 agent 怎么调用
5. 在 `Makefile` 的 `TOOLS` 列表和 `scripts/setup.sh` 的 `TOOLS` 数组中添加 `xxx-cli`

### 发布

```bash
# 1. 更新版本号
#    - scripts/version
#    - .claude-plugin/plugin.json
#    - .claude-plugin/marketplace.json

# 2. 提交并打 tag
git tag v0.2.0
git push origin main --tags
```

GitHub Actions 自动交叉编译并上传到 Release。
