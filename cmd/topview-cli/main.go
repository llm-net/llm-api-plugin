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
	fmt.Fprintf(os.Stderr, `topview-cli - CLI for TopView AI Video Avatar Generation

Usage:
  %[1]s generate --image <path> --audio <path> [flags]   Generate video avatar
  %[1]s models [<model-name>]                             List available models (JSON)
  %[1]s config set-key <API_KEY>                          Set TopView API key
  %[1]s config set-uid <UID>                              Set TopView UID
  %[1]s config show                                       Show current config

Flags for generate:
  --image <path>         Path to portrait image file (required)
  --audio <path>         Path to audio file (required)
  --output <path>        Output file path                         [default: output_<timestamp>.mp4]

Examples:
  %[1]s generate --image portrait.jpg --audio speech.mp3
  %[1]s generate --image photo.png --audio audio.wav --output avatar.mp4
  %[1]s models
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
		fmt.Fprintln(os.Stderr, "Usage: config set-key <KEY> | config set-uid <UID> | config show")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "set-key":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: config set-key <API_KEY>")
			os.Exit(1)
		}
		cfg, _ := config.LoadOrCreate()
		if cfg.TopView == nil {
			cfg.TopView = &config.ServiceConfig{}
		}
		cfg.TopView.APIKey = os.Args[3]
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("TopView API key saved to %s\n", config.Path())
	case "set-uid":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: config set-uid <UID>")
			os.Exit(1)
		}
		cfg, _ := config.LoadOrCreate()
		if cfg.TopView == nil {
			cfg.TopView = &config.ServiceConfig{}
		}
		cfg.TopView.UID = os.Args[3]
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("TopView UID saved to %s\n", config.Path())
	case "show":
		cfg, _ := config.LoadOrCreate()
		apiKey := config.ResolveAPIKey("TOPVIEW_API_KEY", cfg.TopView)
		if apiKey == "" {
			fmt.Println("TopView: not configured")
			return
		}
		source := "config file"
		if os.Getenv("TOPVIEW_API_KEY") != "" {
			source = "env TOPVIEW_API_KEY"
		}
		masked := apiKey
		if len(masked) > 8 {
			masked = masked[:4] + "..." + masked[len(masked)-4:]
		}
		fmt.Printf("Config: %s\nTopView API Key: %s (source: %s)\n", config.Path(), masked, source)

		uid := ""
		if cfg.TopView != nil {
			uid = cfg.TopView.UID
		}
		if envUID := os.Getenv("TOPVIEW_UID"); envUID != "" {
			uid = envUID
		}
		if uid != "" {
			fmt.Printf("TopView UID: %s\n", uid)
		}
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
	var imagePath, audioPath, output string

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--image":
			i++
			if i < len(args) {
				imagePath = args[i]
			}
		case "--audio":
			i++
			if i < len(args) {
				audioPath = args[i]
			}
		case "--output":
			i++
			if i < len(args) {
				output = args[i]
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", args[i])
			os.Exit(1)
		}
	}

	if imagePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --image is required")
		os.Exit(1)
	}
	if audioPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --audio is required")
		os.Exit(1)
	}

	// Resolve API key and UID
	cfg, _ := config.LoadOrCreate()
	apiKey := config.ResolveAPIKey("TOPVIEW_API_KEY", cfg.TopView)
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: TopView API key not set.\n  Option 1: export TOPVIEW_API_KEY=<KEY>\n  Option 2: topview-cli config set-key <KEY>\n")
		os.Exit(1)
	}

	uid := ""
	if cfg.TopView != nil {
		uid = cfg.TopView.UID
	}
	if envUID := os.Getenv("TOPVIEW_UID"); envUID != "" {
		uid = envUID
	}

	// Read image file
	fmt.Fprintf(os.Stderr, "Reading image: %s\n", imagePath)
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading image: %v\n", err)
		os.Exit(1)
	}

	// Read audio file
	fmt.Fprintf(os.Stderr, "Reading audio: %s\n", audioPath)
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading audio: %v\n", err)
		os.Exit(1)
	}

	// Upload image
	fmt.Fprintf(os.Stderr, "Uploading image to TopView...\n")
	imageFormat := getImageFormat(imagePath)
	imageContentType := detectContentType(imagePath)
	imageFileID, err := uploadFile(apiKey, uid, imageData, imageFormat, imageContentType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading image: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Image uploaded: fileId=%s\n", imageFileID)

	// Upload audio
	fmt.Fprintf(os.Stderr, "Uploading audio to TopView...\n")
	audioFormat := getAudioFormat(audioPath)
	audioContentType := detectContentType(audioPath)
	audioFileID, err := uploadFile(apiKey, uid, audioData, audioFormat, audioContentType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading audio: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Audio uploaded: fileId=%s\n", audioFileID)

	// Submit task
	fmt.Fprintf(os.Stderr, "Submitting video avatar task...\n")
	task, err := submitVideoAvatarTask(apiKey, uid, imageFileID, audioFileID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error submitting task: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Task created: %s\n", task.TaskID)

	// Poll for result
	fmt.Fprintf(os.Stderr, "Polling for result (timeout %v)...\n", pollTimeout)
	result, err := waitForTask(apiKey, uid, task.TaskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Download video
	if output == "" {
		output = fmt.Sprintf("output_%s.mp4", time.Now().Format("20060102_150405"))
	}

	fmt.Fprintf(os.Stderr, "Downloading video...\n")
	size, err := downloadVideo(result.OutputVideoURL, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Video saved: %s (%d bytes)\n", output, size)
}
