package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/llm-net/llm-api-plugin/internal/config"
)

func usage() {
	fmt.Fprintf(os.Stderr, `ark-cli - CLI for Volcano Ark (火山方舟) Video Generation API

Usage:
  %[1]s generate <prompt> [flags]    Generate video from text prompt
  %[1]s models [<model-name>]        List available models (JSON)
  %[1]s config set-key <API_KEY>     Set Ark API key
  %[1]s config show                  Show current config

Flags for generate:
  --model <model>          Model name                                  [default: doubao-seedance-1-5-pro-251215]
  --duration <seconds>     Video duration: 5 or 10                     [default: 5]
  --resolution <res>       Resolution: 720p or 1080p                   [default: 720p]
  --ratio <ratio>          Aspect ratio (16:9, 9:16, 1:1, etc.)        [default: 16:9]
  --no-audio               Disable audio generation
  --output <path>          Output file path                            [default: output_<timestamp>.mp4]

Examples:
  %[1]s generate "A cat playing piano in a jazz bar"
  %[1]s generate "Ocean waves at sunset" --duration 10 --resolution 1080p
  %[1]s generate "Dancing robot" --ratio 9:16 --no-audio
  %[1]s models
  %[1]s models doubao-seedance-1-5-pro-251215
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
		fmt.Fprintln(os.Stderr, "Usage: config set-key <KEY> | config show")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "set-key":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: config set-key <API_KEY>")
			os.Exit(1)
		}
		cfg, _ := config.LoadOrCreate()
		cfg.Ark = &config.ServiceConfig{APIKey: os.Args[3]}
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Ark API key saved to %s\n", config.Path())
	case "show":
		cfg, _ := config.LoadOrCreate()
		apiKey := config.ResolveAPIKey("ARK_API_KEY", cfg.Ark)
		if apiKey == "" {
			fmt.Println("Ark: not configured")
			return
		}
		source := "config file"
		if os.Getenv("ARK_API_KEY") != "" {
			source = "env ARK_API_KEY"
		}
		masked := apiKey
		if len(masked) > 8 {
			masked = masked[:4] + "..." + masked[len(masked)-4:]
		}
		fmt.Printf("Config: %s\nArk API Key: %s (source: %s)\n", config.Path(), masked, source)
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
	model := ""
	duration := "5"
	resolution := "720p"
	ratio := "16:9"
	audio := "true"
	output := ""

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			i++
			if i < len(args) {
				model = args[i]
			}
		case "--duration":
			i++
			if i < len(args) {
				duration = args[i]
			}
		case "--resolution":
			i++
			if i < len(args) {
				resolution = args[i]
			}
		case "--ratio":
			i++
			if i < len(args) {
				ratio = args[i]
			}
		case "--no-audio":
			audio = "false"
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
	apiKey := config.ResolveAPIKey("ARK_API_KEY", cfg.Ark)
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: Ark API key not set.\n  Option 1: export ARK_API_KEY=<KEY>\n  Option 2: ark-cli config set-key <KEY>\n")
		os.Exit(1)
	}

	modelName := model
	if modelName == "" {
		modelName = defaultModel
	}
	fmt.Fprintf(os.Stderr, "Creating task with model %s...\n", modelName)

	taskID, err := createTask(apiKey, model, prompt, resolution, duration, ratio, audio)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating task: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Task created: %s\n", taskID)
	fmt.Fprintf(os.Stderr, "Polling for result (timeout %v)...\n", pollTimeout)

	result, err := waitForTask(apiKey, taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result.Output == nil || result.Output.VideoURL == "" {
		fmt.Fprintln(os.Stderr, "Error: task succeeded but no video URL in response")
		os.Exit(1)
	}

	if output == "" {
		output = fmt.Sprintf("output_%s.mp4", time.Now().Format("20060102_150405"))
	}

	fmt.Fprintf(os.Stderr, "Downloading video...\n")
	size, err := downloadVideo(result.Output.VideoURL, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Video saved: %s (%d bytes)\n", output, size)
}
