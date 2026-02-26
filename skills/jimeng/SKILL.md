---
name: jimeng
description: Generate videos using Volcano Ark Seedance API (火山方舟). Use when the user wants to generate or create videos with Jimeng/Seedance.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Jimeng Video Generation (via Volcano Ark Seedance API)

Use the `ark-cli` binary to generate videos via the Volcano Ark API.

## Binary Location

The binary is located at `<plugin-dir>/bin/ark-cli`. If it doesn't exist, run `<plugin-dir>/scripts/setup.sh` first.

## Discover Available Models

**Before generating, always run `models` first** to discover available models and their parameters:

```bash
# List all available models (JSON)
<plugin-dir>/bin/ark-cli models

# Get details for a specific model
<plugin-dir>/bin/ark-cli models doubao-seedance-1-5-pro-251215
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
        "duration": {
          "description": "Video duration in seconds",
          "type": "string",
          "options": ["5", "10"],
          "default": "5"
        },
        "resolution": {
          "description": "Video resolution",
          "type": "string",
          "options": ["720p", "1080p"],
          "default": "720p"
        },
        "ratio": {
          "description": "Aspect ratio of the generated video",
          "type": "string",
          "options": ["16:9", "9:16", "1:1", "4:3", "3:4", "21:9"],
          "default": "16:9"
        },
        "audio": {
          "description": "Whether to generate audio",
          "type": "string",
          "options": ["true", "false"],
          "default": "true"
        }
      }
    }
  ]
}
```

Use the `params` from the JSON output to construct the correct flags for `generate`.

## Usage

```bash
# Generate a video (use flags from models output)
<plugin-dir>/bin/ark-cli generate "<prompt>" --model <model-name> [--duration 5] [--resolution 720p] [--ratio 16:9] [--no-audio] [--output path.mp4]

# Examples
<plugin-dir>/bin/ark-cli generate "A cat playing piano in a jazz bar"
<plugin-dir>/bin/ark-cli generate "Ocean waves at sunset" --duration 10 --resolution 1080p
<plugin-dir>/bin/ark-cli generate "Dancing robot" --ratio 9:16 --no-audio --output robot.mp4
```

## Configuration

If the API key is not configured, run:
```bash
<plugin-dir>/bin/ark-cli config set-key <ARK_API_KEY>
```

Config is stored at `~/.config/llm-api-plugin/config.json`.

## Notes

- The API is **asynchronous**: a task is created first, then the CLI polls for the result every 5 seconds
- Maximum wait time is 300 seconds (5 minutes)
- Progress status is printed to stderr during polling
- Videos are saved as `.mp4` files to the specified output path
- The task ID is printed to stderr for reference
