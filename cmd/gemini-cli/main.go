package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/llm-net/llm-api-plugin/internal/config"
)

func usage() {
	fmt.Fprintf(os.Stderr, `gemini-cli - CLI for Google Gemini API

Usage:
  %[1]s generate <prompt> [flags]    Generate image from text prompt
  %[1]s models [<model-name>]        List available models (JSON)
  %[1]s config set-key <API_KEY>     Set Gemini API key
  %[1]s config show                  Show current config

Flags for generate:
  --model <model>    Model name                            [default: gemini-3-pro-image-preview]
  --ratio <ratio>    Aspect ratio (e.g. 16:9, 1:1, 4:3)   [default: 1:1]
  --size <size>      Image size: 1K, 2K, 4K                [default: 2K]
  --output <path>    Output file path                      [default: output_<timestamp>.png]
  --text-only        Only return text, no image

Examples:
  %[1]s generate "A cat riding a bicycle in watercolor style"
  %[1]s generate "Infographic about climate change" --ratio 16:9 --size 4K
  %[1]s generate "Explain quantum computing" --text-only
  %[1]s models
  %[1]s models gemini-3-pro-image-preview
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
		cfg.Gemini = &config.ServiceConfig{APIKey: os.Args[3]}
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Gemini API key saved to %s\n", config.Path())
	case "show":
		cfg, _ := config.LoadOrCreate()
		apiKey := config.ResolveAPIKey("GEMINI_API_KEY", cfg.Gemini)
		if apiKey == "" {
			fmt.Println("Gemini: not configured")
			return
		}
		source := "config file"
		if os.Getenv("GEMINI_API_KEY") != "" {
			source = "env GEMINI_API_KEY"
		}
		masked := apiKey
		if len(masked) > 8 {
			masked = masked[:4] + "..." + masked[len(masked)-4:]
		}
		fmt.Printf("Config: %s\nGemini API Key: %s (source: %s)\n", config.Path(), masked, source)
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
	ratio := "1:1"
	size := "2K"
	output := ""
	textOnly := false

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			i++
			if i < len(args) {
				model = args[i]
			}
		case "--ratio":
			i++
			if i < len(args) {
				ratio = args[i]
			}
		case "--size":
			i++
			if i < len(args) {
				size = args[i]
			}
		case "--output":
			i++
			if i < len(args) {
				output = args[i]
			}
		case "--text-only":
			textOnly = true
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
	apiKey := config.ResolveAPIKey("GEMINI_API_KEY", cfg.Gemini)
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: Gemini API key not set.\n  Option 1: export GEMINI_API_KEY=<KEY>\n  Option 2: gemini-cli config set-key <KEY>\n")
		os.Exit(1)
	}

	if textOnly {
		ratio = ""
		size = ""
	}

	modelName := model
	if modelName == "" {
		modelName = defaultModel
	}
	fmt.Fprintf(os.Stderr, "Generating with model %s...\n", modelName)

	resp, err := generateContent(apiKey, model, prompt, ratio, size)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Candidates) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no candidates in response")
		os.Exit(1)
	}

	var texts []string
	var imageCount int

	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			texts = append(texts, part.Text)
		}
		if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
			imageCount++
			imgData, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding image: %v\n", err)
				continue
			}

			outPath := output
			if outPath == "" {
				ext := "png"
				if strings.Contains(part.InlineData.MIMEType, "jpeg") {
					ext = "jpg"
				}
				outPath = fmt.Sprintf("output_%s_%d.%s", time.Now().Format("20060102_150405"), imageCount, ext)
			}

			if err := os.WriteFile(outPath, imgData, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving image: %v\n", err)
				continue
			}
			fmt.Fprintf(os.Stderr, "Image saved: %s (%d bytes)\n", outPath, len(imgData))
		}
	}

	if len(texts) > 0 {
		fmt.Println(strings.Join(texts, "\n"))
	}
}
