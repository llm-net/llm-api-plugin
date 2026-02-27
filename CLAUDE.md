# llm-api-plugin

Claude Code Plugin，封装多个 LLM 服务商的 API 为独立 CLI 工具，作为 skills 供其他 Claude Code 项目使用。

## 开发约定

- Go 1.24.0，单 `go.mod`
- `cmd/xxx-cli/` 各 CLI 的 main 包，`internal/` 共享代码
- 每个 CLI 支持 `models` 子命令输出 JSON 自描述（模型名、参数、类型、默认值）
- 配置优先级：环境变量 > `~/.config/llm-api-plugin/config.json`
- 图片生成（gemini-cli）同步，视频生成（ark/jimeng/topview）异步轮询

## CLI 与 Skill 对应

| CLI | Skill | 认证方式 |
|-----|-------|---------|
| gemini-cli | `/llm-api-plugin:gemini-image` | `GEMINI_API_KEY` |
| ark-cli | `/llm-api-plugin:ark` | `ARK_API_KEY` + `JIMENG_ACCESS_KEY_ID`/`SECRET` |
| jimeng-cli | `/llm-api-plugin:jimeng` | `JIMENG_ACCESS_KEY_ID`/`SECRET_ACCESS_KEY` |
| topview-cli | `/llm-api-plugin:topview` | `TOPVIEW_API_KEY` + `TOPVIEW_UID` |

## 关键路径

```
cmd/xxx-cli/                CLI 源码
internal/config/config.go   配置管理（ResolveAPIKey / ResolveAccessKeys）
internal/httpclient/        公共 HTTP client（120s 超时）
internal/models/            模型自描述结构
skills/xxx/SKILL.md         Skill 定义（agent 读取）
hooks/hooks.json            SessionStart hook（自动下载二进制）
scripts/setup.sh            二进制下载脚本
scripts/version             当前版本号
.claude-plugin/plugin.json  插件清单
.claude-plugin/marketplace.json  插件市场清单
.github/workflows/release.yml   tag 触发交叉编译
```

## 发布

同步更新三处版本号 → 提交 → 打 tag → push：
- `scripts/version`
- `.claude-plugin/plugin.json`
- `.claude-plugin/marketplace.json`
