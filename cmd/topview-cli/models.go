package main

import "github.com/llm-net/llm-api-plugin/internal/models"

var registry = &models.Registry{
	Tool: "topview-cli",
	Models: []models.Model{
		{
			Name:         "topview-video-avatar",
			Description:  "Generate video avatar using TopView AI. Upload a portrait image and audio to create a talking avatar video.",
			Capabilities: []string{"image-audio-to-video", "video-avatar"},
			Params: map[string]models.Param{
				"image": {
					Description: "Path to portrait image file (jpg, png, webp)",
					Type:        "string",
					Required:    true,
				},
				"audio": {
					Description: "Path to audio file (mp3, wav, m4a, aac)",
					Type:        "string",
					Required:    true,
				},
			},
		},
	},
}
