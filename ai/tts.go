package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

const (
	TTSModel     = "FunAudioLLM/CosyVoice2-0.5B"
	DefaultVoice = "benjamin"
)

var (
	VoiceList = []string{
		"alex",
		"anna",
		"bella",
		"benjamin",
		"charles",
		"clarie",
		"diana",
		"david",
	}
)

var ttsInstance *Tts

type Tts struct {
	key      string
	endpoint string
	voice    string

	client *openai.Client
}

func NewTts(key string) *Tts {
	t := &Tts{
		key:      key,
		endpoint: "https://api.siliconflow.cn/v1/audio/speech",
		voice:    fmt.Sprintf("%s:%s", TTSModel, DefaultVoice),
	}

	config := openai.DefaultConfig(t.key)
	config.BaseURL = t.endpoint
	t.client = openai.NewClientWithConfig(config)

	ttsInstance = t

	return ttsInstance
}

func GetTTSInstance() *Tts {
	return ttsInstance
}

//	curl --request POST \
//	  --url https://api.siliconflow.cn/v1/audio/speech \
//	  --header 'Authorization: Bearer <token>' \
//	  --header 'Content-Type: application/json' \
//	  --data '{
//	  "model": "FunAudioLLM/CosyVoice2-0.5B",
//	  "input": "Can you say it with a happy emotion? <|endofprompt|>I'\''m so happy, Spring Festival is coming!",
//	  "voice": "FunAudioLLM/CosyVoice2-0.5B:alex",
//	  "response_format": "mp3",
//	  "sample_rate": 32000,
//	  "stream": true,
//	  "speed": 1,
//	  "gain": 0
//	}'
func (t *Tts) GenerateAudio(ctx context.Context, text string) ([]byte, error) {
	if t.client == nil {
		return nil, errors.New("TTS client is not initialized")
	}

	log.Info().Msgf("Generating audio for text: %s", text)

	data := map[string]interface{}{
		"model":           TTSModel,
		"input":           text,
		"voice":           t.voice,
		"response_format": "mp3",
		"sample_rate":     32000,
		"stream":          true,
		"gain":            0.0,
		"speed":           1.0,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal TTS request data")
	}

	req, _ := http.NewRequest(
		http.MethodPost,
		t.endpoint,
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set("Authorization", "Bearer "+t.key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create TTS request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS request failed with status code: %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read TTS response body")
	}

	return responseBody, nil

}
