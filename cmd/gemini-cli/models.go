package main

import "github.com/llm-net/llm-api-plugin/internal/models"

var registry = &models.Registry{
	Tool: "gemini-cli",
	Models: []models.Model{
		{
			Name:         "gemini-3-pro-image-preview",
			Description:  "Image generation and editing from text prompts, returns both text and image",
			Capabilities: []string{"text-to-image", "text"},
			Params: map[string]models.Param{
				"ratio": {
					Description: "Aspect ratio of the generated image",
					Type:        "string",
					Options:     []string{"1:1", "16:9", "9:16", "4:3", "3:4"},
					Default:     "1:1",
				},
				"size": {
					Description: "Image resolution",
					Type:        "string",
					Options:     []string{"1K", "2K", "4K"},
					Default:     "2K",
				},
			},
		},
		{
			Name:         "gemini-3.1-flash-image-preview",
			Description:  "Fast and cost-efficient image generation, supports more aspect ratios and image search grounding",
			Capabilities: []string{"text-to-image", "text"},
			Params: map[string]models.Param{
				"ratio": {
					Description: "Aspect ratio of the generated image",
					Type:        "string",
					Options:     []string{"1:1", "16:9", "9:16", "4:3", "3:4", "2:3", "3:2", "4:5", "5:4", "1:4", "4:1", "1:8", "8:1", "21:9"},
					Default:     "1:1",
				},
				"size": {
					Description: "Image resolution",
					Type:        "string",
					Options:     []string{"1K", "2K", "4K"},
					Default:     "2K",
				},
			},
		},
	},
}
