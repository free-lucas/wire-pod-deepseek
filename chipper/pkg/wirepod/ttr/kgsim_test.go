package wirepod_ttr

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

func preserveKnowledgeTestState(t *testing.T) {
	t.Helper()

	oldConfig := vars.APIConfig
	oldChats := append([]vars.RememberedChat(nil), vars.RememberedChats...)
	oldConfigPath := vars.ApiConfigPath

	vars.ApiConfigPath = filepath.Join(t.TempDir(), "apiConfig.json")
	t.Cleanup(func() {
		vars.APIConfig = oldConfig
		vars.RememberedChats = oldChats
		vars.ApiConfigPath = oldConfigPath
	})
}

func readClientConfigString(t *testing.T, client *openai.Client, field string) string {
	t.Helper()

	clientValue := reflect.ValueOf(client).Elem()
	configField := clientValue.FieldByName("config")
	configValue := reflect.NewAt(configField.Type(), unsafe.Pointer(configField.UnsafeAddr())).Elem()
	targetField := configValue.FieldByName(field)
	return reflect.NewAt(targetField.Type(), unsafe.Pointer(targetField.UnsafeAddr())).Elem().String()
}

func TestNewKnowledgeClientDeepSeekUsesDeepSeekBaseURL(t *testing.T) {
	preserveKnowledgeTestState(t)

	vars.APIConfig.Knowledge.Provider = vars.KnowledgeProviderDeepSeek
	vars.APIConfig.Knowledge.Key = "deepseek-key"
	vars.APIConfig.Knowledge.Model = vars.DeepSeekProModel
	vars.APIConfig.Knowledge.Endpoint = " https://override.invalid/v1 "

	client, err := newKnowledgeClient()
	if err != nil {
		t.Fatalf("newKnowledgeClient returned error: %v", err)
	}
	if got := readClientConfigString(t, client, "BaseURL"); got != vars.DeepSeekBaseURL {
		t.Fatalf("BaseURL = %q, want %q", got, vars.DeepSeekBaseURL)
	}
	if got := readClientConfigString(t, client, "authToken"); got != "deepseek-key" {
		t.Fatalf("authToken = %q, want %q", got, "deepseek-key")
	}
	if vars.APIConfig.Knowledge.Endpoint != "" {
		t.Fatalf("endpoint = %q, want empty string", vars.APIConfig.Knowledge.Endpoint)
	}
}

func TestNewKnowledgeClientCustomUsesEndpointAndDefaultKey(t *testing.T) {
	preserveKnowledgeTestState(t)

	vars.APIConfig.Knowledge.Provider = vars.KnowledgeProviderCustom
	vars.APIConfig.Knowledge.Key = " "
	vars.APIConfig.Knowledge.Model = "qwen:7b"
	vars.APIConfig.Knowledge.Endpoint = " http://127.0.0.1:11434/v1 "

	client, err := newKnowledgeClient()
	if err != nil {
		t.Fatalf("newKnowledgeClient returned error: %v", err)
	}
	if got := readClientConfigString(t, client, "BaseURL"); got != "http://127.0.0.1:11434/v1" {
		t.Fatalf("BaseURL = %q, want %q", got, "http://127.0.0.1:11434/v1")
	}
	if got := readClientConfigString(t, client, "authToken"); got != vars.CustomKnowledgeDefaultKey {
		t.Fatalf("authToken = %q, want %q", got, vars.CustomKnowledgeDefaultKey)
	}
}

func TestNewKnowledgeClientCustomRequiresEndpoint(t *testing.T) {
	preserveKnowledgeTestState(t)

	vars.APIConfig.Knowledge.Provider = vars.KnowledgeProviderCustom
	vars.APIConfig.Knowledge.Key = "custom-key"
	vars.APIConfig.Knowledge.Model = "qwen:7b"

	client, err := newKnowledgeClient()
	if err == nil {
		t.Fatalf("expected error, got client %#v", client)
	}
	if !strings.Contains(err.Error(), "missing endpoint") {
		t.Fatalf("error = %q, want missing endpoint message", err.Error())
	}
}

func TestCreateAIReqDeepSeekUsesRememberedChatAndDefaultModel(t *testing.T) {
	preserveKnowledgeTestState(t)

	vars.APIConfig.Knowledge.Provider = vars.KnowledgeProviderDeepSeek
	vars.APIConfig.Knowledge.Model = ""
	vars.APIConfig.Knowledge.OpenAIPrompt = " Stay concise. "
	vars.APIConfig.Knowledge.SaveChat = true
	vars.RememberedChats = []vars.RememberedChat{{
		ESN: "esn-1",
		Chats: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: "hello"},
			{Role: openai.ChatMessageRoleAssistant, Content: "hi there"},
		},
	}}

	req, err := CreateAIReq("what can you do", "esn-1", false, false)
	if err != nil {
		t.Fatalf("CreateAIReq returned error: %v", err)
	}
	if req.Model != vars.DeepSeekDefaultModel {
		t.Fatalf("model = %q, want %q", req.Model, vars.DeepSeekDefaultModel)
	}
	if !req.Stream {
		t.Fatal("expected request to use streaming")
	}
	if len(req.Messages) != 4 {
		t.Fatalf("len(Messages) = %d, want 4", len(req.Messages))
	}
	if req.Messages[0].Role != openai.ChatMessageRoleSystem {
		t.Fatalf("first role = %q, want %q", req.Messages[0].Role, openai.ChatMessageRoleSystem)
	}
	if !strings.Contains(req.Messages[0].Content, "Stay concise.") {
		t.Fatalf("system prompt = %q, want configured prompt content", req.Messages[0].Content)
	}
	if !strings.Contains(req.Messages[0].Content, "speech-to-text software") {
		t.Fatalf("system prompt = %q, want generated prompt guidance", req.Messages[0].Content)
	}
	if req.Messages[1].Content != "hello" || req.Messages[2].Content != "hi there" {
		t.Fatalf("remembered chat = %#v, want preserved chat history", req.Messages[1:3])
	}
	if req.Messages[3].Role != openai.ChatMessageRoleUser || req.Messages[3].Content != "what can you do" {
		t.Fatalf("last message = %#v, want final user message", req.Messages[3])
	}
}

func TestCreateAIReqCustomUsesConfiguredModelAndDefaultPrompt(t *testing.T) {
	preserveKnowledgeTestState(t)

	vars.APIConfig.Knowledge.Provider = vars.KnowledgeProviderCustom
	vars.APIConfig.Knowledge.Model = "qwen:7b"
	vars.APIConfig.Knowledge.OpenAIPrompt = ""
	vars.APIConfig.Knowledge.SaveChat = false
	vars.APIConfig.Knowledge.CommandsEnable = false

	req, err := CreateAIReq("summarize this", "esn-2", false, false)
	if err != nil {
		t.Fatalf("CreateAIReq returned error: %v", err)
	}
	if req.Model != "qwen:7b" {
		t.Fatalf("model = %q, want %q", req.Model, "qwen:7b")
	}
	if len(req.Messages) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(req.Messages))
	}
	if !strings.Contains(req.Messages[0].Content, "You are a helpful, animated robot called Vector.") {
		t.Fatalf("system prompt = %q, want default Vector prompt", req.Messages[0].Content)
	}
	if req.Messages[1].Content != "summarize this" {
		t.Fatalf("user message = %q, want %q", req.Messages[1].Content, "summarize this")
	}
}
