# 项目总览

## 1. 项目是做什么的

`wire-pod` 是给 Anki / DDL 的 Vector 机器人提供本地化、自建化服务的一套软件。

它解决的核心问题是：

- 让机器人不依赖官方云也能工作
- 接管语音请求、知识问答、配置管理、机器人连接
- 提供网页配置界面和 SDK 控制页面
- 支持自定义意图、插件、Lua 脚本等扩展能力

你可以把它先理解成：

> “一个给 Vector 机器人准备的本地后端平台”

## 2. 仓库的两大主体

### `chipper/`

这是你二开时最常接触的部分，负责：

- 启动 wire-pod 主服务
- 接收机器人发来的语音/文本请求
- 做 STT、意图匹配、知识问答
- 提供网页配置页
- 提供 SDK 控制页
- 管理自定义意图、插件、配置文件、证书

### `vector-cloud/`

这一部分更接近“机器人端兼容层”：

- 模拟/承接原始云服务协议
- 负责机器人和服务端之间的消息流
- 管理 token、jdocs、音频流、网关

如果你后续想做：

- 新增一个网页配置项
- 新增一个自定义命令
- 修改语音匹配逻辑
- 对接新的 STT 或 LLM

优先看 `chipper/` 就够了。

## 3. 仓库目录地图

```text
wire-pod/
├── chipper/                  主机侧核心服务
│   ├── cmd/                  不同 STT 入口程序
│   ├── pkg/initwirepod/      启动和服务编排
│   ├── pkg/wirepod/preqs/    请求进入后的第一层处理
│   ├── pkg/wirepod/ttr/      文本转响应、意图匹配、插件、LLM
│   ├── pkg/wirepod/stt/      各类语音识别实现
│   ├── pkg/wirepod/config-ws/ 配置网页后端
│   ├── pkg/wirepod/sdkapp/   SDK 控制页后端
│   ├── pkg/wirepod/setup/    机器人接入、证书、SSH/BLE
│   ├── pkg/vars/             全局配置和路径
│   ├── intent-data/          多语言意图词库
│   ├── webroot/              配置页和 SDK 页面静态资源
│   └── plugins/              Go 插件示例
├── vector-cloud/             机器人侧兼容模块
├── compose.yaml              Docker 启动方式
├── dockerfile                容器镜像构建
├── setup.sh                  初始化安装脚本
└── update.sh                 更新脚本（谨慎使用）
```

## 4. 程序怎么启动

正常理解路径如下：

1. 顶层 `chipper/start.sh` 根据 `source.sh` 里的 `STT_SERVICE` 选择入口
2. 例如默认常见场景会走 `chipper/cmd/vosk/main.go`
3. `main.go` 调用 `pkg/initwirepod.StartFromProgramInit(...)`
4. `initwirepod` 会继续启动：
   - 语音处理服务
   - gRPC / HTTP 服务
   - 配置页 Web 服务
   - SDK 控制页服务

所以从二开视角看，最重要的主入口是：

- `chipper/start.sh`
- `chipper/cmd/*/main.go`
- `chipper/pkg/initwirepod/startserver.go`

## 5. 这个项目里的“核心能力”

按功能分，可以拆成 6 类：

1. 机器人接入与启动
2. 语音识别与意图匹配
3. 问答/知识能力
4. 网页配置后台
5. SDK 控制和机器人远程操作
6. 自定义扩展能力

## 6. 新手先抓住的关键文件

如果你暂时不想被大量 Go 文件淹没，建议先只看这些：

- `chipper/pkg/initwirepod/startserver.go`
- `chipper/pkg/wirepod/preqs/server.go`
- `chipper/pkg/wirepod/preqs/intent.go`
- `chipper/pkg/wirepod/ttr/matchIntentSend.go`
- `chipper/pkg/wirepod/config-ws/webserver.go`
- `chipper/pkg/wirepod/sdkapp/server.go`
- `chipper/pkg/vars/vars.go`
- `chipper/pkg/vars/config.go`

## 7. 哪些地方最适合二次开发

### 低门槛入口

- 改网页配置逻辑：`pkg/wirepod/config-ws/`
- 改 SDK 控制能力：`pkg/wirepod/sdkapp/`
- 增加自定义意图：配置接口 + `customIntents.json`
- 改意图匹配逻辑：`pkg/wirepod/ttr/`

### 中等难度入口

- 新增 STT 实现：`pkg/wirepod/stt/`
- 改知识问答/LLM 行为：`pkg/wirepod/ttr/kgsim.go`
- 改多语言意图库：`intent-data/*.json`

### 高难度入口

- 改机器人云兼容流程：`vector-cloud/`
- 改设备接入/证书/机器人安装流程：`pkg/wirepod/setup/`

## 8. 你后续二开时的建议心法

- 先改 `chipper/`，后碰 `vector-cloud/`
- 先改“配置页/后端接口”，再改“底层语音链路”
- 先通过已有扩展点实现需求，再考虑重构底层
- 先读流程，再动代码

下一份建议继续看 [`architecture.md`](./architecture.md)。
