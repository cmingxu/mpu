package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

type Txt2Img struct {
	client *http.Client
	key    string
}

var txt2ImgInstance *Txt2Img

func NewTxt2Img(key string) *Txt2Img {
	c := &Txt2Img{
		key: key,
	}

	c.client = &http.Client{}
	txt2ImgInstance = c
	return txt2ImgInstance
}

func GetTxt2Img() *Txt2Img {
	return txt2ImgInstance
}

// curl -X POST https://ark.cn-beijing.volces.com/api/v3/images/generations \
//   -H "Content-Type: application/json" \
//   -H "Authorization: Bearer $ARK_API_KEY" \
//   -d '{
//     "model": "doubao-seedream-3-0-t2i-250415",
//     "prompt": "鱼眼镜头，一只猫咪的头部，画面呈现出猫咪的五官因为拍摄方式扭曲的效果。",
//     "response_format": "url",
//     "size": "1024x1024",
//     "seed": 12,
//     "guidance_scale": 2.5,
//     "watermark": true
// }'
//
//
// {
//   "model": "doubao-seedream-3-0-t2i-250415"
//   "created": 1589478378,
//   "data": [
//     {
//       "url": "https://..."
//     }
//   ],
//   "usage": {
//       "generated_images":1
//   }
// }

func (t *Txt2Img) GenerateImage(ctx context.Context, prompt string) ([]byte, error) {
	log.Debug().Msgf("Generating image with prompt: %s", prompt)

	if txt2ImgInstance == nil {
		return nil, errors.New("txt2img client not initialized")
	}

	body := map[string]interface{}{
		"model":           "doubao-seedream-3-0-t2i-250415",
		"prompt":          prompt,
		"response_format": "b64_json",
		"size":            "1280x720",
		"seed":            12,
		"guidance_scale":  2.5,
		"watermark":       false,
	}

	log.Debug().Msgf("Request body for txt2img API: %+v", body)
	log.Debug().Msgf("Request body for txt2img API: %s", t.key)

	raw, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequestWithContext(ctx, "POST",
		"https://ark.cn-beijing.volces.com/api/v3/images/generations", bytes.NewBuffer(raw))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.key)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request to txt2img API")
	}

	log.Debug().Msgf("txt2img API response status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("txt2img API returned status: %s %s", resp.Status, string(body))
	}

	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body from txt2img API")
	}

	log.Debug().Msgf("txt2img API response body: %s", string(raw))

	result := gjson.Parse(string(raw))
	imgResult := result.Get("data.0.b64_json")

	b64Decoded, err := base64.StdEncoding.DecodeString(imgResult.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base64 response from txt2img API")
	}

	log.Debug().Msgf("Decoded image data length: %d bytes", len(b64Decoded))

	return b64Decoded, nil
}
