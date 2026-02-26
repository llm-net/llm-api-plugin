package main

import "github.com/llm-net/llm-api-plugin/internal/models"

var registry = &models.Registry{
	Tool: "ark-cli",
	Models: []models.Model{
		{
			Name:         "doubao-seedance-1-5-pro-251215",
			Description:  "Video generation from text or image prompts using Seedance 1.5 Pro",
			Capabilities: []string{"text-to-video", "image-to-video"},
			Params: map[string]models.Param{
				"duration": {
					Description: "Video duration in seconds",
					Type:        "string",
					Options:     []string{"5", "10"},
					Default:     "5",
				},
				"resolution": {
					Description: "Video resolution",
					Type:        "string",
					Options:     []string{"720p", "1080p"},
					Default:     "720p",
				},
				"ratio": {
					Description: "Aspect ratio of the generated video",
					Type:        "string",
					Options:     []string{"16:9", "9:16", "1:1", "4:3", "3:4", "21:9"},
					Default:     "16:9",
				},
				"audio": {
					Description: "Whether to generate audio",
					Type:        "string",
					Options:     []string{"true", "false"},
					Default:     "true",
				},
			},
		},
	},
}
