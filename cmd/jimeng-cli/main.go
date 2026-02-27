package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/llm-net/llm-api-plugin/internal/config"
)

func usage() {
	fmt.Fprintf(os.Stderr, `jimeng-cli - CLI for Jimeng Video Generation 3.0 Pro API (即梦视频生成)

Usage:
  %[1]s generate <prompt> [flags]                    Generate video from text prompt
  %[1]s models [<model-name>]                        List available models (JSON)
  %[1]s config set-keys <ACCESS_KEY_ID> <SECRET_KEY> Set Jimeng access keys
  %[1]s config show                                  Show current config

Flags for generate:
  --model <model>          Model name                                  [default: jimeng-video-gen-3-pro]
  --ratio <ratio>          Aspect ratio (16:9, 9:16, 1:1, 4:3, etc.)   [default: 16:9]
  --frames <num>           Total frames: 121 (5s) or 241 (10s)         [default: 121]
  --seed <num>             Random seed (-1 for random)
  --image <url>            First frame image URL (for image-to-video)
  --image-base64 <data>    First frame image base64 (for image-to-video)
  --output <path>          Output file path                            [default: output_<timestamp>.mp4]

Examples:
  %[1]s generate "A cat playing piano in a jazz bar"
  %[1]s generate "Ocean waves at sunset" --frames 241 --ratio 16:9
  %[1]s generate "Dancing robot" --ratio 9:16 --output robot.mp4
  %[1]s generate "Expand this image" --image https://example.com/photo.jpg
  %[1]s models
  %[1]s models jimeng-video-gen-3-pro
`, filepath.Base(os.Args[0]))
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "config":
		handleConfig()
	case "generate":
		handleGenerate()
	case "models":
		handleModels()
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: config set-keys <ACCESS_KEY_ID> <SECRET_KEY> | config show")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "set-keys":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Usage: config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>")
			os.Exit(1)
		}
		cfg, _ := config.LoadOrCreate()
		cfg.Jimeng = &config.ServiceConfig{
			AccessKeyID:    os.Args[3],
			SecretAccessKey: os.Args[4],
		}
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Jimeng access keys saved to %s\n", config.Path())
	case "show":
		cfg, _ := config.LoadOrCreate()
		ak, sk := config.ResolveAccessKeys("JIMENG_ACCESS_KEY_ID", "JIMENG_SECRET_ACCESS_KEY", cfg.Jimeng)
		if ak == "" && sk == "" {
			fmt.Println("Jimeng: not configured")
			return
		}
		akSource := "config file"
		if os.Getenv("JIMENG_ACCESS_KEY_ID") != "" {
			akSource = "env JIMENG_ACCESS_KEY_ID"
		}
		skSource := "config file"
		if os.Getenv("JIMENG_SECRET_ACCESS_KEY") != "" {
			skSource = "env JIMENG_SECRET_ACCESS_KEY"
		}
		fmt.Printf("Config: %s\n", config.Path())
		fmt.Printf("AccessKeyID: %s (source: %s)\n", maskSecret(ak), akSource)
		fmt.Printf("SecretAccessKey: %s (source: %s)\n", maskSecret(sk), skSource)
	default:
		fmt.Fprintf(os.Stderr, "Unknown config command: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func handleModels() {
	data, err := registry.JSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if len(os.Args) >= 3 {
		m := registry.FindModel(os.Args[2])
		if m == nil {
			fmt.Fprintf(os.Stderr, "Unknown model: %s\n", os.Args[2])
			os.Exit(1)
		}
		single, _ := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(single))
		return
	}
	fmt.Println(string(data))
}

func handleGenerate() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: generate <prompt> [flags]")
		os.Exit(1)
	}

	var prompt string
	ratio := ""
	frames := 0
	seed := 0
	image := ""
	imageBase64 := ""
	output := ""

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			i++ // model flag accepted but currently only one model
		case "--ratio":
			i++
			if i < len(args) {
				ratio = args[i]
			}
		case "--frames":
			i++
			if i < len(args) {
				if v, err := strconv.Atoi(args[i]); err == nil {
					frames = v
				}
			}
		case "--seed":
			i++
			if i < len(args) {
				if v, err := strconv.Atoi(args[i]); err == nil {
					seed = v
				}
			}
		case "--image":
			i++
			if i < len(args) {
				image = args[i]
			}
		case "--image-base64":
			i++
			if i < len(args) {
				imageBase64 = args[i]
			}
		case "--output":
			i++
			if i < len(args) {
				output = args[i]
			}
		default:
			if prompt == "" {
				prompt = args[i]
			} else {
				prompt += " " + args[i]
			}
		}
	}

	if prompt == "" {
		fmt.Fprintln(os.Stderr, "Error: prompt is required")
		os.Exit(1)
	}

	cfg, _ := config.LoadOrCreate()
	ak, sk := config.ResolveAccessKeys("JIMENG_ACCESS_KEY_ID", "JIMENG_SECRET_ACCESS_KEY", cfg.Jimeng)
	if ak == "" || sk == "" {
		fmt.Fprintf(os.Stderr, "Error: Jimeng access keys not set.\n"+
			"  Option 1: export JIMENG_ACCESS_KEY_ID=<AK> && export JIMENG_SECRET_ACCESS_KEY=<SK>\n"+
			"  Option 2: jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>\n")
		os.Exit(1)
	}

	p := newProvider(ak, sk)

	fmt.Fprintf(os.Stderr, "Submitting video generation task...\n")

	taskID, err := p.submitTask(prompt, image, imageBase64, ratio, frames, seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error submitting task: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Task created: %s\n", taskID)
	fmt.Fprintf(os.Stderr, "Polling for result (timeout %v)...\n", pollTimeout)

	result, err := p.waitForTask(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result.VideoURL == "" {
		fmt.Fprintln(os.Stderr, "Error: task succeeded but no video URL in response")
		os.Exit(1)
	}

	if output == "" {
		output = fmt.Sprintf("output_%s.mp4", time.Now().Format("20060102_150405"))
	}

	fmt.Fprintf(os.Stderr, "Downloading video...\n")
	size, err := downloadVideo(result.VideoURL, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Video saved: %s (%d bytes)\n", output, size)
}
