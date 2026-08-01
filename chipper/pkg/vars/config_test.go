package vars

import "testing"

func TestNormalizeKnowledgeConfigDeepSeekDefaultsAndClearsEndpoint(t *testing.T) {
	oldConfig := APIConfig
	t.Cleanup(func() {
		APIConfig = oldConfig
	})

	APIConfig = apiConfig{}
	APIConfig.Knowledge.Provider = " deepseek "
	APIConfig.Knowledge.Key = " deepseek-key "
	APIConfig.Knowledge.Model = " invalid-model "
	APIConfig.Knowledge.Endpoint = " https://example.invalid/v1 "

	changed := NormalizeKnowledgeConfig()

	if !changed {
		t.Fatal("expected normalization to report changes")
	}
	if APIConfig.Knowledge.Provider != KnowledgeProviderDeepSeek {
		t.Fatalf("provider = %q, want %q", APIConfig.Knowledge.Provider, KnowledgeProviderDeepSeek)
	}
	if APIConfig.Knowledge.Key != "deepseek-key" {
		t.Fatalf("key = %q, want %q", APIConfig.Knowledge.Key, "deepseek-key")
	}
	if APIConfig.Knowledge.Model != DeepSeekDefaultModel {
		t.Fatalf("model = %q, want %q", APIConfig.Knowledge.Model, DeepSeekDefaultModel)
	}
	if APIConfig.Knowledge.Endpoint != "" {
		t.Fatalf("endpoint = %q, want empty string", APIConfig.Knowledge.Endpoint)
	}
}

func TestNormalizeKnowledgeConfigCustomDefaultsKey(t *testing.T) {
	oldConfig := APIConfig
	t.Cleanup(func() {
		APIConfig = oldConfig
	})

	APIConfig = apiConfig{}
	APIConfig.Knowledge.Provider = " custom "
	APIConfig.Knowledge.Key = " "
	APIConfig.Knowledge.Model = " qwen:7b "
	APIConfig.Knowledge.Endpoint = " http://127.0.0.1:11434/v1 "

	changed := NormalizeKnowledgeConfig()

	if !changed {
		t.Fatal("expected normalization to report changes")
	}
	if APIConfig.Knowledge.Provider != KnowledgeProviderCustom {
		t.Fatalf("provider = %q, want %q", APIConfig.Knowledge.Provider, KnowledgeProviderCustom)
	}
	if APIConfig.Knowledge.Key != CustomKnowledgeDefaultKey {
		t.Fatalf("key = %q, want %q", APIConfig.Knowledge.Key, CustomKnowledgeDefaultKey)
	}
	if APIConfig.Knowledge.Model != "qwen:7b" {
		t.Fatalf("model = %q, want %q", APIConfig.Knowledge.Model, "qwen:7b")
	}
	if APIConfig.Knowledge.Endpoint != "http://127.0.0.1:11434/v1" {
		t.Fatalf("endpoint = %q, want %q", APIConfig.Knowledge.Endpoint, "http://127.0.0.1:11434/v1")
	}
}
