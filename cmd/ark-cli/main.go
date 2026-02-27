package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/llm-net/llm-api-plugin/internal/config"
)

func usage() {
	fmt.Fprintf(os.Stderr, `ark-cli - CLI for Volcano Ark (火山方舟) Video Generation API

Usage:
  %[1]s generate <prompt> [flags]                    Generate video from text prompt
  %[1]s models [<model-name>]                        List available models (JSON)
  %[1]s config set-key <API_KEY>                     Set Ark API key
  %[1]s config set-keys <ACCESS_KEY_ID> <SECRET_KEY> Set Jimeng access keys
  %[1]s config show                                  Show current config

Flags for generate:
  --model <model>              Model name                                  [default: doubao-seedance-1-5-pro-251215]
  --duration <seconds>         Video duration: 5 or 10                     [default: 5]      (Ark models)
  --resolution <res>           Resolution: 720p or 1080p                   [default: 720p]   (Ark models)
  --ratio <ratio>              Aspect ratio (16:9, 9:16, 1:1, etc.)        [default: 16:9]
  --no-audio                   Disable audio generation                                      (Ark models)
  --frames <num>               Total frames: 121 (5s) or 241 (10s)         [default: 121]    (Jimeng models)
  --seed <num>                 Random seed (-1 for random)                                   (Jimeng models)
  --image <url>                First frame image URL                                         (Jimeng i2v models)
  --image-file <path>          First frame image from local file (auto base64-encoded)        (Jimeng i2v models)
  --end-image <url>            Last frame image URL                                          (Jimeng i2v-startend)
  --end-image-file <path>      Last frame image from local file (auto base64-encoded)         (Jimeng i2v-startend)
  --output <path>              Output file path                            [default: output_<timestamp>.mp4]

Examples:
  %[1]s generate "A cat playing piano in a jazz bar"
  %[1]s generate "Ocean waves at sunset" --duration 10 --resolution 1080p
  %[1]s generate "Dancing robot" --ratio 9:16 --no-audio
  %[1]s generate "A dreamy forest" --model jimeng-t2v-3-pro
  %[1]s generate "Expand this image" --model jimeng-i2v-3-pro --image https://example.com/photo.jpg
  %[1]s generate "Morph between" --model jimeng-i2v-startend-3-pro --image https://a.jpg --end-image https://b.jpg
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
		fmt.Fprintln(os.Stderr, "Usage: config set-key <KEY> | config set-keys <AK> <SK> | config show")
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

		// Ark config
		apiKey := config.ResolveAPIKey("ARK_API_KEY", cfg.Ark)
		fmt.Printf("Config: %s\n\n", config.Path())
		if apiKey == "" {
			fmt.Println("Ark: not configured")
		} else {
			source := "config file"
			if os.Getenv("ARK_API_KEY") != "" {
				source = "env ARK_API_KEY"
			}
			masked := apiKey
			if len(masked) > 8 {
				masked = masked[:4] + "..." + masked[len(masked)-4:]
			}
			fmt.Printf("Ark API Key: %s (source: %s)\n", masked, source)
		}

		fmt.Println()

		// Jimeng config
		ak, sk := config.ResolveAccessKeys("JIMENG_ACCESS_KEY_ID", "JIMENG_SECRET_ACCESS_KEY", cfg.Jimeng)
		if ak == "" && sk == "" {
			fmt.Println("Jimeng: not configured")
		} else {
			akSource := "config file"
			if os.Getenv("JIMENG_ACCESS_KEY_ID") != "" {
				akSource = "env JIMENG_ACCESS_KEY_ID"
			}
			skSource := "config file"
			if os.Getenv("JIMENG_SECRET_ACCESS_KEY") != "" {
				skSource = "env JIMENG_SECRET_ACCESS_KEY"
			}
			fmt.Printf("Jimeng AccessKeyID: %s (source: %s)\n", maskSecret(ak), akSource)
			fmt.Printf("Jimeng SecretAccessKey: %s (source: %s)\n", maskSecret(sk), skSource)
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
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: generate <prompt> [flags]")
		os.Exit(1)
	}

	var prompt string
	model := ""
	// Ark flags
	duration := "5"
	resolution := "720p"
	ratio := "16:9"
	audio := "true"
	// Jimeng flags
	frames := 0
	seed := 0
	image := ""
	imageFile := ""
	endImage := ""
	endImageFile := ""
	// Common
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
		case "--image-file":
			i++
			if i < len(args) {
				imageFile = args[i]
			}
		case "--end-image":
			i++
			if i < len(args) {
				endImage = args[i]
			}
		case "--end-image-file":
			i++
			if i < len(args) {
				endImageFile = args[i]
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

	if output == "" {
		output = fmt.Sprintf("output_%s.mp4", time.Now().Format("20060102_150405"))
	}

	// Read image files and base64-encode them
	var imageBase64, endImageBase64 string
	if imageFile != "" {
		data, err := os.ReadFile(imageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading image file %s: %v\n", imageFile, err)
			os.Exit(1)
		}
		imageBase64 = base64.StdEncoding.EncodeToString(data)
	}
	if endImageFile != "" {
		data, err := os.ReadFile(endImageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading end image file %s: %v\n", endImageFile, err)
			os.Exit(1)
		}
		endImageBase64 = base64.StdEncoding.EncodeToString(data)
	}

	// Determine model and provider
	modelName := model
	if modelName == "" {
		modelName = defaultModel
	}

	provider, ok := modelProvider[modelName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown model %q. Run 'ark-cli models' to see available models.\n", modelName)
		os.Exit(1)
	}

	switch provider {
	case "ark":
		generateWithArk(modelName, prompt, resolution, duration, ratio, audio, output)
	case "jimeng":
		generateWithJimeng(modelName, prompt, ratio, frames, seed, image, imageBase64, endImage, endImageBase64, output)
	}
}

func generateWithArk(model, prompt, resolution, duration, ratio, audio, output string) {
	cfg, _ := config.LoadOrCreate()
	apiKey := config.ResolveAPIKey("ARK_API_KEY", cfg.Ark)
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: Ark API key not set.\n  Option 1: export ARK_API_KEY=<KEY>\n  Option 2: ark-cli config set-key <KEY>\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Creating task with model %s...\n", model)

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

	if result.Content == nil || result.Content.VideoURL == "" {
		fmt.Fprintln(os.Stderr, "Error: task succeeded but no video URL in response")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Downloading video...\n")
	size, err := downloadVideo(result.Content.VideoURL, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Video saved: %s (%d bytes)\n", output, size)
}

func generateWithJimeng(model, prompt, ratio string, frames, seed int, image, imageBase64, endImage, endImageBase64, output string) {
	cfg, _ := config.LoadOrCreate()
	ak, sk := config.ResolveAccessKeys("JIMENG_ACCESS_KEY_ID", "JIMENG_SECRET_ACCESS_KEY", cfg.Jimeng)
	if ak == "" || sk == "" {
		fmt.Fprintf(os.Stderr, "Error: Jimeng access keys not set.\n"+
			"  Option 1: export JIMENG_ACCESS_KEY_ID=<AK> && export JIMENG_SECRET_ACCESS_KEY=<SK>\n"+
			"  Option 2: ark-cli config set-keys <ACCESS_KEY_ID> <SECRET_ACCESS_KEY>\n")
		os.Exit(1)
	}

	reqKey := jimengReqKey[model]

	p := newJimengProvider(ak, sk)

	fmt.Fprintf(os.Stderr, "Submitting video generation task (%s)...\n", model)

	taskID, err := p.submitTask(jimengSubmitOpts{
		ReqKey:           reqKey,
		Prompt:           prompt,
		FirstFrameImage:  image,
		FirstFrameBase64: imageBase64,
		EndFrameImage:    endImage,
		EndFrameBase64:   endImageBase64,
		AspectRatio:      ratio,
		Frames:           frames,
		Seed:             seed,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error submitting task: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Task created: %s\n", taskID)
	fmt.Fprintf(os.Stderr, "Polling for result (timeout %v)...\n", pollTimeout)

	result, err := p.jimengWaitForTask(reqKey, taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result.VideoURL == "" {
		fmt.Fprintln(os.Stderr, "Error: task succeeded but no video URL in response")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Downloading video...\n")
	size, err := downloadVideo(result.VideoURL, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Video saved: %s (%d bytes)\n", output, size)
}
