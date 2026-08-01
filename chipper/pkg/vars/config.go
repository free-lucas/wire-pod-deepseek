package vars

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// a way to create a JSON configuration for wire-pod, rather than the use of env vars

var ApiConfigPath = "./apiConfig.json"

var APIConfig apiConfig

const (
	KnowledgeProviderOpenAI   = "openai"
	KnowledgeProviderTogether = "together"
	KnowledgeProviderCustom   = "custom"
	KnowledgeProviderDeepSeek = "deepseek"

	TogetherBaseURL           = "https://api.together.xyz/v1"
	TogetherDefaultModel      = "meta-llama/Llama-3-70b-chat-hf"
	DeepSeekBaseURL           = "https://api.deepseek.com"
	DeepSeekDefaultModel      = "deepseek-v4-flash"
	DeepSeekProModel          = "deepseek-v4-pro"
	CustomKnowledgeDefaultKey = "ollama"

	TTSProviderVolcengine = "volcengine"
	TTSProviderLocal      = "local"

	VolcengineTTSURL       = "https://openspeech.bytedance.com/api/v3/tts/unidirectional"
	VolcengineResourceID   = "seed-tts-2.0"
	VolcengineDefaultVoice = "zh_female_qingxin"
	LocalTTSDefaultURL     = "http://192.168.1.35:8000/generate_voice"
)

type apiConfig struct {
	Weather struct {
		Enable   bool   `json:"enable"`
		Provider string `json:"provider"`
		Key      string `json:"key"`
		Unit     string `json:"unit"`
	} `json:"weather"`
	Knowledge struct {
		Enable                 bool    `json:"enable"`
		Provider               string  `json:"provider"`
		Key                    string  `json:"key"`
		ID                     string  `json:"id"`
		Model                  string  `json:"model"`
		IntentGraph            bool    `json:"intentgraph"`
		RobotName              string  `json:"robotName"`
		OpenAIPrompt           string  `json:"openai_prompt"`
		OpenAIVoice            string  `json:"openai_voice"`
		OpenAIVoiceWithEnglish bool    `json:"openai_voice_with_english"`
		SaveChat               bool    `json:"save_chat"`
		CommandsEnable         bool    `json:"commands_enable"`
		Endpoint               string  `json:"endpoint"`
		TopP                   float32 `json:"top_p"`
		Temperature            float32 `json:"temp"`
	} `json:"knowledge"`
	TTS struct {
		Provider         string `json:"provider"`
		VolcengineAPIKey string `json:"volcengine_api_key"`
		VolcengineVoice  string `json:"volcengine_voice"`
		LocalURL         string `json:"local_url"`
	} `json:"TTS"`
	STT struct {
		Service  string `json:"provider"`
		Language string `json:"language"`
	} `json:"STT"`
	Server struct {
		// false for ip, true for escape pod
		EPConfig bool   `json:"epconfig"`
		Port     string `json:"port"`
	} `json:"server"`
	HasReadFromEnv   bool `json:"hasreadfromenv"`
	PastInitialSetup bool `json:"pastinitialsetup"`
}

func WriteConfigToDisk() {
	NormalizeKnowledgeConfig()
	NormalizeTTSConfig()
	logger.Println("Configuration changed, writing to disk")
	writeBytes, _ := json.Marshal(APIConfig)
	os.WriteFile(ApiConfigPath, writeBytes, 0644)
}

func CreateConfigFromEnv() {
	// if no config exists, create it
	if os.Getenv("WEATHERAPI_ENABLED") == "true" {
		APIConfig.Weather.Enable = true
		APIConfig.Weather.Provider = os.Getenv("WEATHERAPI_PROVIDER")
		APIConfig.Weather.Key = os.Getenv("WEATHERAPI_KEY")
		APIConfig.Weather.Unit = os.Getenv("WEATHERAPI_UNIT")
	} else {
		APIConfig.Weather.Enable = false
	}
	if os.Getenv("KNOWLEDGE_ENABLED") == "true" {
		APIConfig.Knowledge.Enable = true
		APIConfig.Knowledge.Provider = os.Getenv("KNOWLEDGE_PROVIDER")
		if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
			APIConfig.Knowledge.ID = os.Getenv("KNOWLEDGE_ID")
		}
		APIConfig.Knowledge.Key = os.Getenv("KNOWLEDGE_KEY")
	} else {
		APIConfig.Knowledge.Enable = false
	}
	WriteSTT()
	APIConfig.HasReadFromEnv = true
	WriteConfigToDisk()
}

func WriteSTT() {
	// was not part of the original code, so this is its own function
	// launched if stt not found in config
	APIConfig.STT.Service = os.Getenv("STT_SERVICE")
	if os.Getenv("STT_SERVICE") == "vosk" || os.Getenv("STT_SERVICE") == "whisper.cpp" {
		APIConfig.STT.Language = os.Getenv("STT_LANGUAGE")
	}
}

func NormalizeKnowledgeConfig() bool {
	changed := false

	knowledge := &APIConfig.Knowledge
	provider := strings.TrimSpace(knowledge.Provider)
	if knowledge.Provider != provider {
		knowledge.Provider = provider
		changed = true
	}

	key := strings.TrimSpace(knowledge.Key)
	if knowledge.Key != key {
		knowledge.Key = key
		changed = true
	}

	model := strings.TrimSpace(knowledge.Model)
	if knowledge.Model != model {
		knowledge.Model = model
		changed = true
	}

	endpoint := strings.TrimSpace(knowledge.Endpoint)
	if knowledge.Endpoint != endpoint {
		knowledge.Endpoint = endpoint
		changed = true
	}

	switch knowledge.Provider {
	case KnowledgeProviderTogether:
		if knowledge.Model == "" {
			logger.Println("Together knowledge provider missing model, defaulting to " + TogetherDefaultModel)
			knowledge.Model = TogetherDefaultModel
			changed = true
		}
	case KnowledgeProviderCustom:
		if knowledge.Key == "" {
			logger.Println("Custom knowledge provider missing API key, defaulting to 'ollama' for OpenAI-compatible backends")
			knowledge.Key = CustomKnowledgeDefaultKey
			changed = true
		}
	case KnowledgeProviderDeepSeek:
		if knowledge.Model == "" {
			logger.Println("DeepSeek knowledge provider missing model, defaulting to " + DeepSeekDefaultModel)
			knowledge.Model = DeepSeekDefaultModel
			changed = true
		}
		if knowledge.Model != DeepSeekDefaultModel && knowledge.Model != DeepSeekProModel {
			logger.Println("DeepSeek knowledge provider only supports " + DeepSeekDefaultModel + " or " + DeepSeekProModel + "; defaulting to " + DeepSeekDefaultModel)
			knowledge.Model = DeepSeekDefaultModel
			changed = true
		}
		if knowledge.Endpoint != "" {
			logger.Println("DeepSeek knowledge provider ignores custom endpoints; clearing saved endpoint")
			knowledge.Endpoint = ""
			changed = true
		}
	}

	return changed
}

func NormalizeTTSConfig() bool {
	changed := false
	tts := &APIConfig.TTS
	provider := strings.TrimSpace(tts.Provider)
	if tts.Provider != provider {
		tts.Provider = provider
		changed = true
	}
	if tts.Provider == "" {
		tts.Provider = TTSProviderVolcengine
		changed = true
	}
	if tts.VolcengineVoice == "" {
		tts.VolcengineVoice = VolcengineDefaultVoice
		changed = true
	}
	if tts.LocalURL == "" {
		tts.LocalURL = LocalTTSDefaultURL
		changed = true
	}
	return changed
}

func ReadConfig() {
	if _, err := os.Stat(ApiConfigPath); err != nil {
		CreateConfigFromEnv()
		logger.Println("API config JSON created")
	} else {
		// read config
		configBytes, err := os.ReadFile(ApiConfigPath)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to read API config file")
			logger.Println(err)
			return
		}
		err = json.Unmarshal(configBytes, &APIConfig)
		if err != nil {
			APIConfig.Knowledge.Enable = false
			APIConfig.Weather.Enable = false
			logger.Println("Failed to unmarshal API config JSON")
			logger.Println(err)
			return
		}
		// stt service is the only thing controlled by shell
		if APIConfig.STT.Service != os.Getenv("STT_SERVICE") {
			WriteSTT()
		}
		if !APIConfig.HasReadFromEnv {
			if APIConfig.Server.Port != os.Getenv("DDL_RPC_PORT") {
				APIConfig.HasReadFromEnv = true
				APIConfig.PastInitialSetup = true
			}
		}

		if APIConfig.Knowledge.Model == "meta-llama/Llama-2-70b-chat-hf" {
			logger.Println("Setting Together model to Llama3")
			APIConfig.Knowledge.Model = "meta-llama/Llama-3-70b-chat-hf"
		}

		NormalizeKnowledgeConfig()
		NormalizeTTSConfig()
		writeBytes, _ := json.Marshal(APIConfig)
		os.WriteFile(ApiConfigPath, writeBytes, 0644)
		logger.Println("API config successfully read")
	}
}
