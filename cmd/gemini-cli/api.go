package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/llm-net/llm-api-plugin/internal/httpclient"
)

const (
	defaultModel = "gemini-3-pro-image-preview"
	baseURL      = "https://generativelanguage.googleapis.com/v1beta/models/"
)

// Request types
type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type ImageConfig struct {
	AspectRatio string `json:"aspectRatio,omitempty"`
	ImageSize   string `json:"imageSize,omitempty"`
}

type GenerationConfig struct {
	ResponseModalities []string     `json:"responseModalities"`
	ImageConfig        *ImageConfig `json:"imageConfig,omitempty"`
}

type Request struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

// Response types
type ResponsePart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type ResponseContent struct {
	Parts []ResponsePart `json:"parts"`
}

type Candidate struct {
	Content ResponseContent `json:"content"`
}

type Response struct {
	Candidates []Candidate `json:"candidates"`
	Error      *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func generateContent(apiKey, model, prompt, aspectRatio, imageSize string) (*Response, error) {
	if model == "" {
		model = defaultModel
	}
	endpoint := baseURL + model + ":generateContent"

	req := Request{
		Contents: []Content{
			{Parts: []Part{{Text: prompt}}},
		},
		GenerationConfig: &GenerationConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	}

	if aspectRatio != "" || imageSize != "" {
		req.GenerationConfig.ImageConfig = &ImageConfig{
			AspectRatio: aspectRatio,
			ImageSize:   imageSize,
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	headers := map[string]string{
		"x-goog-api-key": apiKey,
	}

	respBody, statusCode, err := httpclient.PostJSON(endpoint, headers, body)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", statusCode, string(respBody))
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w\nraw: %s", err, string(respBody))
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("API error [%d] %s: %s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return &resp, nil
}
