# wire-pod-deepseek 程序设计说明

## 1. 模块总览

```
chipper/
├── cmd/                          # STT入口程序（按STT选择编译不同二进制）
│   ├── vosk/main.go              # Vosk ASR 入口
│   ├── coqui/main.go             # Coqui ASR 入口
│   ├── leopard/main.go           # Leopard ASR 入口
│   └── experimental/whisper.cpp/main.go  # WhisperCPP 入口
│
├── pkg/
│   ├── initwirepod/              # 启动编排层（入口调度，服务注册）
│   │   ├── startserver.go        # StartChipper() / StartFromProgramInit()
│   │   └── web.go                # ChipperHTTPApi() 重启/切换端口接口
│   │
│   ├── vars/                     # 全局配置与路径管理
│   │   ├── config.go             # apiConfig 结构体 + ReadConfig/WriteConfigToDisk
│   │   └── vars.go               # 路径初始化、BotInfo、IntentList、CustomIntents
│   │
│   ├── wirepod/
│   │   ├── preqs/                # 请求处理层：接收 gRPC → 标准化 SpeechRequest
│   │   │   ├── server.go         # New() — STT初始化、插件加载
│   │   │   ├── intent.go         # ProcessIntent() — 传统意图流
│   │   │   ├── intent_graph.go   # ProcessIntentGraph() — 混合意图+KG流
│   │   │   └── knowledgegraph.go # ProcessKnowledgeGraph() — 知识问答流
│   │   │
│   │   ├── ttr/                  # 业务逻辑核心层
│   │   │   ├── matchIntentSend.go  # ProcessTextAll() + customIntentHandler + pluginFunctionHandler
│   │   │   ├── kgsim.go          # KGSim/StreamingKGSim() — LLM/Houndify 调用
│   │   │   ├── intentparam.go    # ParamChecker() / prehistoricParamChecker()
│   │   │   ├── weather.go        # 天气查询
│   │   │   ├── plugins.go        # Go 插件系统注册与执行
│   │   │   ├── bcontrol.go       # 机器人行为控制
│   │   │   └── say_chinese.go    # 中文语音合成辅助
│   │   │
│   │   ├── stt/                  # STT 引擎抽象实现
│   │   │   ├── vosk/Vosk.go
│   │   │   ├── coqui/Coqui.go
│   │   │   ├── leopard/Leopard.go
│   │   │   └── whisper.cpp/WhisperCpp.go
│   │   │
│   │   ├── config-ws/            # 配置页 Web Server
│   │   │   └── webserver.go      # /api/* 路由（天气/KG/STT/自定义意图 CRUD）
│   │   │
│   │   ├── sdkapp/               # SDK 控制页后端
│   │   │   └── server.go         # /api-sdk/* 路由（locale/battery/camera/behavior）
│   │   │
│   │   └── setup/                # 机器人接入流程（证书生成、SSH/BLE 下发）
│   │       ├── certs.go
│   │       ├── ssh.go
│   │       └── ble.go
│   │
│   ├── servers/
│   │   ├── chipper/server.go     # gRPC 服务注册：ChipperGrpc / Jdocs / Token
│   │   ├── jdocs/                # jdocs 协议处理
│   │   └── token/                # token 协议处理
│   │
│   └── scripting/                # Lua 脚本执行环境（自定义意图支持）
│       └── scripting.go
│
└── webroot/                      # 静态资源（HTML/CSS/JS/GIF）
    ├── index.html                # 主页面（脸部动画 + 状态显示）
    ├── setup.html                # 初始设置向导
    ├── css/style.css             # 主样式表
    └── js/main.js                # 主交互逻辑
```

---

## 2. 关键数据结构

### 2.1 `vars.APIConfig`（主配置结构）

位置：`chipper/pkg/vars/config.go:16`

```go
type apiConfig struct {
    Weather struct { Enable bool; Provider string; Key string; Unit string }
    Knowledge struct {
        Enable bool; Provider string; Key string; ID string;
        Model string; IntentGraph bool; RobotName string;
        OpenAIPrompt string; SaveChat bool; CommandsEnable bool;
        Endpoint string; TopP float32; Temperature float32;
    }
    STT struct { Service string; Language string }
    Server struct { EPConfig bool; Port string }
    PastInitialSetup bool
}
```

### 2.2 `vars.CustomIntent`（自定义意图结构）

位置：`chipper/pkg/vars/vars.go`

```go
type CustomIntent struct {
    Name        string
    Description string
    Utterances  []string
    Intent      string
    IsSystemIntent bool
    LuaScript   string
    Exec        string
    ExecArgs    []string
    Params      struct { ParamName string; ParamValue string }
}
```

### 2.3 `JsonIntent`（内置意图结构）

位置：`chipper/pkg/wirepod/preqs/server.go:17`

```go
type JsonIntent struct {
    Name              string   `json:"name"`
    Keyphrases        []string `json:"keyphrases"`
    RequireExactMatch bool     `json:"requiresexact"`
}
```

---

## 3. 核心调用链（从触发到响应）

```
[1] vector-cloud/internal/voice/process.go
    └─→ hotword event received → modeMap lookup → create Streamer
    └─→ chipper.NewIntentGraphStream(...)  [或 NewIntentStream / NewKGStream]

[2] chipper/pkg/servers/chipper/intent_graph.go
    └─→ chipperserver.ProcessIntentGraph()
    └─→ preqs.ProcessIntentGraph(req)

[3] chipper/pkg/wirepod/preqs/intent_graph.go
    └─→ sttHandler(speechReq) → transcribedText
    └─→ ttr.ProcessTextAll(req, transcribedText, intents, isOpus)

[4] chipper/pkg/wirepod/ttr/matchIntentSend.go
    ├─→ pluginFunctionHandler()        // Go 插件匹配（优先级最高）
    ├─→ customIntentHandler()          // JSON 自定义意图匹配
    ├─→ 精确匹配 intent-data/*.json
    └─→ 子串匹配（RequireExactMatch=false 时降级）
    └─→ 返回 false → 上层决定是否走 KG

[5] chipper/pkg/wirepod/preqs/intent_graph.go（未命中分支）
    └─→ ttr.StreamingKGSim() or ttr.KGSim()
    └─→ openai client → streaming response → TTS → IntentPass回写

[6] 最终结果通过 gRPC stream 回到 vector-cloud → 机器人消费
```

---

## 4. 各模块设计职责说明

### 4.1 `initwirepod`（启动编排）

- **职责**：程序的第一个入口，负责初始化所有子系统并按顺序启动
- **关键函数**：
  - `StartFromProgramInit(sttInitFunc, sttHandlerFunc, voiceProcessorName)`
    - 检查配置是否完成，未完成则提示用户打开网页
    - 完成后启动 TLS gRPC 服务和配置页 Web 服务
  - `BeginWirepodSpecific(...)`
    - 初始化 logger、vars，创建 VoiceProcessor
    - 注册 `/api-chipper/` HTTP 路由

### 4.2 `servers/chipper`（gRPC 服务注册）

- **职责**：将三个 gRPC 服务（ChipperGrpc / Jdocs / Token）挂到同一个 grpc.Server
- **使用 cmux**：同一个 TLS 端口同时处理 gRPC（HTTP/2）和普通 HTTP（健康检查）
- **关键代码**：`startserver.go:grpcServe()` 第 56–81 行

### 4.3 `preqs`（请求预处理）

- **职责**：把 gRPC 请求统一转换为内部模型 `SpeechRequest`，调用 STT 后分流给 ttr
- **三种模式分别对应三条代码路径**：
  - `ProcessIntent()` → 只走 intent 匹配
  - `ProcessIntentGraph()` → intent 未命中时自动回退到 StreamingKGSim
  - `ProcessKnowledgeGraph()` → 直接进入 KG 问答路径
- **STT 切换机制**：通过 `ReloadVosk()` 在运行时重新加载语言模型（修改 STT 语言时调用）

### 4.4 `ttr`（业务大脑）

- **优先级**（从高到低）：
  1. Go 插件（`pluginFunctionHandler`）
  2. 自定义意图（`customIntentHandler`，支持 Lua + exec）
  3. 内置意图精确匹配（`Keyphrases` == transcribedText）
  4. 内置意图模糊匹配（`Contains`，`RequireExactMatch=false`）
  5. 上层 fallback：KG/LLM 或 unmatched
- **参数提取**：`intentparam.go` 中的 `ParamChecker` / `prehistoricParamChecker`
- **天气处理**：`weather.go`，调用外部天气 API 并格式化返回文本
- **KG 回退**：`kgsim.go` 的 `StreamingKGSim` 是长函数（~600行），包含：
  - OpenAI compatible streaming chat completion
  - 聊天记忆（每个 ESN 保留最近 16 轮）
  - TTS 输出控制（SayText）
  - 中断处理（`kgsim_interrupt.go`）

### 4.5 `config-ws` 和 `sdkapp`

- **config-ws**：配置管理专用，所有 `/api/*` 路由都落在这里
  - `/api/set_weather_api` / `/api/set_kg_api` / `/api/set_stt_info`
  - `/api/add_custom_intent` / `/api/edit_custom_intent` / `/api/remove_custom_intent`
- **sdkapp**：机器人实时控制专用，所有 `/api-sdk/*` 路由
  - locale 切换、button_wakeword 读取、电池/脸部/摄像头查询

---

## 5. vector-cloud 程序结构（补充说明）

```
vector-cloud/
├── gateway/main.go          # vic-gateway：机器人侧 gRPC 网关
│   └─→ RegisterExternalInterfaceServer(grpcServer, newServer())
│   └─→ JSON gateway via grpc-gateway
│
├── cloud/main.go            # vic-cloud：语音流程编排主入口
│   └─→ getSocketWithRetry(mic_sock/ai_sock)
│   └─→ voice.Process{ AddReceiver, AddIntentWriter, Run }
│
└── internal/
    ├── voice/process.go     # Hotword/Audio/AudioDone 事件循环
    │   └─→ modeMap[mode] → StreamOpts → newStream()
    ├── ipc/                 # Unixgram/Unixsocket IPC 基础设施
    ├── token/               # Token 刷新与管理
    ├── jdocs/               # jdocs 文档读写代理
    └── cloudproc/           # 进程协调器（聚合 voice/token/jdocs）
```
