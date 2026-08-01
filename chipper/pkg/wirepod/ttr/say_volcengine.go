package wirepod_ttr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/google/uuid"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

type volcengineTTSReq struct {
	User      volcengineUser      `json:"user"`
	ReqParams volcengineReqParams `json:"req_params"`
}

type volcengineUser struct {
	UID string `json:"uid"`
}

type volcengineReqParams struct {
	Text        string                `json:"text"`
	Speaker     string                `json:"speaker"`
	AudioParams volcengineAudioParams `json:"audio_params"`
}

type volcengineAudioParams struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
}

type volcengineTTSResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func DoSayText_Volcengine(robot *vector.Vector, input string) error {
	apiKey := vars.APIConfig.TTS.VolcengineAPIKey
	if apiKey == "" {
		logger.Println("Volcengine TTS error: API Key not configured")
		return nil
	}
	voice := vars.APIConfig.TTS.VolcengineVoice
	requestID := uuid.New().String()

	reqBody := volcengineTTSReq{
		User: volcengineUser{
			UID: "wire-pod-vector",
		},
		ReqParams: volcengineReqParams{
			Text:    input,
			Speaker: voice,
			AudioParams: volcengineAudioParams{
				Format:     "pcm",
				SampleRate: 16000,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Println("Volcengine TTS error: failed to marshal request: " + err.Error())
		return err
	}

	req, err := http.NewRequest("POST", vars.VolcengineTTSURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Println("Volcengine TTS error: failed to create request: " + err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("X-Api-Resource-Id", vars.VolcengineResourceID)
	req.Header.Set("X-Api-Request-Id", requestID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Println("Volcengine TTS error: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Println("Volcengine TTS error: failed to read response: " + err.Error())
		return err
	}

	var pcmBytes []byte
	lines := bytes.Split(bodyBytes, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var segment volcengineTTSResp
		if err := json.Unmarshal(line, &segment); err != nil {
			logger.Println("Volcengine TTS error: failed to parse ndjson line: " + err.Error())
			continue
		}
		if segment.Code != 0 {
			logger.Println("Volcengine TTS error: response code " + segment.Message)
			continue
		}
		if segment.Data != "" {
			decoded, err := base64.StdEncoding.DecodeString(segment.Data)
			if err != nil {
				logger.Println("Volcengine TTS error: failed to decode base64 audio: " + err.Error())
				continue
			}
			pcmBytes = append(pcmBytes, decoded...)
		}
	}

	if len(pcmBytes) == 0 {
		logger.Println("Volcengine TTS error: no audio data returned")
		return nil
	}

	vclient, err := robot.Conn.ExternalAudioStreamPlayback(context.Background())
	if err != nil {
		return err
	}
	vclient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
				AudioFrameRate: 16000,
				AudioVolume:    100,
			},
		},
	})

	chunkSize := 1024
	go func() {
		for i := 0; i < len(pcmBytes); i += chunkSize {
			end := i + chunkSize
			if end > len(pcmBytes) {
				end = len(pcmBytes)
			}
			vclient.Send(&vectorpb.ExternalAudioStreamRequest{
				AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
					AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
						AudioChunkSizeBytes: uint32(len(pcmBytes[i:end])),
						AudioChunkSamples:   pcmBytes[i:end],
					},
				},
			})
			time.Sleep(time.Millisecond * 25)
		}
		vclient.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
				AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
			},
		})
	}()
	time.Sleep(pcmLength(pcmBytes) + time.Millisecond*50)
	return nil
}
