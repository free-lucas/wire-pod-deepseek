# DeepSeek API Provider 接入 Spec

## Why
当前项目里虽然可以通过 `custom` provider 手动把 Ollama 当作 OpenAI 兼容接口接入，但用户无法在页面上直接选择 DeepSeek，也无法按官方文档约束完成明确、稳定的 DeepSeek 配置。为了替换本地 Ollama 模型并降低配置门槛，需要新增一个面向 DeepSeek 官方 API 的独立选型和配置流程。

## What Changes
- 新增 `DeepSeek` 作为知识问答的独立 provider 选项，不再要求用户通过 `custom` 手工填写 DeepSeek 兼容参数。
- 在配置页面中新增 DeepSeek 专属配置区，至少支持启用/停用、`API Key`、模型选择，并沿用现有 Prompt、Intent Graph、Save Chat、Commands Enable 能力。
- 后端在选择 `DeepSeek` provider 时，必须按 DeepSeek 官方文档使用 OpenAI 兼容格式接入：
- `base_url` 固定为 `https://api.deepseek.com`
- 对话接口固定为 `POST /chat/completions`
- 鉴权方式为 `Authorization: Bearer <DEEPSEEK_API_KEY>`
- 默认模型使用非废弃模型，优先 `deepseek-v4-flash`，并允许选择 `deepseek-v4-pro`
- 当用户启用 DeepSeek 后，知识问答请求必须走 DeepSeek 官方云端接口，而不是本地 Ollama endpoint。
- 保持现有 `custom` provider 兼容，避免影响仍然使用 Ollama 或其他 OpenAI 兼容服务的用户。

## Impact
- Affected specs: 知识问答提供商选择、Web 配置页、LLM 请求路由、配置持久化
- Affected code: `chipper/webroot/setup.html`、`chipper/webroot/js/main.js`、`chipper/pkg/vars/config.go`、`chipper/pkg/wirepod/config-ws/webserver.go`、`chipper/pkg/wirepod/ttr/kgsim.go`、`chipper/pkg/wirepod/ttr/kgsim_cmds.go`

## ADDED Requirements
### Requirement: DeepSeek 作为独立知识问答提供商
系统 SHALL 在知识问答提供商列表中提供 `DeepSeek` 独立选项，并将其与 `custom` provider 明确区分。

#### Scenario: 用户选择 DeepSeek
- **WHEN** 用户打开设置页中的 Knowledge Graph 配置
- **THEN** 页面应展示 `DeepSeek` provider 选项
- **AND** 选择该选项后只显示 DeepSeek 所需字段，而不是通用 custom endpoint 表单

### Requirement: DeepSeek 页面配置能力
系统 SHALL 允许用户通过页面配置是否启用 DeepSeek、DeepSeek API Key、DeepSeek 模型，并保留现有知识问答高级开关。

#### Scenario: 保存 DeepSeek 配置
- **WHEN** 用户选择 `DeepSeek` provider，填写 API Key，勾选启用，并提交设置
- **THEN** 前端应向 `/api/set_kg_api` 提交 `provider=deepseek`
- **AND** 后端应将 DeepSeek 配置持久化到 `apiConfig.json`
- **AND** 重新加载页面后，`/api/get_kg_api` 返回的数据应能正确回填页面状态

#### Scenario: 关闭知识问答
- **WHEN** 用户在页面上选择 `None` 或关闭知识问答能力后提交设置
- **THEN** 系统应将知识问答配置标记为禁用
- **AND** 不再向 DeepSeek 或 Ollama 发起知识问答请求

### Requirement: 严格遵循 DeepSeek 官方 OpenAI 兼容接入方式
系统 SHALL 严格按照 DeepSeek 官方文档接入聊天补全接口，不得依赖用户自行填写 DeepSeek endpoint。

#### Scenario: 构造 DeepSeek 客户端
- **WHEN** 后端检测到 `provider=deepseek`
- **THEN** 必须使用 `https://api.deepseek.com` 作为 `base_url`
- **AND** 必须调用官方 `chat/completions` 接口完成对话请求
- **AND** 必须使用 Bearer Token 方式传递用户配置的 API Key

#### Scenario: 选择 DeepSeek 模型
- **WHEN** 用户未显式填写模型
- **THEN** 系统应默认使用 `deepseek-v4-flash`
- **AND** 不应再把已弃用的 `deepseek-chat` 或 `deepseek-reasoner` 作为默认模型

### Requirement: DeepSeek 替换本地 Ollama 路由
系统 SHALL 在启用 DeepSeek 后，将知识问答请求从本地 Ollama 路由切换到 DeepSeek 官方云端接口。

#### Scenario: 替换请求目标
- **WHEN** 当前保存的 provider 为 `deepseek`
- **THEN** LLM 客户端初始化不得读取本地 Ollama `endpoint`
- **AND** 请求目标必须是 DeepSeek 官方服务
- **AND** 现有 `custom` provider 仍可继续用于 Ollama

## MODIFIED Requirements
### Requirement: 知识问答 provider 配置
系统 SHALL 支持 `None`、`OpenAI`、`Houndify`、`Together`、`Custom`、`DeepSeek` 六类知识问答 provider，并根据 provider 展示和保存对应字段。

## REMOVED Requirements
