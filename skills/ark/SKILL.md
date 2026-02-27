---
name: ark
description: Generate videos using Volcano Ark Seedance API or Jimeng Video Gen 3.0 Pro. Use when the user wants to generate or create videos with Seedance/Ark/Jimeng.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Volcano Ark Video Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/ark-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Workflow

1. Run `${CLAUDE_PLUGIN_ROOT}/bin/ark-cli models` to discover available models and their parameters (JSON output)
2. Use the params from JSON to construct the correct flags for `generate`
3. Run `${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "<prompt>" --model <model> [flags] [--output path.mp4]`

## Configuration

Ark and Jimeng use different authentication:

```bash
# Ark (API key)
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli config set-key <ARK_API_KEY>

# Jimeng models (Volcano Engine access keys)
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>
```

## Notes

- Asynchronous API: submits task, polls every 5s, max 300s
- Output format: MP4
