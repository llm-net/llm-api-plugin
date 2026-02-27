package main

import "github.com/llm-net/llm-api-plugin/internal/models"

// modelProvider maps model name to its backend provider.
var modelProvider = map[string]string{
	"doubao-seedance-1-5-pro-251215": "ark",
	"jimeng-t2v-3-pro":               "jimeng",
	"jimeng-i2v-3-pro":               "jimeng",
	"jimeng-i2v-startend-3-pro":      "jimeng",
}

// jimengReqKey maps jimeng model name to its API req_key.
var jimengReqKey = map[string]string{
	"jimeng-t2v-3-pro":          "jimeng_t2v_v30_pro",
	"jimeng-i2v-3-pro":          "jimeng_ti2v_v30_pro",
	"jimeng-i2v-startend-3-pro": "jimeng_ti2v_v30_pro",
}

// common jimeng params shared across all 3 jimeng models
var jimengCommonParams = map[string]models.Param{
	"ratio": {
		Description: "Aspect ratio of the generated video",
		Type:        "string",
		Options:     []string{"16:9", "9:16", "1:1", "4:3", "3:4", "21:9"},
		Default:     "16:9",
	},
	"frames": {
		Description: "Total frames: 121 for 5 seconds, 241 for 10 seconds",
		Type:        "string",
		Options:     []string{"121", "241"},
		Default:     "121",
	},
	"seed": {
		Description: "Random seed (-1 for random)",
		Type:        "integer",
		Default:     "-1",
	},
}

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
		{
			Name:         "jimeng-t2v-3-pro",
			Description:  "即梦视频生成 3.0 Pro - 文生视频 (text-to-video)",
			Capabilities: []string{"text-to-video"},
			Params:       jimengCommonParams,
		},
		{
			Name:         "jimeng-i2v-3-pro",
			Description:  "即梦视频生成 3.0 Pro - 图生视频（首帧模式）(image-to-video, first frame)",
			Capabilities: []string{"image-to-video"},
			Params: mergeParams(jimengCommonParams, map[string]models.Param{
				"image": {
					Description: "First frame image URL",
					Type:        "string",
				},
				"image-file": {
					Description: "First frame image from local file (auto base64-encoded)",
					Type:        "string",
				},
			}),
		},
		{
			Name:         "jimeng-i2v-startend-3-pro",
			Description:  "即梦视频生成 3.0 Pro - 图生视频（首尾帧模式）(image-to-video, start+end frames)",
			Capabilities: []string{"image-to-video"},
			Params: mergeParams(jimengCommonParams, map[string]models.Param{
				"image": {
					Description: "First frame image URL",
					Type:        "string",
				},
				"image-file": {
					Description: "First frame image from local file (auto base64-encoded)",
					Type:        "string",
				},
				"end-image": {
					Description: "Last frame image URL",
					Type:        "string",
				},
				"end-image-file": {
					Description: "Last frame image from local file (auto base64-encoded)",
					Type:        "string",
				},
			}),
		},
	},
}

// mergeParams returns a new map combining base and extra params.
func mergeParams(base, extra map[string]models.Param) map[string]models.Param {
	m := make(map[string]models.Param, len(base)+len(extra))
	for k, v := range base {
		m[k] = v
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}
