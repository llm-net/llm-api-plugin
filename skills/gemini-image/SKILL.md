---
name: gemini-image
description: Generate images using Google Gemini API. Use when the user wants to generate, create, or draw images with Gemini.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Gemini Image Generation

Use the `gemini-cli` binary to generate images via the Gemini API.

## Binary Location

The binary is located at `${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli`. If it doesn't exist, run `${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Discover Available Models

**Before generating, always run `models` first** to discover available models and their parameters:

```bash
# List all available models (JSON)
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli models

# Get details for a specific model
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli models gemini-3.1-flash-image-preview
```

The output is JSON, example:
```json
{
  "tool": "gemini-cli",
  "models": [
    {
      "name": "gemini-3-pro-image-preview",
      "description": "Image generation and editing from text prompts, returns both text and image",
      "capabilities": ["text-to-image", "text"],
      "params": {
        "ratio": {
          "description": "Aspect ratio of the generated image",
          "type": "string",
          "options": ["1:1", "16:9", "9:16", "4:3", "3:4"],
          "default": "1:1"
        },
        "size": {
          "description": "Image resolution",
          "type": "string",
          "options": ["1K", "2K", "4K"],
          "default": "2K"
        }
      }
    }
  ]
}
```

Use the `params` from the JSON output to construct the correct flags for `generate`.

## Usage

```bash
# Generate an image (use flags from models output)
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli generate "<prompt>" --model <model-name> [--ratio 16:9] [--size 2K] [--output path.png]

# Text-only response (no image)
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli generate "<prompt>" --text-only
```

## Configuration

If the API key is not configured, run:
```bash
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli config set-key <GEMINI_API_KEY>
```

Config is stored at `~/.config/llm-api-plugin/config.json`.

## Notes

- The API is synchronous, may take 10-30 seconds for image generation
- Images are returned as PNG, saved to the specified output path
- Both text and image can be returned in a single response
- `gemini-3.1-flash-image-preview` is faster and cheaper, supports more aspect ratios
