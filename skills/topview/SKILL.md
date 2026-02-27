---
name: topview
description: Generate video avatar using TopView AI. Upload a portrait image and audio file to create a talking avatar video with digital human.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# TopView Video Avatar Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/topview-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Workflow

1. Run `${CLAUDE_PLUGIN_ROOT}/bin/topview-cli models` to discover available models and their parameters (JSON output)
2. Use the params from JSON to construct the correct flags for `generate`
3. Run `${CLAUDE_PLUGIN_ROOT}/bin/topview-cli generate --image <path> --audio <path> [--output path.mp4]`

## Configuration

```bash
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli config set-key <TOPVIEW_API_KEY>
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli config set-uid <TOPVIEW_UID>
```

## Notes

- Asynchronous API: uploads files, submits task, polls every 5s, max 600s
- Supported images: jpg, png, webp
- Supported audio: mp3, wav, m4a, aac
- Output format: MP4
