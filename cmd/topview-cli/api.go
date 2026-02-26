package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/llm-net/llm-api-plugin/internal/httpclient"
)

const (
	topviewBaseURL  = "https://api.topview.ai/v1"
	pollInterval    = 5 * time.Second
	pollTimeout     = 600 * time.Second // 10 minutes
	maxPollAttempts = 120
)

// TopView API response wrapper

type topviewAPIResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// Upload types

type uploadCredential struct {
	FileID    string `json:"fileId"`
	UploadURL string `json:"uploadUrl"`
	FileName  string `json:"fileName"`
}

// Task types

type submitRequest struct {
	AvatarSourceFrom string `json:"avatarSourceFrom"`
	ImageFileID      string `json:"imageFileId,omitempty"`
	AudioSourceFrom  string `json:"audioSourceFrom"`
	AudioFileID      string `json:"audioFileId,omitempty"`
	ModeType         string `json:"modeType"`
}

type submitResult struct {
	TaskID    string `json:"taskId"`
	Status    string `json:"status"`
	ErrorMsg  string `json:"errorMsg"`
	SubTaskID string `json:"subTaskId"`
}

type queryResult struct {
	TaskID         string `json:"taskId"`
	Status         string `json:"status"`
	ErrorMsg       string `json:"errorMsg"`
	OutputVideoURL string `json:"outputVideoUrl"`
}

func authHeaders(apiKey, uid string) map[string]string {
	h := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}
	if uid != "" {
		h["Topview-Uid"] = uid
	}
	return h
}

func parseTopviewResponse(body []byte, statusCode int) (*topviewAPIResponse, error) {
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(body))
	}

	var resp topviewAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w\nraw: %s", err, string(body))
	}

	if resp.Code != "200" {
		return nil, fmt.Errorf("TopView API error: %s (code: %s)", resp.Message, resp.Code)
	}

	return &resp, nil
}

// Upload flow: credential → S3 PUT → check

func getUploadCredential(apiKey, uid, format string) (*uploadCredential, error) {
	url := fmt.Sprintf("%s/upload/credential?format=%s", topviewBaseURL, format)
	body, status, err := httpclient.GetJSON(url, authHeaders(apiKey, uid))
	if err != nil {
		return nil, err
	}

	resp, err := parseTopviewResponse(body, status)
	if err != nil {
		return nil, err
	}

	var cred uploadCredential
	if err := json.Unmarshal(resp.Result, &cred); err != nil {
		return nil, fmt.Errorf("parse credential: %w", err)
	}

	return &cred, nil
}

func uploadFileToS3(uploadURL string, data []byte, contentType string) error {
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{Timeout: httpclient.DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func checkUpload(apiKey, uid, fileID string) (bool, error) {
	url := fmt.Sprintf("%s/upload/check?fileId=%s", topviewBaseURL, fileID)
	body, status, err := httpclient.GetJSON(url, authHeaders(apiKey, uid))
	if err != nil {
		return false, err
	}

	resp, err := parseTopviewResponse(body, status)
	if err != nil {
		return false, err
	}

	var result bool
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return false, fmt.Errorf("parse result: %w", err)
	}

	return result, nil
}

func uploadFile(apiKey, uid string, data []byte, format, contentType string) (string, error) {
	cred, err := getUploadCredential(apiKey, uid, format)
	if err != nil {
		return "", fmt.Errorf("get credential: %w", err)
	}

	if err := uploadFileToS3(cred.UploadURL, data, contentType); err != nil {
		return "", fmt.Errorf("S3 upload: %w", err)
	}

	for i := 0; i < 10; i++ {
		ok, err := checkUpload(apiKey, uid, cred.FileID)
		if err != nil {
			return "", fmt.Errorf("check upload: %w", err)
		}
		if ok {
			return cred.FileID, nil
		}
		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("upload check timed out for fileId: %s", cred.FileID)
}

// Task flow: submit → poll → download

func submitVideoAvatarTask(apiKey, uid, imageFileID, audioFileID string) (*submitResult, error) {
	req := submitRequest{
		AvatarSourceFrom: "3", // user local photo
		ImageFileID:      imageFileID,
		AudioSourceFrom:  "0", // uploaded audio
		AudioFileID:      audioFileID,
		ModeType:         "2", // avatar4
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	respBody, status, err := httpclient.PostJSON(
		topviewBaseURL+"/video_avatar/task/submit",
		authHeaders(apiKey, uid),
		bodyBytes,
	)
	if err != nil {
		return nil, err
	}

	resp, err := parseTopviewResponse(respBody, status)
	if err != nil {
		return nil, err
	}

	var result submitResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse submit result: %w", err)
	}

	return &result, nil
}

func queryVideoAvatarTask(apiKey, uid, taskID string) (*queryResult, error) {
	url := fmt.Sprintf("%s/video_avatar/task/query?taskId=%s", topviewBaseURL, taskID)
	body, status, err := httpclient.GetJSON(url, authHeaders(apiKey, uid))
	if err != nil {
		return nil, err
	}

	resp, err := parseTopviewResponse(body, status)
	if err != nil {
		return nil, err
	}

	var result queryResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse query result: %w", err)
	}

	return &result, nil
}

func waitForTask(apiKey, uid, taskID string) (*queryResult, error) {
	deadline := time.Now().Add(pollTimeout)

	for i := 0; i < maxPollAttempts; i++ {
		result, err := queryVideoAvatarTask(apiKey, uid, taskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: query failed (attempt %d): %v\n", i+1, err)
			time.Sleep(pollInterval)
			continue
		}

		switch result.Status {
		case "done", "completed", "success":
			if result.OutputVideoURL == "" {
				return nil, fmt.Errorf("task completed but no output video URL")
			}
			return result, nil
		case "failed", "error":
			errMsg := result.ErrorMsg
			if errMsg == "" {
				errMsg = "unknown error"
			}
			return nil, fmt.Errorf("task failed: %s", errMsg)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout after %v, task still in status: %s", pollTimeout, result.Status)
		}

		fmt.Fprintf(os.Stderr, "  Status: %s, waiting %v...\n", result.Status, pollInterval)
		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("polling timed out after %d attempts", maxPollAttempts)
}

func downloadVideo(videoURL, outputPath string) (int64, error) {
	resp, err := http.Get(videoURL)
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

// File format helpers

func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".m4a":
		return "audio/m4a"
	case ".aac":
		return "audio/aac"
	default:
		return "application/octet-stream"
	}
}

func getImageFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "png"
	case ".webp":
		return "webp"
	default:
		return "jpg"
	}
}

func getAudioFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".wav":
		return "wav"
	case ".m4a":
		return "m4a"
	case ".aac":
		return "aac"
	default:
		return "mp3"
	}
}
