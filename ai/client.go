package ai

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

const WordCountPerSecond = 6 // 每秒生成的字数

const SystemPrompt = `
You are a video script generator. Your task is to generate video title, video subtitle in chinese, translation subtitle into english, 
and generate image generation prompt for each subtitle based on the provided user input.

you can extend the idea with conversational text / short story to refflect the idea / old chinese saying to reflect the idea.

response should be in chinese, total chinese characters response size should around {{.WordCount}}. then you have to cut response into sequence of
subtitle, with each subtitle translated into english. each subtitle should be between 10 to 28 characters in chinese,
and the english translation should not exceeding 14 words.
The main language of the script is Chinese, with English translation provided for each subtitle.

## Title requirements:
1, title should be in chinese, and should not exceed 20 characters.
2, title should be catchy and reflect the main idea of the video.
3, title should plain chinese, no english words or symbols.

## subtitle requirements:
1, video scirpt should be in Chinese
2, english translate should all in lowercase, and no punctuation.

## image generation prompt requirements:
1, image generation prompt should be in chinese, and should be descriptive enough for image generation.
2, the target image should be related to the subtitle.
3, the porompt must contain 火柴人，矢量图，黑白图标风格，简约，透明底色，线条粗细适中，线条清晰，线条流畅, 线条简洁, 线条不交叉, 线条不重叠, 线条不模糊, 线条不粗糙, 线条不复杂, 线条不花哨, 线条不夸张, 线条不浮夸, 线条不繁琐, 线条不冗余.

## response
1, response should be a json object with title and script_items, script_items field is an array of objects, each object should contain cn, en, image_prompt fields.
2, return only the json string iteself, do not add any other text, do not add any explanation or comments, and do not add any markdown code block.


## If you failed to generate content
response with "ERROR[actual_message]" if you can not generage a result, where the "actual_message" is where the real message.

## Response format:
{"title":"视频标题","script_items":[{"cn":"中文字幕","en":"English Subtitle","image_prompt":""},{"cn":"中文字幕","en":"English Subtitle","image_prompt":""}]}
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

func (c *Client) GenerateScript(ctx context.Context, prompt string,
	d time.Duration) (string, error) {
	log.Info().Msgf("Generating script with prompt: %s, expect duration: %s", prompt, d)

	req := openai.ChatCompletionRequest{
		Model: c.model,
	}

	calculatedWordCount := int(d.Seconds() * WordCountPerSecond)
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

	log.Info().Msgf("SystemPrompt: %s", systemPromptWithWordCount)
	log.Info().Msgf("user prompt: %s", prompt)
	log.Info().Msgf("Received response: %s", resp.Choices[0].Message.Content)

	content := resp.Choices[0].Message.Content
	if strings.HasPrefix(content, "ERROR[") {
		matched := strings.TrimPrefix(content, "ERROR[")
		matched = strings.TrimSuffix(matched, "]")

		return "", errors.New(matched)
	}

	return content, nil
}
