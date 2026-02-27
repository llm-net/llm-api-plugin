---
name: ark
description: Generate videos using Volcano Ark Seedance API (火山方舟) or Jimeng Video Gen 3.0 Pro (即梦). Use when the user wants to generate or create videos with Seedance/Ark/Jimeng.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Volcano Ark Video Generation (火山方舟 Seedance + 即梦)

Use the `ark-cli` binary to generate videos via the Volcano Ark API or Jimeng API.

## Binary Location

The binary is located at `${CLAUDE_PLUGIN_ROOT}/bin/ark-cli`. If it doesn't exist, run `${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Discover Available Models

**Before generating, always run `models` first** to discover available models and their parameters:

```bash
# List all available models (JSON)
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli models

# Get details for a specific model
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli models doubao-seedance-1-5-pro-251215
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli models jimeng-video-gen-3-pro
```

The output is JSON, example:
```json
{
  "tool": "ark-cli",
  "models": [
    {
      "name": "doubao-seedance-1-5-pro-251215",
      "description": "Video generation from text or image prompts using Seedance 1.5 Pro",
      "capabilities": ["text-to-video", "image-to-video"],
      "params": {
        "duration": { "type": "string", "options": ["5", "10"], "default": "5" },
        "resolution": { "type": "string", "options": ["720p", "1080p"], "default": "720p" },
        "ratio": { "type": "string", "options": ["16:9", "9:16", "1:1", "4:3", "3:4", "21:9"], "default": "16:9" },
        "audio": { "type": "string", "options": ["true", "false"], "default": "true" }
      }
    },
    {
      "name": "jimeng-video-gen-3-pro",
      "description": "Jimeng Video Generation 3.0 Pro - text-to-video and image-to-video",
      "capabilities": ["text-to-video", "image-to-video"],
      "params": {
        "ratio": { "type": "string", "options": ["16:9", "9:16", "1:1", "4:3", "3:4", "21:9"], "default": "16:9" },
        "frames": { "type": "string", "options": ["121", "241"], "default": "121" },
        "seed": { "type": "integer", "default": "-1" },
        "image": { "type": "string", "description": "First frame image URL for image-to-video mode" },
        "image-base64": { "type": "string", "description": "First frame image base64 for image-to-video mode" }
      }
    }
  ]
}
```

Use the `params` from the JSON output to construct the correct flags for `generate`.

## Usage

```bash
# Generate with Ark Seedance (default model)
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "<prompt>" [--duration 5] [--resolution 720p] [--ratio 16:9] [--no-audio] [--output path.mp4]

# Generate with Jimeng Video Gen 3.0 Pro
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "<prompt>" --model jimeng-video-gen-3-pro [--ratio 16:9] [--frames 121] [--seed 42] [--output path.mp4]

# Image-to-video with Jimeng (first frame mode)
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "<prompt>" --model jimeng-video-gen-3-pro --image <image-url> [--output path.mp4]

# Examples
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "A cat playing piano in a jazz bar"
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "Ocean waves at sunset" --duration 10 --resolution 1080p
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "Dancing robot" --ratio 9:16 --no-audio --output robot.mp4
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "Expand this image" --model jimeng-video-gen-3-pro --image https://example.com/photo.jpg
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "A dreamy forest" --model jimeng-video-gen-3-pro --frames 241
```

## Configuration

Ark and Jimeng use different authentication methods:

```bash
# Set Ark API key
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli config set-key <ARK_API_KEY>

# Set Jimeng access keys (Volcano Engine credentials)
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>

# Or via environment variables
export ARK_API_KEY=<your-ark-api-key>
export JIMENG_ACCESS_KEY_ID=<your-access-key-id>
export JIMENG_SECRET_ACCESS_KEY=<your-secret-access-key>

# Show all configured credentials
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli config show
```

Config is stored at `~/.config/llm-api-plugin/config.json`.

## Notes

- The API is **asynchronous**: a task is created first, then the CLI polls for the result every 5 seconds
- Maximum wait time is 300 seconds (5 minutes)
- Progress status is printed to stderr during polling
- Videos are saved as `.mp4` files to the specified output path
- The task ID is printed to stderr for reference
- Jimeng model supports both text-to-video and image-to-video (first frame) modes
- Default video duration is 5 seconds for both models
