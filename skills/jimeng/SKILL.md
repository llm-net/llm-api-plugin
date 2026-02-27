---
name: jimeng
description: Generate videos using Jimeng APIs - Action Imitation V2 (action reenactment from person image + template video) and OmniHuman 1.5 (talking-head from portrait + audio). Use when the user wants to do action imitation, motion transfer, or talking-head video generation.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Jimeng Video Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Models

| Model | Type | Required inputs | What it does |
|-------|------|-----------------|--------------|
| `jimeng-action-imitation-v2` (default) | image+video → video | Person image + template video | Transfers actions/motions from a template video onto a person image. No prompt needed. |
| `jimeng-omnihuman` | image+audio → video | Portrait image + audio + optional prompt | Generates a talking-head video where the portrait speaks the given audio. Supports 720p/1080p, fast mode. |

**How to choose**:
- **Motion transfer / dance reenactment**: Use `jimeng-action-imitation-v2` — provide a person photo and a template video with the desired actions.
- **Talking head / digital human speaking**: Use `jimeng-omnihuman` — provide a portrait and an audio file (< 60s).

**Local image files**: Use `--image-file <path>` instead of `--image <url>` to read files directly (avoids shell argument size limits).

## Usage

```bash
# Action Imitation (no prompt needed)
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli generate --model jimeng-action-imitation-v2 --image <url_or_file> --video <url>

# OmniHuman talking-head
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli generate "prompt text" --model jimeng-omnihuman --image <url_or_file> --audio <url>
```

## Configuration

```bash
# Volcano Engine access keys
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>
```

## Notes

- Asynchronous API: submits task, polls every 5s, max 300s
- Output format: MP4
