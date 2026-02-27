package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/llm-net/llm-api-plugin/cmd/jimeng-cli/provider"
	"github.com/llm-net/llm-api-plugin/internal/config"
)

func usage() {
	fmt.Fprintf(os.Stderr, `jimeng-cli - CLI for Jimeng Video Generation APIs (即梦视频生成)

Usage:
  %[1]s generate <prompt> [flags]                    Generate video from text/image prompt
  %[1]s models [<model-name>]                        List available models (JSON)
  %[1]s config set-keys <ACCESS_KEY_ID> <SECRET_KEY> Set Jimeng access keys
  %[1]s config show                                  Show current config

Flags for generate:
  --model <model>          Model name                                  [default: jimeng-action-imitation-v2]
  --output <path>          Output file path                            [default: output_<timestamp>.mp4]

Flags for jimeng-action-imitation-v2:
  --image <url>            Person image URL (required)
  --video <url>            Template video URL (required)
  --cut-first-second       Whether to cut the first second of result     [default: true]

Flags for jimeng-omnihuman:
  --image <url>            Portrait image URL (required)
  --audio <url>            Audio URL, under 60s (required)
  --resolution <num>       Output resolution: 720 or 1080               [default: 1080]
  --fast-mode              Enable fast mode (trades quality for speed)
  --seed <num>             Random seed (-1 for random)

Examples:
  %[1]s generate --model jimeng-action-imitation-v2 --image https://example.com/person.jpg --video https://example.com/dance.mp4
  %[1]s generate "Hello world" --model jimeng-omnihuman --image https://example.com/portrait.jpg --audio https://example.com/speech.wav
  %[1]s models
  %[1]s models jimeng-action-imitation-v2
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
	// Parse all flags
	var prompt string
	modelName := defaultModel
	seed := 0
	image := ""
	video := ""
	audio := ""
	resolution := 0
	fastMode := false
	cutFirstSecond := true
	cutFirstSecondSet := false
	output := ""

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			i++
			if i < len(args) {
				modelName = args[i]
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
		case "--video":
			i++
			if i < len(args) {
				video = args[i]
			}
		case "--audio":
			i++
			if i < len(args) {
				audio = args[i]
			}
		case "--resolution":
			i++
			if i < len(args) {
				if v, err := strconv.Atoi(args[i]); err == nil {
					resolution = v
				}
			}
		case "--fast-mode":
			fastMode = true
		case "--cut-first-second":
			i++
			if i < len(args) {
				if v, err := strconv.ParseBool(args[i]); err == nil {
					cutFirstSecond = v
					cutFirstSecondSet = true
				}
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

	// Validate model name
	providerKey, ok := modelProvider[modelName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown model %q\nRun `jimeng-cli models` to list available models.\n", modelName)
		os.Exit(1)
	}

	// Default output path
	if output == "" {
		output = fmt.Sprintf("output_%s.mp4", time.Now().Format("20060102_150405"))
	}

	// Resolve credentials (shared by all models)
	cfg, _ := config.LoadOrCreate()
	ak, sk := config.ResolveAccessKeys("JIMENG_ACCESS_KEY_ID", "JIMENG_SECRET_ACCESS_KEY", cfg.Jimeng)
	if ak == "" || sk == "" {
		fmt.Fprintf(os.Stderr, "Error: Jimeng access keys not set.\n"+
			"  Option 1: export JIMENG_ACCESS_KEY_ID=<AK> && export JIMENG_SECRET_ACCESS_KEY=<SK>\n"+
			"  Option 2: jimeng-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>\n")
		os.Exit(1)
	}

	// Dispatch to model-specific function
	switch providerKey {
	case "action-imitation-v2":
		generateWithActionImitationV2(ak, sk, image, video, cutFirstSecond, cutFirstSecondSet, output)
	case "omnihuman":
		generateWithOmniHuman(ak, sk, prompt, image, audio, resolution, fastMode, seed, output)
	}
}

// generateWithActionImitationV2 handles jimeng-action-imitation-v2 model.
func generateWithActionImitationV2(ak, sk, image, video string, cutFirstSecond, cutFirstSecondSet bool, output string) {
	if image == "" {
		fmt.Fprintln(os.Stderr, "Error: --image is required for jimeng-action-imitation-v2")
		os.Exit(1)
	}
	if video == "" {
		fmt.Fprintln(os.Stderr, "Error: --video is required for jimeng-action-imitation-v2")
		os.Exit(1)
	}

	p := provider.NewJimengActionImitationV2Provider(ak, sk)
	ctx := context.Background()

	req := &provider.ActionImitationV2Request{
		ImageURL: image,
		VideoURL: video,
	}
	if cutFirstSecondSet {
		req.CutFirstSecond = &cutFirstSecond
	}

	fmt.Fprintf(os.Stderr, "Submitting action imitation task...\n")

	submitResult, err := p.SubmitTask(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error submitting task: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Task created: %s\n", submitResult.TaskID)
	fmt.Fprintf(os.Stderr, "Polling for result (timeout %v)...\n", pollTimeout)

	deadline := time.Now().Add(pollTimeout)
	for {
		qr, err := p.QueryTask(ctx, submitResult.TaskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		switch qr.Status {
		case "done":
			if qr.VideoURL == "" {
				fmt.Fprintln(os.Stderr, "Error: task succeeded but no video URL in response")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Downloading video...\n")
			size, err := downloadVideo(qr.VideoURL, output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Video saved: %s (%d bytes)\n", output, size)
			return
		case "failed":
			fmt.Fprintf(os.Stderr, "Error: task failed: %s\n", qr.Message)
			os.Exit(1)
		}

		if time.Now().After(deadline) {
			fmt.Fprintf(os.Stderr, "Error: timeout after %v, task still in status: %s\n", pollTimeout, qr.Status)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "  Status: %s, waiting %v...\n", qr.Status, pollInterval)
		time.Sleep(pollInterval)
	}
}

// generateWithOmniHuman handles jimeng-omnihuman model.
func generateWithOmniHuman(ak, sk, prompt, image, audio string, resolution int, fastMode bool, seed int, output string) {
	if image == "" {
		fmt.Fprintln(os.Stderr, "Error: --image is required for jimeng-omnihuman")
		os.Exit(1)
	}
	if audio == "" {
		fmt.Fprintln(os.Stderr, "Error: --audio is required for jimeng-omnihuman")
		os.Exit(1)
	}

	p := provider.NewJimengOmniHumanProvider(ak, sk)
	ctx := context.Background()

	req := &provider.OmniHumanRequest{
		ImageURL:         image,
		AudioURL:         audio,
		Prompt:           prompt,
		Seed:             seed,
		OutputResolution: resolution,
		FastMode:         fastMode,
	}

	fmt.Fprintf(os.Stderr, "Submitting OmniHuman task...\n")

	submitResult, err := p.SubmitTask(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error submitting task: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Task created: %s\n", submitResult.TaskID)
	fmt.Fprintf(os.Stderr, "Polling for result (timeout %v)...\n", pollTimeout)

	deadline := time.Now().Add(pollTimeout)
	for {
		qr, err := p.QueryTask(ctx, submitResult.TaskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		switch qr.Status {
		case "done":
			if qr.VideoURL == "" {
				fmt.Fprintln(os.Stderr, "Error: task succeeded but no video URL in response")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Downloading video...\n")
			size, err := downloadVideo(qr.VideoURL, output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Video saved: %s (%d bytes)\n", output, size)
			return
		case "failed":
			fmt.Fprintf(os.Stderr, "Error: task failed: %s\n", qr.Message)
			os.Exit(1)
		}

		if time.Now().After(deadline) {
			fmt.Fprintf(os.Stderr, "Error: timeout after %v, task still in status: %s\n", pollTimeout, qr.Status)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "  Status: %s, waiting %v...\n", qr.Status, pollInterval)
		time.Sleep(pollInterval)
	}
}
