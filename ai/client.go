package ai

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

const WordCountPerSecond = 6 // 每秒生成的字数

const SystemPrompt = `
You are a video script generator. Your task is to generate video script based on the provided user input.
you need to generate long text which further be used to generate a video script base on the provided text, you can
extend the idea with conversational text / short story to refflect the idea / old chinese saying to reflect the idea.

response should be in chinese, total chinese characters response size should around {{.WordCount}}. then you have to cut response into sequence of
subtitle, with each subtitle translated into english. each subtitle should be between 10 to 28 characters in chinese,
and the english translation should not exceeding 14 words.
The main language of the script is Chinese, with English translation provided for each subtitle.

## Requirements:
1, response should follow strickly format specified in the response format section.
2, video scirpt should be in Chinese and English, with each subtitle in a separate line.
3, script should be in chinese first, followed by english translation.
4, response with "ERROR[actual_message]" if you can not generage a result, where the "actual_message" is where the real message.
5, do not respond any other information, just the script in the specified format.
6, english translate should all in lowercase, and no punctuation.


## Response format:
[{"cn":"中文字幕","en":"English Subtitle"},{"cn":"中文字幕","en":"English Subtitle"}]
`

var (
	instance *Client
	once     sync.Once
)

type Client struct {
	endpoint string         // API endpoint
	key      string         // API key
	model    string         // Model name
	client   *openai.Client // OpenAI client
}

type ScriptItem struct {
	ZhSubtitle string `json:"zh_subtitle"` // Chinese subtitle
	EnSubtitle string `json:"en_subtitle"` // English subtitle
}

func GetClient() *Client {
	if instance == nil {
		panic("instance already created, use GetClient() to get the instance")
	}

	return instance
}

func NewClient(model, key, endpoint string) *Client {
	c := &Client{
		model:    model,
		key:      key,
		endpoint: endpoint,
	}

	config := openai.DefaultConfig(c.key)
	config.BaseURL = c.endpoint
	c.client = openai.NewClientWithConfig(config)

	instance = c

	return c
}

func (c *Client) GenerateScript(ctx context.Context, prompt string, expectDuration time.Duration) (string, error) {
	log.Infof("Generating script with prompt: %s, expect duration: %s", prompt, expectDuration)

	req := openai.ChatCompletionRequest{
		Model: c.model,
	}

	calculatedWordCount := int(expectDuration.Seconds() * WordCountPerSecond)
	systemPromptWithWordCount := strings.ReplaceAll(SystemPrompt, "{{.WordCount}}", strconv.Itoa(calculatedWordCount))
	req.Messages = []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPromptWithWordCount,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	log.Infof("SystemPrompt: %s", systemPromptWithWordCount)
	log.Infof("user prompt: %s", prompt)
	log.Infof("Received response: %s", resp.Choices[0].Message.Content)

	content := resp.Choices[0].Message.Content
	if strings.HasPrefix(content, "ERROR[") {
		matched := strings.TrimPrefix(content, "ERROR[")
		matched = strings.TrimSuffix(matched, "]")

		return "", errors.New(matched)
	}

	return content, nil
}
