---
name: gemini-image
description: Generate images using Google Gemini API. Use when the user wants to generate, create, or draw images with Gemini.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Gemini Image Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Workflow

1. Run `${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli models` to discover available models and their parameters (JSON output)
2. Use the params from JSON to construct the correct flags for `generate`
3. Run `${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli generate "<prompt>" --model <model> [flags] [--output path.png]`

## Configuration

```bash
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli config set-key <GEMINI_API_KEY>
```

## Notes

- Synchronous API, may take 10-30 seconds
- Output format: PNG
