# wire-pod 二开文档索引

> 本文档集与源码 commit `5b35c8ee` 逐层核对，覆盖架构、流程、程序设计、二次开发四个维度。

---

## 系统架构交互图

- **[wire-pod-architecture.html](../.archify/architecture-wire-pod-20251108-230500/wire-pod-architecture.html)**  
  交互式 SVG 架构图，含主机侧 / 机器人侧两大边界、全部组件和连接关系，支持 dark/light 主题切换。  
  校验状态：9/9 设计检查通过，4 个浏览器视口全绿，最小投影文字 6.3px（≥ 6px 底线）。

---

## 建议阅读顺序

| 顺序 | 文档 | 目标 |
| --- | --- | --- |
| 1 | [`project-overview.md`](./project-overview.md) | 仓库级别全局认识 |
| 2 | [`software-design.md`](./software-design.md) | 软件设计视角：上下文 / 目标 / 视图 |
| 3 | [`architecture.md`](./architecture.md) | 组件拆分、职责边界、进程关系 |
| 4 | [`program-design.md`](./program-design.md) ⭐新增 | 模块代码结构、关键数据结构、完整调用链 |
| 5 | [`core-flows.md`](./core-flows.md) | 启动 / 语音 / 配置 / 接入时序流程 |
| 6 | [`secondary-dev-guide.md`](./secondary-dev-guide.md) | 按需求反推代码入口，含新手避坑指南 |
| 7 | [架构交互图](../.archify/architecture-wire-pod-20251108-230500/wire-pod-architecture.html) | 可视化全局组件拓扑 |

---

## 这套系统先记住 3 件事

- `wire-pod` 不是单体应用，而是"主机侧服务 + 机器人侧兼容层"的组合：
  - **`chipper/`**：主机侧核心，STT · 意图匹配 · LLM/KG · 配置页 · SDK 控制 · 自定义意图 · 插件
  - **`vector-cloud/`**：机器人侧协议桥接，把 Engine/mic/ai_sock 消息转接到本地 chipper
- 日常二开里，**大部分业务改动集中在 `chipper/`**，尤其是 `preqs/`、`ttr/`、`config-ws/`、`sdkapp/`
- `vector-cloud/` 平时不用动，做深度机器人协议改造时才涉及

---

## 新手词汇表

| 术语 | 含义 |
| --- | --- |
| `main.go` | Go 程序入口文件 |
| `package` | Go 的代码分组方式（一个模块） |
| `handler` | 处理请求的函数 |
| `gRPC` | 高性能 RPC 框架，机器人与 chipper 之间大量使用 |
| `STT` | Speech To Text，语音转文字引擎 |
| `Intent` | 意图，表达"用户想做什么" |
| `IntentGraph` | 新版意图流，未命中时可自动回退 KG |
| `KG / Knowledge Graph` | 知识问答能力（LLM / Houndify / Together） |
| `Jdocs` | 机器人配置 / 状态文档缓存 |
| `voice.Process` | vector-cloud 中 orchestrating hotword/audio 事件的核心循环 |

---

## 各文档重点速查

| 想看什么 | 看哪份 |
| --- | --- |
| 整体长什么样 | [`architecture.md`](./architecture.md) + 架构交互图 |
| 为什么这样拆 | [`software-design.md`](./software-design.md) §2–§3 |
| 启动时发生了什么 | [`core-flows.md`](./core-flows.md) §1–§2 |
| 一次语音请求怎么跑完 | [`core-flows.md`](./core-flows.md) §3–§7 |
| 具体改哪个文件 | [`secondary-dev-guide.md`](./secondary-dev-guide.md) §2 速查表 |
| 模块内部结构和调用链 | [`program-design.md`](./program-design.md) §3 核心链路 |
| 数据存在哪里 | [`software-design.md`](./software-design.md) §6 数据视图 |
