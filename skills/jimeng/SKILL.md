---
name: jimeng
description: Generate videos using Jimeng APIs - Action Imitation V2 (action reenactment) and OmniHuman 1.5 (talking-head from portrait+audio). Use when the user wants to generate or create videos with Jimeng.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Jimeng Video Generation (即梦视频生成)

Use the `jimeng-cli` binary to generate videos via the Volcano Engine Jimeng API.

## Binary Location

The binary is located at `${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli`. If it doesn't exist, run `${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Discover Available Models

**Before generating, always run `models` first** to discover available models and their parameters:

```bash
# List all available models (JSON)
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli models

# Get details for a specific model
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli models jimeng-action-imitation-v2
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli models jimeng-omnihuman
```

The output is JSON, example:
```json
{
  "tool": "jimeng-cli",
  "models": [
    {
      "name": "jimeng-action-imitation-v2",
      "description": "Jimeng Action Imitation 2.0 - generate video by imitating actions from a template video onto a person image",
      "capabilities": ["image+video-to-video"],
      "params": {
        "image": { "description": "Person image URL (required)", "type": "string", "required": true },
        "video": { "description": "Template video URL with actions to imitate (required)", "type": "string", "required": true },
        "cut-first-second": { "description": "Whether to cut the first second of result video", "type": "boolean", "default": "true" }
      }
    },
    {
      "name": "jimeng-omnihuman",
      "description": "Jimeng OmniHuman 1.5 - generate talking-head video from a portrait image and audio",
      "capabilities": ["image+audio-to-video"],
      "params": {
        "image": { "description": "Portrait image URL (required)", "type": "string", "required": true },
        "audio": { "description": "Audio URL, must be under 60 seconds (required)", "type": "string", "required": true },
        "resolution": { "description": "Output video resolution", "type": "string", "options": ["720","1080"], "default": "1080" },
        "fast-mode": { "description": "Enable fast mode (trades quality for speed)", "type": "boolean", "default": "false" },
        "seed": { "description": "Random seed (-1 for random)", "type": "integer", "default": "-1" }
      }
    }
  ]
}
```

Use the `params` from the JSON output to construct the correct flags for `generate`.

## Usage

### jimeng-action-imitation-v2

Generate video by transferring actions from a template video onto a person image. No prompt needed.

```bash
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli generate --model jimeng-action-imitation-v2 \
  --image <person-image-url> \
  --video <template-video-url> \
  [--cut-first-second true] \
  [--output path.mp4]

# Example
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli generate --model jimeng-action-imitation-v2 \
  --image https://example.com/person.jpg \
  --video https://example.com/dance.mp4 \
  --output dance_result.mp4
```

### jimeng-omnihuman

Generate talking-head video from a portrait image and audio file (OmniHuman 1.5).

```bash
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli generate "<optional-prompt>" --model jimeng-omnihuman \
  --image <portrait-url> \
  --audio <audio-url> \
  [--resolution 1080] \
  [--fast-mode] \
  [--seed 42] \
  [--output path.mp4]

# Example
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli generate "Hello world" --model jimeng-omnihuman \
  --image https://example.com/portrait.jpg \
  --audio https://example.com/speech.wav \
  --resolution 1080 \
  --output talking.mp4
```

## Configuration

Jimeng uses AccessKeyID + SecretAccessKey authentication (Volcano Engine credentials).

```bash
# Set credentials via CLI
${CLAUDE_PLUGIN_ROOT}/bin/jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>

# Or via environment variables
export JIMENG_ACCESS_KEY_ID=<your-access-key-id>
export JIMENG_SECRET_ACCESS_KEY=<your-secret-access-key>
```

Config is stored at `~/.config/llm-api-plugin/config.json`.

## Notes

- All APIs are **asynchronous**: a task is created first, then the CLI polls for the result every 5 seconds
- Maximum wait time is 300 seconds (5 minutes)
- Progress status is printed to stderr during polling
- Videos are saved as `.mp4` files to the specified output path
- The task ID is printed to stderr for reference
- **jimeng-action-imitation-v2**: requires `--image` (person) and `--video` (template), no prompt needed
- **jimeng-omnihuman**: requires `--image` (portrait) and `--audio` (under 60s), prompt is optional
