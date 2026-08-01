# TTS 提供商可配置化 Spec

## Why
当前中文 TTS 硬编码使用本地服务器 `192.168.1.35:8000`，该服务器不可用时 Vector 无法发出中文语音。DeepSeek 不提供 TTS API。需要将 TTS 后端改为可配置，推荐使用国内便宜好用的云 TTS 服务（火山引擎豆包 TTS），无需本地部署额外服务，树莓派可直接调用。

## What Changes
- 在 `apiConfig` 中新增 `TTS` 配置区块，支持 provider 选型和参数配置
- TTS 提供商选项：`volcengine`（火山引擎豆包 TTS，国内直连，1.3元/千字）、`local`（自建本地 TTS 服务器）
- `volcengine` 模式：Go 侧直接 HTTP POST 调用火山引擎 API，无需任何本地代理
- `local` 模式：可配置任意 HTTP TTS 服务器地址
- 网页配置页新增 TTS 设置区域
- `DoSayText` 路由逻辑改为读取 TTS 配置
- 本地 TTS 地址由硬编码改为可配置

## Impact
- Affected specs: add-deepseek-provider-config（共用 web 配置页）
- Affected code:
  - `chipper/pkg/vars/config.go` — 配置结构体，新增 TTS 区块
  - `chipper/pkg/wirepod/ttr/kgsim_cmds.go` — DoSayText 路由逻辑
  - `chipper/pkg/wirepod/ttr/say_chinese.go` — 本地 TTS 地址参数化
  - `chipper/pkg/wirepod/ttr/say_volcengine.go` — **新文件**，火山引擎 TTS 调用实现
  - `chipper/pkg/wirepod/config-ws/webserver.go` — 配置读写 API
  - `chipper/webroot/setup.html` — 前端表单
  - `chipper/webroot/js/main.js` — 前端逻辑

## ADDED Requirements

### Requirement: TTS 提供商配置
系统 SHALL 支持在网页配置页中选择 TTS 提供商，可选值为 `volcengine` 或 `local`。

#### Scenario: 用户选择火山引擎 TTS
- **WHEN** 用户在网页配置中选择 TTS 提供商为 "火山引擎豆包 TTS（推荐）"
- **THEN** 系统使用火山引擎 TTS API 生成中文语音
- **AND** 用户需配置 APP ID 和 Access Key（从火山引擎控制台获取）
- **AND** 用户可选择语音角色（如 zh_female_qingxin、zh_male_zhixing 等）
- **AND** Go 代码直接调用 `https://openspeech.bytedance.com/api/v3/tts/unidirectional`，无需本地代理

#### Scenario: 用户选择本地 TTS
- **WHEN** 用户在网页配置中选择 TTS 提供商为 "Local（自建 TTS 服务器）"
- **THEN** 系统使用用户配置的本地 TTS 服务器 URL
- **AND** 用户可配置服务器地址

#### Scenario: 英文场景不使用中文 TTS
- **WHEN** STT Language 为 "en-US"
- **THEN** 系统使用 Vector 内置语音，不走 TTS 提供商

### Requirement: 火山引擎 TTS 实现
系统 SHALL 通过 HTTP POST 直接调用火山引擎 API 实现中文语音合成，无需任何本地中间服务。

#### Scenario: 火山引擎语音生成
- **WHEN** 调用火山引擎 TTS 生成语音
- **THEN** 系统发送 HTTP POST 到 `https://openspeech.bytedance.com/api/v3/tts/unidirectional`
- **AND** 携带 Headers: `X-Api-App-Key`、`X-Api-Access-Key`、`X-Api-Resource-Id`（固定 `seed-tts-2.0`）、`X-Api-Request-Id`（UUID）
- **AND** 请求 Body 包含 text、speaker、audio_params（format=pcm, sample_rate=16000）
- **AND** 解析 NDJSON 流式响应，提取 base64 音频数据
- **AND** 拼接所有音频 chunks 为完整 PCM 16kHz 16bit mono 数据
- **AND** 通过 ExternalAudioStreamPlayback 在 Vector 上播放

#### Scenario: 火山引擎调用失败
- **WHEN** HTTP 请求失败或返回错误码
- **THEN** 记录错误日志，跳过该句语音，不阻塞后续对话

### Requirement: DoSayText 路由逻辑
系统 SHALL 根据 STT Language 和 TTS 配置决定语音输出方式。

#### Scenario: 非英文 + 火山引擎 TTS
- **WHEN** STT Language 不是 "en-US" 且 TTS Provider 为 "volcengine"
- **THEN** 调用火山引擎 TTS 生成语音并通过 ExternalAudioStreamPlayback 播放

#### Scenario: 非英文 + 本地 TTS
- **WHEN** STT Language 不是 "en-US" 且 TTS Provider 为 "local"
- **THEN** 调用配置的本地 TTS URL 生成语音并通过 ExternalAudioStreamPlayback 播放

#### Scenario: TTS 调用失败
- **WHEN** 任何 TTS 提供商调用失败（网络不通、鉴权失败等）
- **THEN** 系统记录错误日志，跳过该段语音，不阻塞后续对话

## MODIFIED Requirements

### Requirement: DoSayText 条件判断（修改）
原逻辑仅检查 `Provider == "openai"` 来决定是否走 TTS。修改为读取 TTS 配置区块的 `Provider` 字段，根据 Provider 类型路由到对应实现。

### Requirement: say_chinese.go 本地 TTS 地址（修改）
原 `callTTSAPI` 硬编码 `const ttsAPIURL = "http://192.168.1.35:8000/generate_voice"`。修改为从配置中读取用户设定的 URL。

### Requirement: DoSayText_DeepSeek 函数（删除）
之前临时添加的 `DoSayText_DeepSeek` 删除，由新的 TTS Provider 路由取代。

## REMOVED Requirements
无。
