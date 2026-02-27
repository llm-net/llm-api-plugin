package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/llm-net/llm-api-plugin/internal/httpclient"
)

const (
	defaultModel = "doubao-seedance-1-5-pro-251215"
	baseURL      = "https://ark.cn-beijing.volces.com/api/v3"
	pollInterval = 5 * time.Second
	pollTimeout  = 300 * time.Second
)

// Request types

type CreateTaskRequest struct {
	Model      string             `json:"model"`
	Content    []TaskContent      `json:"content"`
	Parameters TaskParameters     `json:"parameters,omitempty"`
}

type TaskContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	URL  string `json:"image_url,omitempty"`
}

type TaskParameters struct {
	Duration   string `json:"duration,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	Ratio      string `json:"ratio,omitempty"`
	Audio      bool   `json:"with_audio"`
}

// Response types

type CreateTaskResponse struct {
	ID    string    `json:"id"`
	Error *APIError `json:"error,omitempty"`
}

type TaskResult struct {
	ID      string              `json:"id"`
	Status  string              `json:"status"`
	Content *TaskResultContent  `json:"content,omitempty"`
	Error   *APIError           `json:"error,omitempty"`
}

type TaskResultContent struct {
	VideoURL string `json:"video_url,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func authHeaders(apiKey string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + apiKey,
	}
}

func createTask(apiKey, model, prompt, resolution, duration, ratio, audio string) (string, error) {
	if model == "" {
		model = defaultModel
	}

	endpoint := baseURL + "/contents/generations/tasks"

	withAudio := audio != "false"

	req := CreateTaskRequest{
		Model: model,
		Content: []TaskContent{
			{Type: "text", Text: prompt},
		},
		Parameters: TaskParameters{
			Duration:   duration,
			Resolution: resolution,
			Ratio:      ratio,
			Audio:      withAudio,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	respBody, statusCode, err := httpclient.PostJSON(endpoint, authHeaders(apiKey), body)
	if err != nil {
		return "", err
	}

	if statusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", statusCode, string(respBody))
	}

	var resp CreateTaskResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBody))
	}

	if resp.Error != nil {
		return "", fmt.Errorf("API error [%s]: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.ID == "" {
		return "", fmt.Errorf("no task ID in response: %s", string(respBody))
	}

	return resp.ID, nil
}

func queryTask(apiKey, taskID string) (*TaskResult, error) {
	endpoint := fmt.Sprintf("%s/contents/generations/tasks/%s", baseURL, taskID)

	respBody, statusCode, err := httpclient.GetJSON(endpoint, authHeaders(apiKey))
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(respBody))
	}

	var result TaskResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBody))
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API error [%s]: %s", result.Error.Code, result.Error.Message)
	}

	return &result, nil
}

func waitForTask(apiKey, taskID string) (*TaskResult, error) {
	deadline := time.Now().Add(pollTimeout)

	for {
		result, err := queryTask(apiKey, taskID)
		if err != nil {
			return nil, err
		}

		switch result.Status {
		case "succeeded":
			return result, nil
		case "failed":
			msg := "task failed"
			if result.Error != nil {
				msg = fmt.Sprintf("task failed [%s]: %s", result.Error.Code, result.Error.Message)
			}
			return nil, fmt.Errorf("%s", msg)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout after %v, task still in status: %s", pollTimeout, result.Status)
		}

		fmt.Fprintf(os.Stderr, "  Status: %s, waiting %v...\n", result.Status, pollInterval)
		time.Sleep(pollInterval)
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
