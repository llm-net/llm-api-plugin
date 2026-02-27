package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/volcengine/volc-sdk-golang/service/visual"
)

const (
	jimengActionImitationV2ReqKey = "jimeng_dreamactor_m20_gen_video"
)

// JimengActionImitationV2Provider 即梦动作模仿2.0 Provider
type JimengActionImitationV2Provider struct {
	client *visual.Visual
}

// NewJimengActionImitationV2Provider 创建即梦动作模仿2.0 Provider
func NewJimengActionImitationV2Provider(accessKeyID, secretAccessKey string) *JimengActionImitationV2Provider {
	client := visual.NewInstance()
	client.Client.SetAccessKey(accessKeyID)
	client.Client.SetSecretKey(secretAccessKey)
	return &JimengActionImitationV2Provider{
		client: client,
	}
}

// ActionImitationV2Request 动作模仿2.0请求参数
type ActionImitationV2Request struct {
	ImageURL       string // 输入图片URL
	ImageBase64    string // 输入图片base64（从本地文件读取）
	VideoURL       string // 模板视频URL
	CutFirstSecond *bool  // 是否裁剪结果视频的第1秒，默认true
}

// ActionImitationV2SubmitResult 提交任务结果
type ActionImitationV2SubmitResult struct {
	TaskID string
}

// ActionImitationV2QueryResult 查询任务结果
type ActionImitationV2QueryResult struct {
	TaskID    string
	Status    string // pending, running, done, failed
	VideoURL  string
	Message   string
	ErrorCode int // 错误码，10000 表示成功
}

// SubmitTask 提交动作模仿2.0任务
func (p *JimengActionImitationV2Provider) SubmitTask(ctx context.Context, req *ActionImitationV2Request) (*ActionImitationV2SubmitResult, error) {
	reqBody := map[string]interface{}{
		"req_key":   jimengActionImitationV2ReqKey,
		"video_url": req.VideoURL,
	}

	// base64 优先于 URL
	if req.ImageBase64 != "" {
		reqBody["binary_data_base64"] = []string{req.ImageBase64}
	} else if req.ImageURL != "" {
		reqBody["image_urls"] = []string{req.ImageURL}
	}

	// cut_result_first_second_switch 默认为 true，仅在显式设置时传递
	if req.CutFirstSecond != nil {
		reqBody["cut_result_first_second_switch"] = *req.CutFirstSecond
	}

	resp, statusCode, err := p.client.CVSync2AsyncSubmitTask(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("submit task failed with status code: %d", statusCode)
	}

	// 解析响应
	var result actionImitationV2SubmitResponse
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Code != 10000 {
		return nil, fmt.Errorf("submit task failed: code=%d, message=%s", result.Code, result.Message)
	}

	fmt.Fprintf(os.Stderr, "ActionImitationV2 SubmitTask success: taskID=%s, fullResponse=%s\n", result.Data.TaskID, string(respBytes))

	return &ActionImitationV2SubmitResult{
		TaskID: result.Data.TaskID,
	}, nil
}

// QueryTask 查询动作模仿2.0任务结果
func (p *JimengActionImitationV2Provider) QueryTask(ctx context.Context, taskID string) (*ActionImitationV2QueryResult, error) {
	reqJSON := `{"aigc_meta": {"content_producer": "001191440300192203821610000", "producer_id": "producer_id_test123", "content_propagator": "001191440300192203821610000", "propagate_id": "propagate_id_test123"}}`

	reqBody := map[string]interface{}{
		"req_key":  jimengActionImitationV2ReqKey,
		"task_id":  taskID,
		"req_json": reqJSON,
	}

	resp, statusCode, err := p.client.CVSync2AsyncGetResult(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	// 打印完整响应以便调试
	debugBytes, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stderr, "ActionImitationV2 QueryTask: statusCode=%d, taskID=%s, response=%s\n", statusCode, taskID, string(debugBytes))

	if statusCode != 200 {
		return nil, fmt.Errorf("query task failed with status code: %d", statusCode)
	}

	// 解析响应
	var result actionImitationV2QueryResponse
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 如果业务码不是 10000，返回失败状态和错误信息
	if result.Code != 10000 {
		return &ActionImitationV2QueryResult{
			TaskID:    taskID,
			Status:    "failed",
			Message:   result.Message,
			ErrorCode: result.Code,
		}, nil
	}

	// 映射状态
	status := mapActionImitationV2Status(result.Data.Status)

	queryResult := &ActionImitationV2QueryResult{
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

// mapActionImitationV2Status 映射火山引擎任务状态到内部状态
func mapActionImitationV2Status(volcStatus string) string {
	switch volcStatus {
	case "in_queue":
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

type actionImitationV2SubmitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

type actionImitationV2QueryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID         string `json:"task_id"`
		Status         string `json:"status"`           // in_queue, generating, done, not_found, expired
		VideoURL       string `json:"video_url"`        // 生成的视频URL（有效期为 1 小时）
		AIGCMetaTagged bool   `json:"aigc_meta_tagged"` // 隐式标识是否打标成功
	} `json:"data"`
}
