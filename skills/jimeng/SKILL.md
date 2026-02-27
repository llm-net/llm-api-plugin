---
name: jimeng
description: Generate videos using Jimeng Video Generation 3.0 Pro API (即梦视频生成). Use when the user wants to generate or create videos with Jimeng.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Jimeng Video Generation 3.0 Pro (即梦视频生成)

Use the `jimeng-cli` binary to generate videos via the Volcano Engine Jimeng API.

## Binary Location

The binary is located at `<plugin-dir>/bin/jimeng-cli`. If it doesn't exist, run `<plugin-dir>/scripts/setup.sh` first.

## Discover Available Models

**Before generating, always run `models` first** to discover available models and their parameters:

```bash
# List all available models (JSON)
<plugin-dir>/bin/jimeng-cli models

# Get details for a specific model
<plugin-dir>/bin/jimeng-cli models jimeng-video-gen-3-pro
```

The output is JSON, example:
```json
{
  "tool": "jimeng-cli",
  "models": [
    {
      "name": "jimeng-video-gen-3-pro",
      "description": "Jimeng Video Generation 3.0 Pro - text-to-video and image-to-video",
      "capabilities": ["text-to-video", "image-to-video"],
      "params": {
        "ratio": {
          "description": "Aspect ratio of the generated video",
          "type": "string",
          "options": ["16:9", "9:16", "1:1", "4:3", "3:4", "21:9"],
          "default": "16:9"
        },
        "frames": {
          "description": "Total frames: 121 for 5 seconds, 241 for 10 seconds",
          "type": "string",
          "options": ["121", "241"],
          "default": "121"
        },
        "seed": {
          "description": "Random seed (-1 for random)",
          "type": "integer",
          "default": "-1"
        },
        "image": {
          "description": "First frame image URL for image-to-video mode",
          "type": "string"
        },
        "image-base64": {
          "description": "First frame image base64 for image-to-video mode",
          "type": "string"
        }
      }
    }
  ]
}
```

Use the `params` from the JSON output to construct the correct flags for `generate`.

## Usage

```bash
# Text-to-video generation
<plugin-dir>/bin/jimeng-cli generate "<prompt>" [--ratio 16:9] [--frames 121] [--seed 42] [--output path.mp4]

# Image-to-video generation (first frame mode)
<plugin-dir>/bin/jimeng-cli generate "<prompt>" --image <image-url> [--ratio 16:9] [--frames 121] [--output path.mp4]

# Examples
<plugin-dir>/bin/jimeng-cli generate "A cat playing piano in a jazz bar"
<plugin-dir>/bin/jimeng-cli generate "Ocean waves at sunset" --frames 241 --ratio 16:9
<plugin-dir>/bin/jimeng-cli generate "Dancing robot" --ratio 9:16 --output robot.mp4
<plugin-dir>/bin/jimeng-cli generate "Expand this image" --image https://example.com/photo.jpg
```

## Configuration

Jimeng uses AccessKeyID + SecretAccessKey authentication (Volcano Engine credentials).

```bash
# Set credentials via CLI
<plugin-dir>/bin/jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>

# Or via environment variables
export JIMENG_ACCESS_KEY_ID=<your-access-key-id>
export JIMENG_SECRET_ACCESS_KEY=<your-secret-access-key>
```

Config is stored at `~/.config/llm-api-plugin/config.json`.

## Notes

- The API is **asynchronous**: a task is created first, then the CLI polls for the result every 5 seconds
- Maximum wait time is 300 seconds (5 minutes)
- Progress status is printed to stderr during polling
- Videos are saved as `.mp4` files to the specified output path
- The task ID is printed to stderr for reference
- Supports both text-to-video and image-to-video (first frame) modes
- Default video duration is 5 seconds (121 frames), use `--frames 241` for 10 seconds
