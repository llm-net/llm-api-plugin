---
name: topview
description: Generate video avatar using TopView AI. Upload a portrait image and audio file to create a talking-head avatar video with lip-sync. Use when the user wants a digital human / avatar speaking from portrait + audio.
allowed-tools: Bash, Read, Write
user-invocable: true
---

# TopView Video Avatar Generation

Binary: `${CLAUDE_PLUGIN_ROOT}/bin/topview-cli`

If it doesn't exist, run `bash ${CLAUDE_PLUGIN_ROOT}/scripts/setup.sh` first.

## Models

| Model | Type | Required inputs | What it does |
|-------|------|-----------------|--------------|
| `topview-video-avatar` | image+audio → video | Local portrait image + local audio file | Creates a talking avatar video with lip-sync from a portrait photo and audio. Files are uploaded to TopView then processed. |

**When to use**: Best for creating professional-looking talking-head videos with natural lip-sync. Takes local files directly (no URL needed). Supports longer processing time (up to 600s) for higher quality output.

**vs jimeng-omnihuman**: Both create talking-head videos. TopView takes local files directly; jimeng-omnihuman requires URLs. Choose based on available API keys and whether inputs are local files or URLs.

## Usage

```bash
${CLAUDE_PLUGIN_ROOT}/bin/topview-cli generate --image <local_image_path> --audio <local_audio_path> [--output path.mp4]
```

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
