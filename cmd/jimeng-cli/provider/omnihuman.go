package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/volcengine/volc-sdk-golang/service/visual"
)

const (
	// OmniHuman1.5 视频生成 req_key
	omniHumanVideoGenerationReqKey = "jimeng_realman_avatar_picture_omni_v15"
)

// JimengOmniHumanProvider 即梦OmniHuman1.5 Provider
type JimengOmniHumanProvider struct {
	client *visual.Visual
}

// NewJimengOmniHumanProvider 创建即梦OmniHuman1.5 Provider
func NewJimengOmniHumanProvider(accessKeyID, secretAccessKey string) *JimengOmniHumanProvider {
	client := visual.NewInstance()
	client.Client.SetAccessKey(accessKeyID)
	client.Client.SetSecretKey(secretAccessKey)
	return &JimengOmniHumanProvider{
		client: client,
	}
}

// OmniHumanRequest OmniHuman请求参数
type OmniHumanRequest struct {
	ImageURL         string // 人像图片URL（必填）
	ImageBase64      string // 人像图片base64（从本地文件读取）
	AudioURL         string // 音频URL（必填，时长必须小于60秒）
	Prompt           string // 提示词（可选，仅限中文、英语、日语、韩语、墨西哥语、印尼语，建议≤300字符）
	Seed             int    // 随机种子（可选，默认-1随机）
	OutputResolution int    // 输出分辨率（可选，720或1080，默认1080）
	FastMode         bool   // 快速模式（可选，默认false，开启会牺牲部分效果加快速度）
}

// OmniHumanSubmitResult 提交任务结果
type OmniHumanSubmitResult struct {
	TaskID string
}

// OmniHumanQueryResult 查询任务结果
type OmniHumanQueryResult struct {
	TaskID    string
	Status    string // pending, running, done, failed
	VideoURL  string
	Message   string
	ErrorCode int // 错误码，10000 表示成功
}

// SubmitTask 提交OmniHuman1.5视频生成任务
func (p *JimengOmniHumanProvider) SubmitTask(ctx context.Context, req *OmniHumanRequest) (*OmniHumanSubmitResult, error) {
	reqBody := map[string]interface{}{
		"req_key":   omniHumanVideoGenerationReqKey,
		"audio_url": req.AudioURL,
	}

	// base64 优先于 URL
	if req.ImageBase64 != "" {
		reqBody["binary_data_base64"] = []string{req.ImageBase64}
	} else if req.ImageURL != "" {
		reqBody["image_url"] = req.ImageURL
	}

	// 添加可选的提示词
	if req.Prompt != "" {
		reqBody["prompt"] = req.Prompt
	}

	// 添加可选的随机种子
	if req.Seed != 0 {
		reqBody["seed"] = req.Seed
	}

	// 添加可选的输出分辨率
	if req.OutputResolution == 720 || req.OutputResolution == 1080 {
		reqBody["output_resolution"] = req.OutputResolution
	}

	// 添加可选的快速模式
	if req.FastMode {
		reqBody["pe_fast_mode"] = true
	}

	resp, statusCode, err := p.client.CVSubmitTask(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("submit task failed with status code: %d", statusCode)
	}

	// 解析响应
	var result omniHumanSubmitTaskResponse
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Fprintf(os.Stderr, "OmniHuman SubmitTask: statusCode=%d, response=%s\n", statusCode, string(respBytes))

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Code != 10000 {
		return nil, fmt.Errorf("submit task failed: code=%d, message=%s", result.Code, result.Message)
	}

	return &OmniHumanSubmitResult{
		TaskID: result.Data.TaskID,
	}, nil
}

// QueryTask 查询OmniHuman1.5视频生成任务结果
func (p *JimengOmniHumanProvider) QueryTask(ctx context.Context, taskID string) (*OmniHumanQueryResult, error) {
	reqBody := map[string]interface{}{
		"req_key": omniHumanVideoGenerationReqKey,
		"task_id": taskID,
	}

	resp, statusCode, err := p.client.CVGetResult(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	// 打印完整响应以便调试
	debugBytes, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stderr, "OmniHuman QueryTask: statusCode=%d, taskID=%s, response=%s\n", statusCode, taskID, string(debugBytes))

	if statusCode != 200 {
		return nil, fmt.Errorf("query task failed with status code: %d", statusCode)
	}

	// 解析响应
	var result omniHumanQueryTaskResponse
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 如果业务码不是 10000，返回失败状态和错误信息
	if result.Code != 10000 {
		return &OmniHumanQueryResult{
			TaskID:    taskID,
			Status:    "failed",
			Message:   result.Message,
			ErrorCode: result.Code,
		}, nil
	}

	// 映射状态
	status := mapOmniHumanTaskStatus(result.Data.Status)

	queryResult := &OmniHumanQueryResult{
		TaskID:    taskID,
		Status:    status,
		Message:   result.Message,
		ErrorCode: result.Code,
	}

	// 如果任务完成，提取视频URL
	if status == "done" {
		queryResult.VideoURL = result.Data.VideoURL
	}

	return queryResult, nil
}

// mapOmniHumanTaskStatus 映射火山引擎任务状态到内部状态
// 火山引擎状态: processing, in_queue, generating, done, not_found, expired
func mapOmniHumanTaskStatus(volcStatus string) string {
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

// 火山引擎API响应结构

type omniHumanSubmitTaskResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

type omniHumanQueryTaskResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID         string `json:"task_id"`
		Status         string `json:"status"`           // processing, in_queue, generating, done, not_found, expired
		VideoURL       string `json:"video_url"`        // 生成的视频URL（有效期1小时）
		AIGCMetaTagged bool   `json:"aigc_meta_tagged"` // 隐式标识是否打标成功
	} `json:"data"`
}
