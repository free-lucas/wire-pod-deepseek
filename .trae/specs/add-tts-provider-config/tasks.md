# Tasks

- [x] Task 1: 扩展配置结构体（config.go）
  - [x] 在 `apiConfig` 中新增 `TTS` 结构体：`Provider`（volcengine/local）、`VolcengineAppID`、`VolcengineAccessKey`、`VolcengineVoice`、`LocalURL`
  - [x] 新增 TTS 提供商常量 `TTSProviderVolcengine`、`TTSProviderLocal`
  - [x] 新增火山引擎默认值：Voice 默认 `zh_female_qingxin`，API 地址 `https://openspeech.bytedance.com/api/v3/tts/unidirectional`
  - [x] 新增 `NormalizeTTSConfig()` 设置默认值（Provider 默认 volcengine）
  - [x] 在 `WriteConfigToDisk()` 中调用 `NormalizeTTSConfig()`

- [x] Task 2: 新建火山引擎 TTS 调用模块（say_volcengine.go）
  - [x] 实现 `DoSayText_Volcengine(robot *vector.Vector, input string) error`
  - [x] 构造 HTTP POST 到 `https://openspeech.bytedance.com/api/v3/tts/unidirectional`
  - [x] 携带 Headers: `X-Api-App-Key`、`X-Api-Access-Key`、`X-Api-Resource-Id`（固定 `seed-tts-2.0`）、`X-Api-Request-Id`（UUID）
  - [x] 请求 Body: text、speaker、audio_params（format=pcm, sample_rate=16000）
  - [x] 解析 NDJSON 流式响应（每行一个 JSON，提取 base64 audio data）
  - [x] 将所有 PCM 音频 chunks 拼接，通过 ExternalAudioStreamPlayback 播放
  - [x] 错误处理：HTTP 失败、鉴权失败、返回空音频时记录日志并跳过

- [x] Task 3: 修改 Go 侧 TTS 路由（kgsim_cmds.go）
  - [x] 删除 `DoSayText_DeepSeek` 函数
  - [x] `DoSayText` 中非英文场景改为读取 `vars.APIConfig.TTS.Provider`
  - [x] Provider 为 `volcengine` 时调用 `DoSayText_Volcengine`
  - [x] Provider 为 `local` 时调用本地 TTS（复用 say_chinese.go 的 callTTSAPI，传入配置的 LocalURL）
  - [x] Provider 为 `openai` 时保留原有 OpenAI TTS 逻辑（向后兼容）

- [x] Task 4: 本地 TTS URL 参数化（say_chinese.go）
  - [x] `callTTSAPI` 改为接受 URL 参数
  - [x] 删除硬编码 `const ttsAPIURL`

- [x] Task 5: 配置读写 API（webserver.go）
  - [x] `handleSetKGAPI` 扩展支持 TTS 配置字段的保存（Provider、VolcengineAppID、VolcengineAccessKey、VolcengineVoice、LocalURL）
  - [x] `handleGetKGAPI` 扩展返回 TTS 配置字段

- [x] Task 6: 网页配置 UI（setup.html + main.js）
  - [x] 新增 TTS 配置区域：Provider 下拉（火山引擎豆包 TTS / 本地 TTS）
  - [x] 选择火山引擎时显示：APP ID、Access Key、Voice 选择框
  - [x] 选择本地 TTS 时显示：Local URL 输入框
  - [x] 保存逻辑：将 TTS 配置字段发送到 API
  - [x] 回填逻辑：页面加载时从 API 获取并填充

- [x] Task 7: 编译验证并部署到树莓派
  - [x] `go build` 验证编译通过
  - [x] 部署源码到树莓派并重启服务
  - [x] 验证网页可配置 TTS 提供商
  - [x] 引导用户在火山引擎控制台获取 APP ID 和 Access Key
  - [ ] 验证机器人中文发音正常（需用户配置 API Key 后测试）

# Task Dependencies
- Task 2 可独立并行
- Task 3 依赖 Task 1
- Task 4 可独立并行
- Task 5 依赖 Task 1
- Task 6 可与 Task 2-5 并行
- Task 7 依赖 Task 1-6
