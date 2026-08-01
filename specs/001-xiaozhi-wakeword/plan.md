# Implementation Plan: 将机器人唤醒词改为“小智，小智”

**Branch**: `001-xiaozhi-wakeword` | **Date**: 2026-07-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-xiaozhi-wakeword/spec.md`

## Summary

本次改造目标是把机器人默认唤醒体验从 “hey, vector” 迁移到 “小智，小智”，并保持唤醒后的问答链路继续可用。实现策略遵循 `chipper` 优先原则：先在 `chipper/webroot/sdkapp/`、`chipper/pkg/wirepod/sdkapp/`、`chipper/pkg/wirepod/config-ws/` 和 `chipper/pkg/wirepod/localization/` 内统一页面文案、机器人 locale 设置入口、中文 STT 前置条件和验证文档；同时仅在必要时触及 `vector-cloud/internal/voice/process.go`，因为当前代码证据显示仓库内并没有直接配置热词文本 `"hey, vector"` 的位置，`vector-cloud` 更像是在消费机器人已经识别完成的 `Hotword` 事件。

因此，这份计划将本次需求拆成两层边界：

1. **仓库内可直接实现的部分**：中文相关 UI/配置入口、机器人 `locale` 设置、中文 STT 前置条件、验证流程、兼容性和迁移说明。
2. **高风险外部依赖部分**：真实语音热词从 `"hey, vector"` 改为 `"小智，小智"` 是否能生效，可能取决于机器人固件或上游唤醒检测，而不是仅靠本仓库 Go 代码。

## Boundary Lock

本次实现阶段对边界做如下锁定，后续代码改动必须遵守：

1. **可以直接在仓库内实现的内容**：
   - `chipper` setup 页面中的 `zh-CN` 入口、回显和提示文案
   - `sdkapp` 中与 `locale`、`button_wakeword`、远程触发按钮相关的页面、接口和兼容说明
   - `jdocs/jdocs.json` 的 `vic.RobotSettings` 缓存链路说明
   - `quickstart.md`、`data-model.md`、`contracts/` 中的验证分层与已知限制记录
2. **只能显式记录风险、不能在本轮假定已解决的内容**：
   - 真实热词文本 `"小智，小智"` 是否会被机器人原生识别
   - 机器人固件或上游热词检测是否支持中文热词替换
   - `vector-cloud` 是否具备中文 `LanguageCode` 映射；若缺失，只能改成显式回退和清晰日志，不能伪装成“已支持中文热词”
3. **实施判定规则**：
   - 若页面、配置和 jdoc 层完成，但真实语音热词仍失败，必须记录为“外部热词能力限制”
   - 若 `vector-cloud/internal/voice/process.go` 仍无法把 `zh-CN` 映射到中文语言码，必须显式记录为“语言映射回退”，而不是静默吞掉

## Technical Context

**Language/Version**: Go（主实现），HTML + JavaScript（静态页面）

**Primary Dependencies**: `vector-go-sdk`、仓库内现有 gRPC/protobuf 接口、`sdkapp`、`config-ws`、`localization`、`preqs`、`vector-cloud/internal/voice`

**Storage**: 本地 JSON 文件与机器人 jdocs；主要涉及 `chipper/apiConfig.json`、`chipper/jdocs/jdocs.json`、`chipper/jdocs/botSdkInfo.json`

**Testing**: 以真实机器人手工验证为主；文档化验证步骤写入 `quickstart.md`，代码层以最小范围编译/静态检查为辅

**Target Platform**: 主机侧 `chipper` 服务 + 机器人侧 `vector-cloud` 协作环境，运行在现有 wire-pod 支持的平台上

**Project Type**: Go 单仓库应用，包含 Web 配置页、SDK 控制页、gRPC 服务和机器人兼容层

**Performance Goals**: 不引入额外常驻服务或明显增加语音链路时延；唤醒后问答链路保持与现状同级别可用性

**Constraints**: 必须优先在 `chipper/` 落地；不得新增大依赖或大重构；中文链路依赖 `zh-CN` STT 模型；真实热词文本切换可能受机器人固件/协议限制

**Scale/Scope**: 1 个功能分支；预计影响 1 组 spec 文档、`chipper` 下若干 UI/接口文件，以及最多 1 个 `vector-cloud` 语言映射点

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] 默认将工作范围限定在 `chipper/`；只有在验证 `locale -> LanguageCode` 链路时，才评估 `vector-cloud/internal/voice/process.go`
- [x] 所有配置、JSON、setup 页面或机器人侧契约改动，都已写清存储位置、默认值、向后兼容和迁移或回退策略
- [x] 验证方案覆盖主用户流程和至少一个相邻回归检查，并写明具体页面、接口和机器人操作
- [x] 计划中写清了受影响模块、入口、数据文件和用户可见影响
- [x] 已记录关键确认点：仓库内未发现直接修改热词文本的实现位置，实现阶段必须按“仓库内可改范围”和“固件/外部依赖范围”分开处理
- [x] 不引入新依赖、新服务或大重构；如需触及 `vector-cloud`，仅限最小范围修正现有语言映射

## Project Structure

### Documentation (this feature)

```text
specs/001-xiaozhi-wakeword/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── wakeword-surface.md
└── tasks.md
```

### Source Code (repository root)

```text
chipper/
├── start.sh
├── apiConfig.json
├── jdocs/
│   ├── jdocs.json
│   └── botSdkInfo.json
├── pkg/
│   ├── initwirepod/
│   ├── vars/
│   └── wirepod/
│       ├── config-ws/
│       ├── localization/
│       ├── preqs/
│       ├── sdkapp/
│       └── ttr/
└── webroot/
    ├── setup.html
    ├── js/
    └── sdkapp/
        ├── settings.html
        └── js/

vector-cloud/
└── internal/
    ├── clad/cloud/
    └── voice/
```

**Structure Decision**: 采用现有 wire-pod 单仓库结构，不创建并行子项目。本次优先改 `chipper` 的页面、接口和配置入口；`vector-cloud` 仅作为语音 hotword 事件和 locale 透传链路的风险点记录与必要修正点。

## Phase 0: Outline & Research

Phase 0 的调研结论见 [research.md](./research.md)，核心输出如下：

1. `chipper` 侧确实是本次需求的首选落点：
   - SDK 页面文案位于 `chipper/webroot/sdkapp/settings.html`
   - SDK 页面逻辑位于 `chipper/webroot/sdkapp/js/main.js` 与 `chipper/webroot/sdkapp/js/heyvector.js`
   - 机器人设置更新入口位于 `chipper/pkg/wirepod/sdkapp/server.go` 与 `chipper/pkg/wirepod/sdkapp/urlreqs.go`
   - STT 语言与配置落盘位于 `chipper/pkg/wirepod/config-ws/webserver.go` 和 `chipper/pkg/vars/config.go`
2. `apiConfig.json` 保存的是 wire-pod 主配置，包含 STT 语言；机器人 `locale` 与 `button_wakeword` 更接近机器人 `RobotSettings`，会被缓存到 `jdocs/jdocs.json`，不会写入 `botSdkInfo.json`
3. `vector-cloud` 侧目前只消费 `Hotword` 事件，并使用 `locale` 选择后续语音流语言；仓库内没有发现直接把 `"hey, vector"` 改成别的字符串的热词文本实现位置
4. `vector-cloud/internal/voice/process.go` 中的 `getLanguage()` 当前并不支持 `zh-CN`，未知语言会回退到 `ENGLISH_US`，这是中文链路中的明确风险点
5. 实现阶段不会把“页面文案已切换”错误等同于“真实热词文本已替换”，验证和结论必须分层记录

## Phase 1: Design Artifacts

- 领域实体与关系定义见 [data-model.md](./data-model.md)
- 受影响接口与用户表面契约见 [contracts/wakeword-surface.md](./contracts/wakeword-surface.md)
- 端到端验证步骤见 [quickstart.md](./quickstart.md)

## 影响地图（Impact Map）

- **受影响目录**:
  - `chipper/webroot/sdkapp/`
  - `chipper/pkg/wirepod/sdkapp/`
  - `chipper/pkg/wirepod/config-ws/`
  - `chipper/pkg/wirepod/localization/`
  - `chipper/pkg/vars/`
  - `vector-cloud/internal/voice/`
  - `specs/001-xiaozhi-wakeword/`

- **入口 / 请求路径**:
  - setup 页面 `setup.html` -> `webroot/js/main.js` -> `/api/set_stt_info` -> `config-ws/webserver.go` -> `vars.APIConfig.STT.Language` -> `apiConfig.json`
  - SDK 页面 `sdkapp/settings.html` -> `sdkapp/js/main.js` -> `/api-sdk/locale` -> `sdkapp/server.go` -> `sdkapp/urlreqs.go` -> 机器人 `/v1/update_settings`
  - SDK 页面 `sdkapp/settings.html` -> `sdkapp/js/heyvector.js` -> `/api-sdk/trigger_wake_word` -> `sdkapp/server.go` -> 机器人 `:8889/consolevarset`
  - 机器人热词事件 -> `vector-cloud/internal/voice/process.go` -> `getLanguage(locale)` -> Chipper 流式请求 -> `preqs` / `ttr`

- **配置 / 数据文件**:
  - `chipper/apiConfig.json`: STT 服务与 `STT.Language`
  - `chipper/jdocs/jdocs.json`: 机器人 `vic.RobotSettings` 缓存，包含 `locale`、`button_wakeword`
  - `chipper/jdocs/botSdkInfo.json`: 机器人连接信息，不保存唤醒词设置

- **用户可见变化**:
  - SDK 页面中与 “Hey Vector” 相关的默认文案将统一为“小智，小智”
  - SDK 页面将补齐中文 `locale` 展示或说明
  - 用户验证文档将改为中文唤醒路径，并明确旧词不再作为默认入口

- **兼容性 / 迁移说明**:
  - 不新增持久化文件，不重命名已有 JSON 文件
  - `apiConfig.json` 继续沿用原结构，只调整 `STT.Language`
  - `jdocs/jdocs.json` 继续作为机器人设置缓存来源，不做破坏性迁移
  - 为避免外部依赖导致功能被误判，计划中明确区分“页面/配置层已切换”和“真实语音热词已生效”两个验证层级
  - 若 `vector-cloud` 仍无中文语言码映射，则保留显式英文回退，并在日志、验证文档和最终报告中明确标注，不将其包装成已完成能力
  - 如果实施后确认真实热词文本无法通过现有仓库修改，则回退方案是保留本次文档与页面改造，并在最终报告中标记该需求被机器人固件/外部唤醒能力阻塞

- **待确认事项**:
  - 当前仓库未发现直接配置热词文本的位置；进入实现前，需要接受如下边界：本仓库可以先完成中文 STT、UI、设置入口和验证文档的统一，但“真实语音说出小智，小智即可唤醒”是否最终成立，可能仍取决于固件或上游热词检测能力

## Implementation Strategy

### 1. `chipper` 优先的最小改动路径

先在 `chipper` 内完成以下动作：

1. 在 `sdkapp` 页面层统一唤醒文案与用户提示
2. 在 `sdkapp` 接口层统一 `locale` 和 `button_wakeword` 的设置入口与兼容命名
3. 在 `config-ws` / `vars` 层确认 `zh-CN` 的 STT 前置条件与落盘方式
4. 在 Spec Kit 产物中固化验证路径、兼容性说明和风险边界

### 2. `vector-cloud` 仅做必要修正

只有在以下条件成立时才触及 `vector-cloud`：

1. 已确认 `chipper` 层改造完成，但中文链路仍因 `locale -> LanguageCode` 回退而失效
2. 改动可以收敛在 `vector-cloud/internal/voice/process.go` 的最小范围内
3. 不需要引入新协议、新服务或大规模结构调整

### 3. 实施优先级

1. 先确保中文 STT 和 SDK 页面/接口语义一致
2. 再确认机器人设置读取与回显一致
3. 最后验证真实机器人语音热词是否受仓库内改动影响
4. 如验证表明热词文本不可由本仓库控制，则停止继续扩大实现范围，回到澄清/收敛阶段

## Validation Plan

完整验证步骤见 [quickstart.md](./quickstart.md)。本计划要求至少覆盖以下两类验证：

1. **主流程验证**:
   - 选择 `zh-CN` STT 语言
   - 在 SDK 页面确认中文文案与机器人 locale 设置
   - 对机器人说“小智，小智”，随后提问一个普通问题
   - 观察是否进入响应流程并返回答案

2. **相邻回归检查**:
   - 分别验证只说一次“小智”与继续说旧词“hey, vector”时的表现
   - 确认现有页面设置读取、机器人连接和普通问答链路没有因本次改动退化

若当前环境无法完成真实机器人验证，仍需完成页面/API 层验证，并在 `quickstart.md` 中记录缺失条件和预期阻塞点。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 需要评估 `vector-cloud/internal/voice/process.go` | 当前 `getLanguage()` 对 `zh-CN` 会回退到 `ENGLISH_US`，这会影响中文链路 | 仅改 `chipper` 无法解释 `locale` 在机器人热词事件进入后的语言映射行为 |

## Current Implementation Status

截至当前实现阶段，已经确认并落地了以下结论：

1. setup 页面和 SDK 页面都已补齐 `zh-CN` 相关入口或说明
2. `sdkapp` 读取 `RobotSettings` 时，如果主机 STT 已经切到 `zh-CN`，会把机器人 locale 同步到 `zh-CN`
3. `vic.RobotSettings` 的缓存入口已经统一到同一条 host 侧缓存链路
4. `vector-cloud/internal/voice/process.go` 已把 `zh-CN` 无法映射中文 `LanguageCode` 的行为显式记录为英文回退
5. 因为依赖里的 `LanguageCode` 枚举没有中文值，所以真实中文语音链路仍然存在协议级限制
6. 真实机器人环境已验证：`/api-sdk/get_sdk_settings` 可读回 `locale = zh-CN`，但 spoken wakeword 层仍未恢复；“小智，小智”和“hey, vector”都不能让机器人进入聆听状态
7. 物理按钮唤醒后机器人可以正常回答问题，说明当前 host 侧主问答链路可用，阻塞点集中在机器人侧 hotword / spoken wakeword 能力
8. 兼容接口 `/api-sdk/trigger_wake_word` 当前不可作为主验证入口，因为它依赖机器人开放 `:8889/consolevarset`；当前实测 `192.168.1.39:8889` 为 `filtered`
