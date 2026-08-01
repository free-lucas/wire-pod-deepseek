# Tasks: 将机器人唤醒词改为“小智，小智”

**Input**: Design documents from `/specs/001-xiaozhi-wakeword/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `quickstart.md`, `contracts/`

**Tests**: 不强制先写自动化测试；每个用户故事都必须包含可复现的手工验证，另外补充针对改动模块的构建/运行检查。

**Organization**: 任务按用户故事分组，确保每个故事都可以独立实现和验证。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行执行（不同文件、无未完成前置依赖）
- **[Story]**: 任务所属用户故事（如 `US1`、`US2`、`US3`）
- 每条任务都包含明确文件路径

## Phase 1: Setup (范围锁定与实现准备)

**Purpose**: 把实现边界、验证基线和文档契约锁定，避免后续按错误假设改代码。

- [X] T001 在 `specs/001-xiaozhi-wakeword/plan.md`、`specs/001-xiaozhi-wakeword/research.md` 和 `specs/001-xiaozhi-wakeword/contracts/wakeword-surface.md` 复核并锁定“仓库内可改范围 / 外部热词限制”的实现边界
- [X] T002 [P] 在 `specs/001-xiaozhi-wakeword/quickstart.md` 补齐真实机器人环境、`zh-CN` 模型前置条件、缺失环境和阻塞项记录模板
- [X] T003 [P] 在 `specs/001-xiaozhi-wakeword/data-model.md` 补充 `HotwordEvent`、`RobotSettingsJdoc` 和 `STTConfig` 的最终实现检查点，避免后续实现偏离设计

---

## Phase 2: Foundational (阻塞性前置工作)

**Purpose**: 梳理 STT、locale、唤醒入口、jdoc 缓存和语音链路的共享基础，完成后再进入用户故事。

**⚠️ CRITICAL**: 本阶段未完成前，不进入用户故事实现。

- [X] T004 在 `chipper/pkg/wirepod/config-ws/webserver.go` 确认 `/api/set_stt_info` 对 `zh-CN` 的写入、落盘和 reload 行为
- [X] T005 [P] 在 `chipper/webroot/setup.html` 和 `chipper/webroot/js/main.js` 复核并补齐 `zh-CN` setup 页面入口与回显行为
- [X] T006 [P] 在 `chipper/pkg/wirepod/sdkapp/urlreqs.go` 统一 `locale` 与 `button_wakeword` 的机器人设置下发 helper
- [X] T007 在 `chipper/pkg/wirepod/sdkapp/server.go` 统一 `/api-sdk/locale`、`/api-sdk/button_hey_vector` 和 `/api-sdk/trigger_wake_word` 的兼容处理与注释说明
- [X] T008 [P] 在 `chipper/pkg/wirepod/sdkapp/jdocspinger.go` 和 `chipper/pkg/vars/vars.go` 确认 `vic.RobotSettings` 的缓存、回写和回显链路
- [X] T009 在 `vector-cloud/internal/voice/process.go` 明确 `zh-CN` 的 `locale -> LanguageCode` 行为；若协议不支持则增加显式限制说明并同步到 `specs/001-xiaozhi-wakeword/plan.md`
- [X] T010 [P] 在 `specs/001-xiaozhi-wakeword/quickstart.md` 增补“页面验证 / jdoc 回显验证 / 真实语音验证”三层验证步骤

**Checkpoint**: STT、locale、jdoc 缓存和语音链路边界已经清晰，可以开始分用户故事推进。

---

## Phase 3: User Story 1 - 中文唤醒并回答问题 (Priority: P1) 🎯 MVP

**Goal**: 让机器人能以“小智，小智”进入现有响应链路，并继续完成后续问答。

**Independent Test**: 对机器人说“小智，小智”并紧接着提出一个普通问题；若机器人成功进入响应状态并给出回答，则该故事成立。

### Implementation for User Story 1

- [X] T011 [US1] 在 `chipper/pkg/wirepod/sdkapp/server.go` 增加将目标机器人切换到 `zh-CN` locale 的默认设置或同步逻辑
- [X] T012 [P] [US1] 在 `chipper/webroot/sdkapp/settings.html` 增加 `zh-CN` locale 选项，并将默认唤醒文案改为“小智，小智”
- [X] T013 [P] [US1] 在 `chipper/webroot/sdkapp/js/main.js` 适配 `zh-CN` 当前值回显、`button_wakeword` 展示和中文唤醒文案
- [X] T014 [P] [US1] 在 `chipper/webroot/sdkapp/js/heyvector.js` 更新远程触发提示文案，同时保持现有远程触发能力
- [X] T015 [US1] 在 `vector-cloud/internal/voice/process.go` 落实中文 locale 进入现有问答链路的最小修正，或把受限行为显式记录到 `specs/001-xiaozhi-wakeword/quickstart.md`
- [X] T016 [US1] 在 `specs/001-xiaozhi-wakeword/quickstart.md` 写入“小智，小智 + 提问”的 10 次验证步骤、预期结果和记录表
- [X] T017 [US1] 运行针对 `chipper/pkg/wirepod/sdkapp/server.go`、`chipper/pkg/wirepod/config-ws/webserver.go` 和 `vector-cloud/internal/voice/process.go` 的构建/手工验证，并把结果记录到 `specs/001-xiaozhi-wakeword/quickstart.md`

**Checkpoint**: 到这里，用户故事 1 应该已经能独立演示和验证。

---

## Phase 4: User Story 2 - 替换旧唤醒方式 (Priority: P2)

**Goal**: 移除旧默认唤醒方式的用户可见入口，确保新旧口令不会同时作为默认方式出现。

**Independent Test**: 在同一台机器人上分别尝试“hey, vector”和“小智，小智”；只有新唤醒词能够拉起响应流程，则该故事成立。

### Implementation for User Story 2

- [X] T018 [US2] 在 `chipper/webroot/sdkapp/settings.html` 移除、重命名或降级展示默认 `Hey Vector` 入口
- [X] T019 [US2] 在 `chipper/webroot/sdkapp/js/main.js` 清理 `"Hey Vector"` 默认状态展示，并统一页面回显逻辑
- [X] T020 [US2] 在 `chipper/pkg/wirepod/sdkapp/server.go` 为旧接口 `/api-sdk/button_hey_vector` 和 `/api-sdk/trigger_wake_word` 增加兼容说明，避免它们继续作为默认推荐入口
- [X] T021 [P] [US2] 在 `specs/001-xiaozhi-wakeword/contracts/wakeword-surface.md` 同步默认入口已迁移到“小智，小智”的页面/接口契约
- [X] T022 [P] [US2] 在 `vector-cloud/README.md` 和 `docs/secondary-dev-guide.md` 更新旧词迁移说明与默认行为变化
- [X] T023 [US2] 在 `specs/001-xiaozhi-wakeword/quickstart.md` 写入“旧词失败 / 新词成功 / 旧接口仅兼容保留”的对照验证步骤
- [X] T024 [US2] 按 `specs/001-xiaozhi-wakeword/quickstart.md` 完成旧词与新词对照验证，并将结果记录回 `specs/001-xiaozhi-wakeword/quickstart.md`

**Checkpoint**: 到这里，用户故事 1 和用户故事 2 都应能独立成立，且默认行为已统一到中文唤醒词。

---

## Phase 5: User Story 3 - 降低误触发和使用困惑 (Priority: P3)

**Goal**: 让新唤醒词的触发边界更清晰，减少误触发和排查成本。

**Independent Test**: 用只说一次“小智”、噪音场景、后续问题识别失败和旧词误用进行尝试；若系统不会轻易错误进入完整问答流程，则该故事成立。

### Implementation for User Story 3

- [X] T025 [US3] 在 `vector-cloud/internal/voice/process.go` 增加或整理中文 locale、回退和异常场景日志，便于区分误触发、语言回退和链路限制
- [X] T026 [P] [US3] 在 `docs/core-flows.md` 补充“只说一次小智”“噪音场景”“后续问题识别失败”“旧词误用”的链路与排查说明
- [X] T027 [P] [US3] 在 `specs/001-xiaozhi-wakeword/quickstart.md` 增补误触发、噪音场景、后续问题识别失败和无关设置回归检查步骤
- [ ] T028 [US3] 按 `specs/001-xiaozhi-wakeword/quickstart.md` 完成单次“小智”、噪音场景、后续问题识别失败、旧词误用和无关设置回归验证，并记录结果回 `specs/001-xiaozhi-wakeword/quickstart.md`

**Checkpoint**: 到这里，全部用户故事都应具备独立验证路径。

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 收尾文档、兼容性说明、最终验证结论和已知限制。

- [X] T029 [P] 在 `docs/software-design.md` 和 `docs/architecture.md` 标注本次改动仍以 `chipper/` 为主，`vector-cloud/` 仅承担 hotword / locale 透传影响
- [X] T030 [P] 在 `specs/001-xiaozhi-wakeword/plan.md` 和 `specs/001-xiaozhi-wakeword/data-model.md` 同步最终实现结论、兼容性结论和受限点
- [X] T031 在 `specs/001-xiaozhi-wakeword/quickstart.md` 汇总最终验证结果、缺失环境条件、回退策略和已知限制

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup**: 无依赖，可立即开始
- **Phase 2: Foundational**: 依赖 Setup 完成；阻塞全部用户故事
- **Phase 3: User Story 1**: 依赖 Foundational 完成；这是 MVP
- **Phase 4: User Story 2**: 依赖 User Story 1 的中文唤醒基线已可用
- **Phase 5: User Story 3**: 依赖 User Story 1 已跑通，并复用 User Story 2 的默认行为结论
- **Phase 6: Polish**: 依赖全部目标用户故事完成

### User Story Dependencies

- **US1 (P1)**: Foundational 完成后即可开始；不依赖其他用户故事
- **US2 (P2)**: 依赖 US1 已建立中文唤醒基线
- **US3 (P3)**: 依赖 US1 的链路可用，并复用 US2 的默认行为定义

### Parallel Opportunities

- Phase 1 中 `T002` 与 `T003` 可并行执行
- Phase 2 中 `T005`、`T006`、`T008`、`T010` 可并行执行
- US1 中 `T012`、`T013`、`T014` 可在 `T011` 后并行执行
- US2 中 `T021` 与 `T022` 可并行执行
- US3 中 `T026` 与 `T027` 可并行执行
- Polish 阶段 `T029` 与 `T030` 可并行执行

---

## Parallel Example: User Story 1

```bash
Task: "在 chipper/webroot/sdkapp/settings.html 增加 zh-CN locale 选项，并将默认唤醒文案改为“小智，小智”"
Task: "在 chipper/webroot/sdkapp/js/main.js 适配 zh-CN 当前值回显、button_wakeword 展示和中文唤醒文案"
Task: "在 chipper/webroot/sdkapp/js/heyvector.js 更新远程触发提示文案，同时保持现有远程触发能力"
```

---

## Implementation Strategy

### MVP First (只先做 User Story 1)

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational
3. 完成 Phase 3: User Story 1
4. 按 `specs/001-xiaozhi-wakeword/quickstart.md` 做独立验证
5. 先确认“小智，小智 + 提问”链路是否稳定，再继续后续故事

### Incremental Delivery

1. Setup + Foundational 完成后，先交付 US1
2. 在 US1 稳定后，继续交付 US2，统一默认唤醒方式
3. 最后交付 US3，补足误触发、回退说明和排查路径
4. 每个故事完成后都回写 `specs/001-xiaozhi-wakeword/quickstart.md`

### Notes

- 当前最大的技术不确定性仍然是：真实热词文本是否能由现有仓库直接控制，还是主要受机器人固件 / 上游唤醒检测影响；因此 `T001` 和 `T009` 是必须先完成的边界锁定任务
- 如果 `T009` 的结论显示现有协议或语言码不支持中文热词链路，则后续实现必须在 `specs/001-xiaozhi-wakeword/plan.md`、`quickstart.md` 和最终报告中明确写出限制、替代路径和回退策略
