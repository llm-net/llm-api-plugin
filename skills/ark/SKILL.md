---
name: ark
description: Generate videos using Volcano Ark Seedance API or Jimeng Video Gen 3.0 Pro. Use when the user wants to generate or create videos with Seedance/Ark/Jimeng.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# Volcano Ark Video Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/ark-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Models

| Model | Type | Input | Key features |
|-------|------|-------|--------------|
| `doubao-seedance-1-5-pro-251215` (default) | text-to-video | Text prompt | 720p/1080p, 5s/10s duration, auto audio generation, best overall quality |
| `jimeng-t2v-3-pro` | text-to-video | Text prompt | 5s/10s (via frames 121/241), multiple aspect ratios |
| `jimeng-i2v-3-pro` | image-to-video | Text + first frame image | Animates a single image into video, use `--image <url>` or `--image-file <path>` |
| `jimeng-i2v-startend-3-pro` | image-to-video | Text + first & last frame images | Generates video transitioning between two images, use `--image`/`--end-image` or `--image-file`/`--end-image-file` |

**How to choose**:
- **Text only → video**: Use `doubao-seedance-1-5-pro-251215` for best quality with audio; use `jimeng-t2v-3-pro` for alternative style.
- **Image → video**: Use `jimeng-i2v-3-pro` to animate one image; use `jimeng-i2v-startend-3-pro` to morph between two images.
- **Local image files**: Use `--image-file <path>` / `--end-image-file <path>` instead of URL to read files directly (avoids shell argument size limits).

## Usage

```bash
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli generate "<prompt>" [--model <model>] [flags] [--output path.mp4]
```

## Configuration

Ark and Jimeng use different authentication:

```bash
# Ark (API key) — for doubao-seedance model
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli config set-key <ARK_API_KEY>

# Jimeng models (Volcano Engine access keys) — for jimeng-* models
${CLAUDE_PLUGIN_ROOT}/bin/ark-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>
```

## Notes

- Asynchronous API: submits task, polls every 5s, max 300s
- Output format: MP4
