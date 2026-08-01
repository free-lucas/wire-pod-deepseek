# wire-pod 二开文档索引

这套文档的目标是把 `wire-pod` 从“代码很多、看起来很乱”梳理成一张可落地的软件设计地图，方便你后面做二次开发。

这次文档整理是结合源码和 `codegraph` 索引一起核对出来的，重点不是只讲目录，而是尽量讲清楚：

1. 系统由哪些进程和模块组成
2. 它们之间怎么通信
3. 启动时做了什么
4. 一次语音请求怎么从机器人流到最终响应
5. 配置、jdocs、自定义意图这些数据落在哪里
6. 二开时优先改哪里、哪些地方改动成本高

## 建议阅读顺序

1. [`project-overview.md`](./project-overview.md)
   先建立仓库级别的全局认识。
2. [`software-design.md`](./software-design.md)
   从“软件设计文档”的视角看系统上下文、设计目标、运行视图和数据视图。
3. [`architecture.md`](./architecture.md)
   看组件拆分、职责边界、进程关系和部署关系。
4. [`core-flows.md`](./core-flows.md)
   看启动流程、语音流程、配置流程、机器人接入流程。
5. [`secondary-dev-guide.md`](./secondary-dev-guide.md)
   进入二开落地阶段，按需求反推代码入口。

## 这套系统先记住 3 件事

- `wire-pod` 不是单体应用，而是一个“主机侧服务 + 机器人侧兼容层”的组合：
  - `chipper/`: 主机侧核心服务，负责 STT、意图匹配、LLM/KG、配置页、SDK 页面、自定义意图、插件等。
  - `vector-cloud/`: 机器人侧云兼容层，负责把机器人原本面向官方云的协议转接到本地 `chipper`。
- 日常二开里，**大部分业务改动集中在 `chipper/`**，尤其是 `preqs/`、`ttr/`、`config-ws/`、`sdkapp/`。
- `vector-cloud/` 虽然平时不一定要改，但它决定了机器人如何把音频、意图、jdocs、token 流转到主机侧，所以做深度改造时绕不过去。

## 新手词汇表

- `main.go`: 程序入口文件，可以理解为“从这里启动”。
- `package`: Go 的代码分组方式，可以理解为“一个模块”。
- `handler`: 处理请求的函数。
- `gRPC`: 一种服务通信协议，这个项目里机器人和服务端会大量使用。
- `STT`: Speech To Text，语音转文字。
- `Intent`: 意图，机器人识别出“你想做什么”。
- `IntentGraph`: 新一点的语音意图链路，既可以返回普通 intent，也可以直接返回知识回答。
- `KG / Knowledge Graph`: 这里可以理解成问答能力，当前代码里既兼容 Houndify，也兼容 LLM。
- `Jdocs`: 机器人配置/状态文档。

## 你后面最常用的文档

- 想看整体设计：[`software-design.md`](./software-design.md)
- 想确认模块边界：[`architecture.md`](./architecture.md)
- 想跟一条请求跑代码：[`core-flows.md`](./core-flows.md)
- 想直接开始改：[`secondary-dev-guide.md`](./secondary-dev-guide.md)
