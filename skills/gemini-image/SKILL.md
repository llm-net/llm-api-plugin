---
name: gemini-image
description: Generate images using Google Gemini API. Use when the user wants to generate, create, or draw images with Gemini.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Gemini Image Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Models

| Model | Best for | Speed | Aspect ratios |
|-------|----------|-------|---------------|
| `gemini-3-pro-image-preview` (default) | High-quality image generation | Slower, higher quality | 1:1, 16:9, 9:16, 4:3, 3:4 |
| `gemini-3.1-flash-image-preview` | Fast generation, more ratios | Faster, cost-efficient | 1:1, 16:9, 9:16, 4:3, 3:4, 2:3, 3:2, 4:5, 5:4, 1:4, 4:1, 1:8, 8:1, 21:9 |

Both models support `--size 1K/2K/4K` (default 2K) and `--ratio` options. Output format: PNG.

**How to choose**: Use `gemini-3-pro-image-preview` for best quality. Use `gemini-3.1-flash-image-preview` when you need speed, lower cost, or uncommon aspect ratios (e.g. ultra-wide 21:9, vertical 1:4).

## Usage

```bash
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli generate "<prompt>" [--model <model>] [--ratio <ratio>] [--size <size>] [--output path.png]
```

## Configuration

```bash
${CLAUDE_PLUGIN_ROOT}/bin/gemini-cli config set-key <GEMINI_API_KEY>
```

## Notes

- Synchronous API, may take 10-30 seconds
- Output format: PNG
