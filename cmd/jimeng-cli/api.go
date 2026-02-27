package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/volcengine/volc-sdk-golang/service/visual"
)

const (
	// 即梦视频生成3.0 Pro req_key
	videoGen3ProT2VReqKey          = "jimeng_t2v_v30_pro"      // 文生视频模式
	videoGen3ProI2VReqKey          = "jimeng_ti2v_v30_pro"     // 图生视频-首帧模式
	videoGen3ProI2VStartEndReqKey  = "jimeng_ti2v_v30_pro"     // 图生视频-首尾帧模式

	pollInterval = 5 * time.Second
	pollTimeout  = 300 * time.Second
)

// provider wraps the Volcano Engine visual client.
type provider struct {
	client *visual.Visual
}

// newProvider creates a provider with the given access keys.
func newProvider(accessKeyID, secretAccessKey string) *provider {
	client := visual.NewInstance()
	client.Client.SetAccessKey(accessKeyID)
	client.Client.SetSecretKey(secretAccessKey)
	return &provider{client: client}
}

// maskSecret masks a secret string for logging, showing only first 4 and last 4 chars.
func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// submitTask submits a video generation task and returns the task ID.
func (p *provider) submitTask(prompt, firstFrameImage, firstFrameImageBase64, aspectRatio string, frames, seed int) (string, error) {
	// Determine req_key based on input
	reqKey := videoGen3ProT2VReqKey
	if firstFrameImageBase64 != "" || firstFrameImage != "" {
		reqKey = videoGen3ProI2VReqKey
	}

	reqBody := map[string]interface{}{
		"req_key": reqKey,
		"prompt":  prompt,
	}

	// Add first frame image (image-to-video mode), prefer base64
	if firstFrameImageBase64 != "" {
		reqBody["binary_data_base64"] = []string{firstFrameImageBase64}
	} else if firstFrameImage != "" {
		reqBody["image_urls"] = []string{firstFrameImage}
	}

	if seed != 0 {
		reqBody["seed"] = seed
	}

	if aspectRatio != "" {
		reqBody["aspect_ratio"] = aspectRatio
	}

	if frames > 0 {
		reqBody["frames"] = frames
	} else {
		reqBody["frames"] = 121 // default 5 seconds
	}

	log.Printf("[jimeng] Submitting task: req_key=%s, prompt=%s", reqKey, prompt)

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

	var result submitTaskResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBytes))
	}

	if result.Code != 10000 {
		return "", fmt.Errorf("API error: code=%d, message=%s", result.Code, result.Message)
	}

	if result.Data.TaskID == "" {
		return "", fmt.Errorf("no task ID in response: %s", string(respBytes))
	}

	log.Printf("[jimeng] Task submitted: %s", result.Data.TaskID)
	return result.Data.TaskID, nil
}

// queryTask queries the status of a video generation task.
func (p *provider) queryTask(taskID string) (*queryResult, error) {
	reqBody := map[string]interface{}{
		"req_key": videoGen3ProI2VReqKey, // query uses fixed req_key per API docs
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

	var result queryTaskResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBytes))
	}

	if result.Code != 10000 {
		return &queryResult{
			TaskID:  taskID,
			Status:  "failed",
			Message: result.Message,
		}, nil
	}

	status := mapTaskStatus(result.Data.Status)

	qr := &queryResult{
		TaskID:   taskID,
		Status:   status,
		VideoURL: result.Data.VideoURL,
		Message:  result.Message,
	}

	return qr, nil
}

// waitForTask polls until the task completes or times out.
func (p *provider) waitForTask(taskID string) (*queryResult, error) {
	deadline := time.Now().Add(pollTimeout)

	for {
		result, err := p.queryTask(taskID)
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

// mapTaskStatus maps Volcano Engine task status to internal status.
func mapTaskStatus(volcStatus string) string {
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

func downloadVideo(url, outputPath string) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return 0, fmt.Errorf("write file: %w", err)
	}

	return n, nil
}

// Response types

type queryResult struct {
	TaskID   string
	Status   string
	VideoURL string
	Message  string
}

type submitTaskResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

type queryTaskResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID   string `json:"task_id"`
		Status   string `json:"status"`
		VideoURL string `json:"video_url"`
	} `json:"data"`
}
