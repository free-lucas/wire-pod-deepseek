# Debug Session: robot-wakeword-loss
- **Status**: [OPEN]
- **Issue**: 机器人在当前环境下无法通过 spoken wakeword 被唤醒；“小智，小智”和“hey, vector”都不能进入聆听状态，但物理按钮唤醒后可以正常回答。
- **Debug Server**: http://192.168.1.58:7777/event
- **Log File**: .dbg/trae-debug-log-robot-wakeword-loss.ndjson

## Reproduction Steps
1. 确认 `wire-pod` 已启动，`/api-sdk/get_sdk_info` 可返回机器人 `003097ff`
2. 确认 `/api-sdk/get_sdk_settings?serial=003097ff` 返回 `locale = zh-CN`
3. 在真实机器人待机状态下说“小智，小智”或“hey, vector”
4. 观察机器人没有进入聆听状态
5. 改用物理按钮唤醒，机器人可以进入聆听状态并正常回答

## Hypotheses & Verification
| ID | Hypothesis | Likelihood | Effort | Evidence |
|----|------------|------------|--------|----------|
| A | 机器人侧 hotword / spoken wakeword 引擎未运行或失效，导致根本没有产出唤醒事件 | High | Medium | Confirmed |
| B | 机器人当前固件或运行模式禁用了远场唤醒，但保留了按钮唤醒后的问答链路 | High | Medium | Pending |
| C | 机器人麦克风常驻监听链路异常，但按钮唤醒后的录音链路仍正常 | Medium | Medium | Pending |
| D | `vector-cloud` 或 host 侧其实收到了唤醒相关事件，但日志与可见行为不足，造成“看起来没唤醒” | Low | Low | Rejected |

## Log Evidence
- 已在 `vector-cloud/internal/voice/process.go` 增加 runtime instrumentation：
  - `D:incoming-message`：记录进入 voice process 的非音频消息类型
  - `A:hotword-received`：记录 hotword 的 locale / mode / robotVersion
  - `D:stream-open`：记录云端语音流是否成功打开
  - `D:intent-result`：记录是否收到了意图结果
- 下一步通过真实机器人分别复现“spoken wakeword 尝试”和“物理按钮唤醒”来对比日志
- 第一次复现后 `.dbg/trae-debug-log-robot-wakeword-loss.ndjson` 为 `0` 行，说明当前部署环境没有命中 `vector-cloud/internal/voice/process.go` 的插桩
- 已确认树莓派启动路径来自 `chipper/start.sh`，因此第二轮改为在 `chipper/pkg/wirepod/sdkapp/server.go` 的 SDK EventStream 上插桩，直接观察机器人是否上报 `wake_word`
- **第二轮复现结果分析**：
  - 用户物理按键唤醒了两次，对应日志中分别产生了两次 `wake_word` 事件（带有 `"intentHeard": false` 作为起点，并在随后收到 `"intentHeard": true` 和意图 JSON 负载）。
  - 用户口头呼喊 "Hey Vector" 或 "小智，小智" 时，**没有任何日志产生**。
  - 这证明了 `chipper` 的 EventStream 监听是完全正常的，但是机器人侧**根本没有发出语音唤醒的事件**。

## Verification Conclusion
- **结论**：完全推翻假设 D，证实了假设 A。机器人侧的语音热词引擎（Hotword Engine）目前处于死机或失效状态。
- **推测根因**：我们将机器人的 `locale` 强行改为了 `zh-CN`。由于 Vector 机器人的官方固件内部并没有 `zh-CN` 对应的热词声学模型（如 `.pmdl` 或 Sensory 模型），这可能导致机器人内部的 `vic-engine` 启动热词引擎失败，从而变成了“聋子”（连原本的 Hey Vector 也不认了）。
- **下一步验证**：将机器人的 locale 恢复为 `en-US`，测试 "Hey Vector" 是否能恢复工作。如果能恢复，说明纯改代码配置无法实现中文唤醒词，必须深入机器人内部系统替换声学模型文件。
