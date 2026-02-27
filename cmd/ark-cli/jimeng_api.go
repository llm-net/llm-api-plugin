package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/volcengine/volc-sdk-golang/service/visual"
)

// jimengProvider wraps the Volcano Engine visual client.
type jimengProvider struct {
	client *visual.Visual
}

// newJimengProvider creates a jimengProvider with the given access keys.
func newJimengProvider(accessKeyID, secretAccessKey string) *jimengProvider {
	client := visual.NewInstance()
	client.Client.SetAccessKey(accessKeyID)
	client.Client.SetSecretKey(secretAccessKey)
	return &jimengProvider{client: client}
}

// maskSecret masks a secret string for logging, showing only first 4 and last 4 chars.
func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// jimengSubmitOpts holds all parameters for a jimeng video generation task.
type jimengSubmitOpts struct {
	ReqKey              string
	Prompt              string
	FirstFrameImage     string
	FirstFrameBase64    string
	EndFrameImage       string
	EndFrameBase64      string
	AspectRatio         string
	Frames              int
	Seed                int
}

// submitTask submits a video generation task and returns the task ID.
func (p *jimengProvider) submitTask(opts jimengSubmitOpts) (string, error) {
	reqBody := map[string]interface{}{
		"req_key": opts.ReqKey,
		"prompt":  opts.Prompt,
	}

	// Collect image URLs and base64 data (first frame + optional end frame)
	var imageURLs []string
	var base64Data []string

	if opts.FirstFrameBase64 != "" {
		base64Data = append(base64Data, opts.FirstFrameBase64)
	} else if opts.FirstFrameImage != "" {
		imageURLs = append(imageURLs, opts.FirstFrameImage)
	}

	if opts.EndFrameBase64 != "" {
		base64Data = append(base64Data, opts.EndFrameBase64)
	} else if opts.EndFrameImage != "" {
		imageURLs = append(imageURLs, opts.EndFrameImage)
	}

	if len(base64Data) > 0 {
		reqBody["binary_data_base64"] = base64Data
	}
	if len(imageURLs) > 0 {
		reqBody["image_urls"] = imageURLs
	}

	if opts.Seed != 0 {
		reqBody["seed"] = opts.Seed
	}

	if opts.AspectRatio != "" {
		reqBody["aspect_ratio"] = opts.AspectRatio
	}

	if opts.Frames > 0 {
		reqBody["frames"] = opts.Frames
	} else {
		reqBody["frames"] = 121 // default 5 seconds
	}

	fmt.Fprintf(os.Stderr, "[jimeng] Submitting task: req_key=%s, prompt=%s\n", opts.ReqKey, opts.Prompt)

	resp, statusCode, err := p.client.CVSync2AsyncSubmitTask(reqBody)
	if err != nil {
		return "", fmt.Errorf("submit task: %w", err)
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("marshal response: %w", err)
	}

	if statusCode != 200 {
		return "", fmt.Errorf("HTTP %d: %s", statusCode, string(respBytes))
	}

	var result jimengSubmitResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBytes))
	}

	if result.Code != 10000 {
		return "", fmt.Errorf("API error: code=%d, message=%s", result.Code, result.Message)
	}

	if result.Data.TaskID == "" {
		return "", fmt.Errorf("no task ID in response: %s", string(respBytes))
	}

	fmt.Fprintf(os.Stderr, "[jimeng] Task submitted: %s\n", result.Data.TaskID)
	return result.Data.TaskID, nil
}

// jimengQueryTask queries the status of a video generation task.
func (p *jimengProvider) jimengQueryTask(reqKey, taskID string) (*jimengQueryResult, error) {
	reqBody := map[string]interface{}{
		"req_key": reqKey,
		"task_id": taskID,
	}

	resp, statusCode, err := p.client.CVSync2AsyncGetResult(reqBody)
	if err != nil {
		return nil, fmt.Errorf("query task: %w", err)
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(respBytes))
	}

	var result jimengQueryResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBytes))
	}

	if result.Code != 10000 {
		return &jimengQueryResult{
			TaskID:  taskID,
			Status:  "failed",
			Message: result.Message,
		}, nil
	}

	status := jimengMapTaskStatus(result.Data.Status)

	qr := &jimengQueryResult{
		TaskID:   taskID,
		Status:   status,
		VideoURL: result.Data.VideoURL,
		Message:  result.Message,
	}

	return qr, nil
}

// jimengWaitForTask polls until the task completes or times out.
func (p *jimengProvider) jimengWaitForTask(reqKey, taskID string) (*jimengQueryResult, error) {
	deadline := time.Now().Add(pollTimeout)

	for {
		result, err := p.jimengQueryTask(reqKey, taskID)
		if err != nil {
			return nil, err
		}

		switch result.Status {
		case "done":
			return result, nil
		case "failed":
			return nil, fmt.Errorf("task failed: %s", result.Message)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout after %v, task still in status: %s", pollTimeout, result.Status)
		}

		fmt.Fprintf(os.Stderr, "  Status: %s, waiting %v...\n", result.Status, pollInterval)
		time.Sleep(pollInterval)
	}
}

// jimengMapTaskStatus maps Volcano Engine task status to internal status.
func jimengMapTaskStatus(volcStatus string) string {
	switch volcStatus {
	case "processing", "in_queue":
		return "pending"
	case "generating":
		return "running"
	case "done":
		return "done"
	case "not_found", "expired":
		return "failed"
	default:
		return "pending"
	}
}

// Jimeng response types

type jimengQueryResult struct {
	TaskID   string
	Status   string
	VideoURL string
	Message  string
}

type jimengSubmitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

type jimengQueryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID   string `json:"task_id"`
		Status   string `json:"status"`
		VideoURL string `json:"video_url"`
	} `json:"data"`
}
