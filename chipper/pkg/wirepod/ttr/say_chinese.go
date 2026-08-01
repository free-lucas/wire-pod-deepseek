package wirepod_ttr

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ParamsRefineText 定义 params_refine_text 对象的结构
type ParamsRefineText struct {
	Prompt            string  `json:"prompt"`
	TopP              float64 `json:"top_P"`
	TopK              int     `json:"top_K"`
	Temperature       float64 `json:"temperature"`
	RepetitionPenalty float64 `json:"repetition_penalty"`
	MaxNewToken       int     `json:"max_new_token"`
	MinNewToken       int     `json:"min_new_token"`
	ShowTqdm          bool    `json:"show_tqdm"`
	EnsureNonEmpty    bool    `json:"ensure_non_empty"`
	StreamBatch       int     `json:"stream_batch"`
}

// ParamsInferCode 定义 params_infer_code 对象的结构
type ParamsInferCode struct {
	Prompt            string  `json:"prompt"`
	TopP              float64 `json:"top_P"`
	TopK              int     `json:"top_K"`
	Temperature       float64 `json:"temperature"`
	RepetitionPenalty float64 `json:"repetition_penalty"`
	MaxNewToken       int     `json:"max_new_token"`
	MinNewToken       int     `json:"min_new_token"`
	ShowTqdm          bool    `json:"show_tqdm"`
	EnsureNonEmpty    bool    `json:"ensure_non_empty"`
	StreamBatch       bool    `json:"stream_batch"`
	SpkEmb            *string `json:"spk_emb"` // 指针以允许 null
}

// RequestPayload 定义整个请求体的结构
type RequestPayload struct {
	Text                   []string         `json:"text"`
	Stream                 bool             `json:"stream"`
	Lang                   *string          `json:"lang"` // 指针以允许 null
	SkipRefineText         bool             `json:"skip_refine_text"`
	RefineTextOnly         bool             `json:"refine_text_only"`
	UseDecoder             bool             `json:"use_decoder"`
	AudioSeed              int              `json:"audio_seed"`
	TextSeed               int              `json:"text_seed"`
	DoTextNormalization    bool             `json:"do_text_normalization"`
	DoHomophoneReplacement bool             `json:"do_homophone_replacement"`
	ParamsRefineText       ParamsRefineText `json:"params_refine_text"`
	ParamsInferCode        ParamsInferCode  `json:"params_infer_code"`
}

// 您的TTS API服务器地址
//const ttsAPIURL = "http://192.168.1.35:5000/synthesize"
const ttsAPIURL = "http://192.168.1.35:8000/generate_voice"

// TTSResponse 用于解析TTS API的响应
type TTSResponse struct {
	AudioBase64 string `json:"audio"` // 假设 TTS API 返回的是 base64 编码的音频数据
}

//  调用TTS API并返回音频数据
func callTTSAPI(text string) ([]byte, error) {
	// 创建HTTP请求
	payload := RequestPayload{
		Text:                   []string{text},
		//Text:                   []string{"四川美食确实以辣闻名，但也有不辣的选择。比如甜水面、赖汤圆、蛋烘糕、叶儿粑等，就非常fashion，而且现在AI，我也很LIKE,甜而不腻，也很受欢迎。"},
		Stream:                 false,
		Lang:                   nil,
		SkipRefineText:         true,
		RefineTextOnly:         false,
		UseDecoder:             true,
		AudioSeed:              12345678,
		TextSeed:               87654321,
		DoTextNormalization:    true,
		DoHomophoneReplacement: false,
		ParamsRefineText: ParamsRefineText{
			Prompt:            "",
			TopP:              0.7,
			TopK:              20,
			Temperature:       0.7,
			RepetitionPenalty: 1,
			MaxNewToken:       384,
			MinNewToken:       0,
			ShowTqdm:          true,
			EnsureNonEmpty:    true,
			StreamBatch:       24,
		},
		ParamsInferCode: ParamsInferCode{
			Prompt:            "[speed_5]",
			TopP:              0.1,
			TopK:              20,
			Temperature:       0.3,
			RepetitionPenalty: 1.05,
			MaxNewToken:       2048,
			MinNewToken:       0,
			ShowTqdm:          true,
			EnsureNonEmpty:    true,
			StreamBatch:       true,
			SpkEmb:            nil,
		},
	}
	logger.Println("Alan----get response from LLM="+text)
	jsonData, _ := json.Marshal(payload)
	logger.Println("Alan---begin request TTS API")
	resp, err := http.Post(ttsAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	logger.Println("Alan---requested TTS API successfully")
	defer resp.Body.Close()

	// 解析响应
	var ttsResp TTSResponse
	if err := json.NewDecoder(resp.Body).Decode(&ttsResp); err != nil {
		return nil, err
	}

	// 解码 base64 音频数据
	audioData, err := base64.StdEncoding.DecodeString(ttsResp.AudioBase64)
	if err != nil {
		return nil, err
	}
	length := len(audioData)
	logger.Println("Length of binary data: %d bytes\n", length)

	// 将音频数据保存到文件
	writeAudioToFile(audioData)

	logger.Println("Alan---base64 DecodeString audio successfully")
	return audioData, nil
}
func generateRandomFileName(preName string) string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf(preName+"_%d.wav", rand.Int())
}

// 确保文件夹是否存在与否
func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755) // Create directory with 0755 permissions
		if err != nil {
			return err
		}
	}
	return nil
}

func saveWavFile(data []byte, sampleRate, bitsPerSample, numChannels int) error {
	// Define the directory and file name
	dir := "./tmp-audio/"
	fileName := generateRandomFileName("tts_converted_audio")
	filePath := filepath.Join(dir, fileName)

	// Ensure the directory exists
	if err := ensureDirectoryExists(dir); err != nil {
		fmt.Println("Error creating directory:", err)
		return err
	}

	// Create a temporary file with the generated name in the target directory
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return err
	}
	defer file.Close() // Ensure the file is closed when done

	// Specify the target directory
	//dir := "./wire-pod/"
	//
	//// Ensure the directory exists
	//if err := os.MkdirAll(dir, 0755); err != nil {
	//	fmt.Println("Error creating directory:", err)
	//	return err
	//}
	//file, err := ioutil.TempFile(dir, "tts_converted_audio_*.wav")
	if err != nil {
		return err
	}
	defer file.Close()

	// WAV header parameters
	byteRate := sampleRate * numChannels * (bitsPerSample / 8)
	blockAlign := numChannels * (bitsPerSample / 8)
	subChunk2Size := len(data)
	chunkSize := 36 + subChunk2Size

	// Write WAV header
	writeString(file, "RIFF")
	binary.Write(file, binary.LittleEndian, uint32(chunkSize))
	writeString(file, "WAVE")
	writeString(file, "fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16)) // Subchunk1Size for PCM
	binary.Write(file, binary.LittleEndian, uint16(1))  // AudioFormat for PCM
	binary.Write(file, binary.LittleEndian, uint16(numChannels))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))
	binary.Write(file, binary.LittleEndian, uint32(byteRate))
	binary.Write(file, binary.LittleEndian, uint16(blockAlign))
	binary.Write(file, binary.LittleEndian, uint16(bitsPerSample))
	writeString(file, "data")
	binary.Write(file, binary.LittleEndian, uint32(subChunk2Size))

	// Write audio data
	file.Write(data)
	logger.Println("Alan----write converted file name=" + file.Name())

	return nil
}

func writeString(file *os.File, s string) {
	file.WriteString(s)
}

// 将音频数据写入到临时文件
func writeAudioToFile(audioData []byte) (string, error) {
	// 写入到临时文件
	var buffer bytes.Buffer
	buffer.Write(audioData)

	// Define the directory and file name
	dir := "./tmp-audio/"
	fileName := generateRandomFileName("tts_audio")
	filePath := filepath.Join(dir, fileName)

	// Ensure the directory exists
	if err := ensureDirectoryExists(dir); err != nil {
		fmt.Println("Error creating directory:", err)
		return "", err
	}

	// Create a temporary file with the generated name in the target directory
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return "", err
	}
	defer file.Close() // Ensure the file is closed when done

	if _, err := file.Write(buffer.Bytes()); err != nil {
		return "", err
	}
	logger.Println("Alan----write no convert file name=" + file.Name())
	return file.Name(), nil
}