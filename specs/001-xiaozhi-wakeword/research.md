# Research: 将机器人唤醒词改为“小智，小智”

## Locked Boundary

为了避免实现阶段误改，先把本轮允许落地的边界锁定如下：

1. 本轮默认只把“页面默认体验”“STT 中文前置条件”“机器人 locale 设置入口”“jdoc 缓存链路”“验证与限制文档”视为仓库内可交付范围。
2. 本轮不把“真实语音热词一定能从 hey, vector 变成 小智，小智”视为已知可解问题，而是保留为需要真实机器人和上游能力共同验证的高风险项。
3. 若 `vector-cloud` 侧没有中文 `LanguageCode`，本轮只能把回退行为显式化，并同步到日志和文档，不能把回退伪装成已支持中文热词。

## Decision 1: 优先把改动落在 `chipper` 的页面、设置入口和中文 STT 配置链路

**Rationale**:

- `sdkapp` 页面文案与按钮都在 `chipper/webroot/sdkapp/`
- 机器人设置写入入口在 `chipper/pkg/wirepod/sdkapp/server.go` 和 `urlreqs.go`
- STT 语言配置与落盘位于 `chipper/pkg/wirepod/config-ws/webserver.go` 和 `chipper/pkg/vars/config.go`
- 这符合项目宪章的 `chipper` 优先原则，也符合仓库现有职责边界

**Alternatives considered**:

- 直接从 `vector-cloud` 开始改：放弃，因为这会过早进入机器人兼容层，成本更高，而且无法先解决 UI、配置入口和验证文档的问题

## Decision 2: 将“真实热词文本替换”视为高风险外部依赖，而不是假设仓库内已有直接配置点

**Rationale**:

- 当前仓库中没有找到直接把 `"hey, vector"` 改成其他字符串的热词文本实现位置
- `vector-cloud/internal/clad/cloud/mic.go` 里的 `Hotword` 结构只有 `Mode`、`Locale`、`Timezone`、`NoLogging`，没有热词文本字段
- `vector-cloud/internal/voice/process.go` 的职责是消费热词事件并继续建立语音流，而不是做热词文本匹配
- `sdkapp` 中的 `"Hey Vector"` 更多是页面文案和远程触发按钮，不等于真实语音热词引擎

**Alternatives considered**:

- 假设仅改 UI 文案即可让真实语音热词生效：放弃，因为代码证据不足，而且很可能误导后续实现
- 假设 `button_wakeword` 就是可自由定义热词文本：放弃，因为现有协议更像是固定枚举而不是任意文本配置
- 在没有真实机器人验证前直接宣称需求已闭环：放弃，因为这会掩盖固件或上游热词能力限制

## Decision 3: 把中文链路的关键技术风险定位为 `zh-CN` 在 `vector-cloud` 的语言映射缺失

**Rationale**:

- `chipper` 已支持 `zh-CN` STT 模型与下载入口
- setup 页面也已暴露 `zh-CN` 选项
- 但 `vector-cloud/internal/voice/process.go` 的 `getLanguage()` 只显式处理 `en/fr/de`，未知语言会回退到 `ENGLISH_US`
- 这意味着即使页面和 STT 侧都改成中文，机器人热词事件进入后的语言映射仍可能退回英文链路

**Alternatives considered**:

- 完全忽略 `vector-cloud`：放弃，因为这会遗漏中文链路中最明显的技术风险点
- 直接大范围重构 `vector-cloud` 语言处理：放弃，因为当前只需要最小范围确认和必要修正
- 在无法确认协议支持时静默保持英文回退：放弃，因为这会让排查过程变得不透明

## Decision 4: 明确三类数据/配置的职责边界

**Rationale**:

- `apiConfig.json` 保存 wire-pod 主配置，负责 STT 服务和 `STT.Language`
- `jdocs/jdocs.json` 缓存机器人 `vic.RobotSettings`，可间接保存 `locale`、`button_wakeword`
- `jdocs/botSdkInfo.json` 保存机器人连接信息，不承载唤醒词设置
- 只有把这三者边界写清楚，后续任务拆解和兼容性说明才不会混乱

**Alternatives considered**:

- 把所有唤醒相关设置都视为 `apiConfig.json` 内容：放弃，因为这与实际代码和数据流不一致

## Decision 5: 验证结果必须分层记录

**Rationale**:

- 本次需求的“页面/配置已改好”和“真实语音热词已成功替换”并不一定能由一次代码改动同时保证
- 因此验证必须至少分成：
  1. 页面和接口层验证
  2. 机器人设置和 jdoc 回显验证
  3. 真实机器人语音唤醒验证
- 如果第 3 层失败，才能准确判断是仓库内实现问题，还是外部热词能力限制

**Alternatives considered**:

- 只做一次笼统的“试着说小智看看”：放弃，因为这不利于定位失败点，也不符合可复现验证要求
