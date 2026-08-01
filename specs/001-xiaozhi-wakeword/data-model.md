# Data Model: 将机器人唤醒词改为“小智，小智”

## Entity 1: STTConfig

**来源**: `chipper/apiConfig.json`

**字段**:

- `provider`: 当前 STT 服务，例如 `vosk`、`whisper.cpp`
- `language`: 当前 STT 语言，例如 `en-US`、`zh-CN`

**约束**:

- 本次需求中，中文链路至少需要 `language = zh-CN`
- 如果 STT 服务是 `vosk` 或 `whisper.cpp`，对应中文模型必须已存在或可下载

**状态变化**:

- 初始状态：沿用当前系统已有 STT 配置
- 目标状态：支持将 STT 语言切换到 `zh-CN`
- 当前实现结论：setup 页面和 `/api/set_stt_info` 已以 `zh-CN` 为有效值，并会把结果写入 `apiConfig.json`

**Implementation Checkpoints**:

- `/api/set_stt_info` 成功后需要同时满足：`apiConfig.json` 更新、`PastInitialSetup = true`、相关 STT reload 被触发
- 若中文模型未下载，接口返回值和页面状态必须把“正在下载模型”明确展示出来

## Entity 2: RobotSettingsJdoc

**来源**: 机器人 `ROBOT_SETTINGS`，缓存到 `chipper/jdocs/jdocs.json`

**关键字段**:

- `locale`: 机器人 locale
- `button_wakeword`: 机器人按钮对应的唤醒行为

**约束**:

- 本次需求会读取和可能更新 `locale`
- `button_wakeword` 当前更像固定枚举，不应在 plan 阶段假设它支持任意中文热词文本

**状态变化**:

- 机器人通过 SDK 接口更新设置
- `sdkapp` 主动拉取 jdoc 后回写到本地缓存
- 当前实现结论：当主机 STT 为 `zh-CN` 时，`sdkapp` 读取设置会把 robot locale 同步到 `zh-CN`，并通过统一缓存入口写回 `jdocs/jdocs.json`

**Implementation Checkpoints**:

- `sdkapp/server.go` 的 `get_sdk_settings` 与 `jdocspinger.go` 必须走同一条 `vic.RobotSettings` 缓存语义
- `jdocs/jdocs.json` 应继续作为读取与回显来源，不新增并行缓存格式

## Entity 3: HotwordEvent

**来源**: `vector-cloud/internal/clad/cloud/mic.go`

**字段**:

- `Mode`
- `Locale`
- `Timezone`
- `NoLogging`

**约束**:

- 当前结构中没有热词文本字段
- 只能证明“机器人已经识别到热词后发送了事件”，不能证明仓库能直接定义热词文本内容

**状态变化**:

- 机器人侧检测到热词后产生事件
- `vector-cloud/internal/voice/process.go` 消费事件并创建后续语音流
- 当前实现结论：`zh-CN` locale 会被显式记录为“回退到 English (US)”的受限行为

**Implementation Checkpoints**:

- `Locale` 为空时必须有显式默认行为
- 若 `Locale = zh-CN` 仍会回退英文，必须通过日志和文档把该限制暴露出来

## Entity 4: WakewordPresentation

**来源**: `chipper/webroot/sdkapp/settings.html` 和相关 JS

**字段**:

- 页面展示文案
- 远程触发状态提示
- 当前 `button_wakeword` 展示名称
- `locale` 可选项

**约束**:

- 必须与本次需求的默认交互语义一致
- 页面文案切换不能被误认为等同于真实热词引擎已切换
- 当前实现结论：默认页面文案和按钮展示已迁移到“小智，小智”，旧接口只作为兼容入口保留

**Implementation Checkpoints**:

- SDK 页面默认文案、按钮说明、远程触发提示和设置回显必须保持一致
- 旧词相关入口如因兼容性保留，必须以“兼容入口”而不是“默认入口”的方式表述

## Relationships

1. `STTConfig.language` 决定 `chipper` 使用哪种 STT 语言模型
2. `RobotSettingsJdoc.locale` 影响机器人与语音链路的 locale 信息
3. `HotwordEvent.Locale` 会被 `vector-cloud` 映射到后续请求的语言码
4. `WakewordPresentation` 负责把系统当前默认体验展示给用户

## Validation Rules

1. 只有当 `STTConfig.language = zh-CN` 且中文模型可用时，中文 STT 前提才算满足
2. 只有当 `WakewordPresentation`、`RobotSettingsJdoc` 和真实语音验证结论一致时，才能认定默认唤醒体验已经完成迁移
3. 如果 `HotwordEvent` 链路仍回退英文，则应判定为语言映射或外部热词能力问题，而不是单纯页面问题
