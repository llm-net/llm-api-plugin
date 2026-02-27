package main

import "github.com/llm-net/llm-api-plugin/internal/models"

const defaultModel = "jimeng-action-imitation-v2"

// modelProvider maps model name to a dispatch key used in handleGenerate.
var modelProvider = map[string]string{
	"jimeng-action-imitation-v2": "action-imitation-v2",
	"jimeng-omnihuman":           "omnihuman",
}

var registry = &models.Registry{
	Tool: "jimeng-cli",
	Models: []models.Model{
		{
			Name:         "jimeng-action-imitation-v2",
			Description:  "Jimeng Action Imitation 2.0 - generate video by imitating actions from a template video onto a person image (即梦动作模仿2.0)",
			Capabilities: []string{"image+video-to-video"},
			Params: map[string]models.Param{
				"image": {
					Description: "Person image URL",
					Type:        "string",
				},
				"image-file": {
					Description: "Person image from local file (auto base64-encoded)",
					Type:        "string",
				},
				"video": {
					Description: "Template video URL with actions to imitate (required)",
					Type:        "string",
					Required:    true,
				},
				"cut-first-second": {
					Description: "Whether to cut the first second of result video",
					Type:        "boolean",
					Default:     "true",
				},
			},
		},
		{
			Name:         "jimeng-omnihuman",
			Description:  "Jimeng OmniHuman 1.5 - generate talking-head video from a portrait image and audio (即梦OmniHuman1.5)",
			Capabilities: []string{"image+audio-to-video"},
			Params: map[string]models.Param{
				"image": {
					Description: "Portrait image URL",
					Type:        "string",
				},
				"image-file": {
					Description: "Portrait image from local file (auto base64-encoded)",
					Type:        "string",
				},
				"audio": {
					Description: "Audio URL, must be under 60 seconds (required)",
					Type:        "string",
					Required:    true,
				},
				"resolution": {
					Description: "Output video resolution",
					Type:        "string",
					Options:     []string{"720", "1080"},
					Default:     "1080",
				},
				"fast-mode": {
					Description: "Enable fast mode (trades quality for speed)",
					Type:        "boolean",
					Default:     "false",
				},
				"seed": {
					Description: "Random seed (-1 for random)",
					Type:        "integer",
					Default:     "-1",
				},
			},
		},
	},
}
