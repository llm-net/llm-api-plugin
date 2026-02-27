---
name: jimeng
description: Generate videos using Jimeng APIs - Action Imitation V2 (action reenactment) and OmniHuman 1.5 (talking-head from portrait+audio). Use when the user wants to generate or create videos with Jimeng.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Jimeng Video Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Workflow

1. Run `${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli models` to discover available models and their parameters (JSON output)
2. Use the params from JSON to construct the correct flags for `generate`
3. Run `${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli generate [prompt] --model <model> [flags] [--output path.mp4]`

Note: some models don't need a prompt (e.g. action-imitation only needs --image and --video).

## Configuration

```bash
# Volcano Engine access keys
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>
```

## Notes

- Asynchronous API: submits task, polls every 5s, max 300s
- Output format: MP4
