package zhipu

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

const (
	GLMPro  = "chatglm_pro"
	GLMStd  = "chatglm_std"
	GLMLite = "chatglm_lite"
	GLM6B   = "chatglm_6b"
)
