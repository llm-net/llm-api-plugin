---
name: topview
description: Generate video avatar using TopView AI. Upload a portrait image and audio file to create a talking avatar video with digital human.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# TopView Video Avatar Generation

Use the `topview-cli` binary to generate talking avatar videos via the TopView AI API.

## Binary Location

The binary is located at `${CLAUDE_PLUGIN_ROOT}/bin/topview-cli`. If it doesn't exist, run `${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Discover Available Models

**Before generating, always run `models` first** to discover available models and their parameters:

```bash
# List all available models (JSON)
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli models

# Get details for a specific model
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli models topview-video-avatar
```

The output is JSON, example:
```json
{
  "tool": "topview-cli",
  "models": [
    {
      "name": "topview-video-avatar",
      "description": "Generate video avatar using TopView AI. Upload a portrait image and audio to create a talking avatar video.",
      "capabilities": ["image-audio-to-video", "video-avatar"],
      "params": {
        "image": {
          "description": "Path to portrait image file (jpg, png, webp)",
          "type": "string",
          "required": true
        },
        "audio": {
          "description": "Path to audio file (mp3, wav, m4a, aac)",
          "type": "string",
          "required": true
        }
      }
    }
  ]
}
```

Use the `params` from the JSON output to construct the correct flags for `generate`.

## Usage

```bash
# Generate a video avatar
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli generate --image <portrait.jpg> --audio <speech.mp3> [--output path.mp4]

# Examples
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli generate --image portrait.jpg --audio speech.mp3
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli generate --image photo.png --audio audio.wav --output avatar.mp4
```

## Configuration

If the API key is not configured, run:
```bash
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli config set-key <TOPVIEW_API_KEY>
```

Optionally set the TopView UID:
```bash
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli config set-uid <TOPVIEW_UID>
```

Config is stored at `~/.config/llm-api-plugin/config.json`.

## Notes

- The API is **asynchronous**: files are uploaded first, then a task is created, and the CLI polls for the result every 5 seconds
- Maximum wait time is 600 seconds (10 minutes)
- Progress status is printed to stderr during polling
- Videos are saved as `.mp4` files to the specified output path
- Supported image formats: jpg, png, webp
- Supported audio formats: mp3, wav, m4a, aac
- The task ID is printed to stderr for reference
