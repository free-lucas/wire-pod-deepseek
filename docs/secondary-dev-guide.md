# 二次开发指南（v2 — 基于源码核对版）

> 本文档与当前源码（commit `5b35c8ee`）逐层核对，重点标记**真正需要改代码的位置**。

---

## 1. 推荐的二开路径

按难度从低到高：

1. **改配置页和接口** → `chipper/pkg/wirepod/config-ws/webserver.go` + `webroot/`
2. **改自定义意图** → `customIntents.json` 或 `pkg/wirepod/ttr/matchIntentSend.go`
3. **改内置意图匹配逻辑** → `pkg/wirepod/ttr/matchIntentSend.go` 的 `ProcessTextAll()`
4. **改知识问答 / LLM** → `pkg/wirepod/ttr/kgsim.go` + `pkg/vars/config.go`
5. **接入新 STT** → `pkg/wirepod/stt/` + `cmd/` + `start.sh`
6. **改 `vector-cloud`** → 仅机器人侧协议适配需要时再碰

---

## 2. 常见需求 → 精确落点

| 需求 | 优先改动位置 | 关键函数 / 行号 |
| --- | --- | --- |
| 增加一个网页设置项 | `config-ws/webserver.go` + `webroot/setup.html` | `set_weather_api` / `set_kg_api` 已有模式可抄 |
| 增加一个机器人控制按钮 | `sdkapp/server.go` + `webroot/sdkapp/settings.html` | `/api-sdk/locale` / `/api-sdk/get_sdk_settings` |
| 新增"说某句话就触发"功能 | `customIntents.json` 或 `ttr/matchIntentSend.go` | `customIntentHandler()` @ line 100 |
| 修改现有意图匹配方式 | `ttr/matchIntentSend.go` | `ProcessTextAll()` @ line 241 |
| 接入新的 LLM 提供商 | `ttr/kgsim.go` + `vars/config.go` | `StreamingKGSim()` @ line 189；Provider switch @ line 242 |
| 接入新的 STT 引擎 | `stt/` + `cmd/` + `start.sh` | `sttHandler func(sr.SpeechRequest) (string, error)` |
| 改机器人安装 / 配网 / 证书 | `setup/certs.go` / `setup/ssh.go` / `setup/ble.go` | `CreateCertCombo()` / `CreateServerConfig()` |
| 改机器人侧协议兼容 | `vector-cloud/gateway/` / `vector-cloud/internal/voice/` | `gateway/main.go:226` ExternalInterface 注册 |

---

## 3. 最容易下手的 4 个方向

### 方向 A：先改网页配置

为什么适合新手：看到效果快、风险小、不接触音频流。

完整步骤：
1. 在 `chipper/webroot/` 加 HTML 字段
2. 在 `pkg/wirepod/config-ws/webserver.go` 加 `POST /api/xxx` handler
3. 读取请求体写入 `vars.APIConfig.XxxField`
4. 调用 `vars.WriteConfigToDisk()`
5. （可选）若涉及 STT 语言变化，调用 `preqs.ReloadVosk()`

### 方向 B：先改自定义意图

自定义意图支持三种执行模式：
- **纯 intent 返回**：直接返回一个 intent name，走正常 intent 处理链
- **系统意图（IsSystemIntent=true）**：返回 JSON `{"status":"ok","returnIntent":"intent_xxx"}`，解析后继续走 IntentPass
- **Lua 脚本**：`LuaScript` 字段指向 `.lua` 文件，通过 `scripting.RunLuaScript()` 异步执行
- **外部命令**：`Exec` + `ExecArgs`，支持 `!botSerial`、`!speechText`、`!intentName`、`!locale` 四个模板变量

重点看：
- `vars.CustomIntents` 结构（`vars/vars.go`）
- `matchIntentSend.go:100` 的 `customIntentHandler()`

### 方向 C：改内置 intent 匹配

优先级规则（`matchIntentSend.go:241` `ProcessTextAll`）：
1. Go 插件（`pluginFunctionHandler`）
2. 自定义意图（`customIntentHandler`）
3. 精确匹配 `intent-data/*.json`（== 全等）
4. 模糊匹配（Contains，`RequireExactMatch=false`）

修改意图词库直接在 `chipper/intent-data/<lang>.json` 编辑，无需重新编译。

### 方向 D：接新 LLM

当前支持的 Provider（`kgsim.go:242`）：
- `openai` — 直接调 OpenAI API
- `together` — 调 Together AI（默认模型 Llama-3-70b）
- `custom` — 任意 OpenAI-compatible endpoint（填 Endpoint）
- `houndify` — Houndify 语音搜索（在 `preqs/knowledgegraph.go` 单独处理）

新增 Provider 需要在 `kgsim.go` 的 Provider switch 中加分支，并在 `vars/config.go` 确认结构体覆盖该值。

---

## 4. 最常改到的文件类型

### 入口文件
- `chipper/cmd/*/main.go` — 一般不改，除非换 STT 入口

### 配置结构体
- `chipper/pkg/vars/config.go` — 新增任何配置项都从这里开始
- `chipper/pkg/vars/vars.go` — 全局状态变量、路径定义、BotInfo 加载

### 请求处理
- `chipper/pkg/wirepod/preqs/intent.go` — `ProcessIntent()`
- `chipper/pkg/wirepod/preqs/intent_graph.go` — `ProcessIntentGraph()`
- `chipper/pkg/wirepod/preqs/knowledgegraph.go` — `ProcessKnowledgeGraph()`

### 核心业务
- `chipper/pkg/wirepod/ttr/matchIntentSend.go` — 意图匹配总调度（最重要的一的文件之一）
- `chipper/pkg/wirepod/ttr/kgsim.go` — KG/LLM 调用（646行的核心长函数）
- `chipper/pkg/wirepod/ttr/weather.go` — 天气查询

### 前端静态页
- `chipper/webroot/setup.html` — 初始设置向导
- `chipper/webroot/index.html` — 主页面（脸部动画 + 状态）
- `chipper/webroot/sdkapp/settings.html` — SDK 控制面板
- `chipper/webroot/js/main.js` — 主页面交互

---

## 5. 值得优先读的源代码顺序

### 第 1 轮（建立全局地图，必读）
```
chipper/pkg/initwirepod/startserver.go   ← 启动调度中枢
chipper/pkg/wirepod/preqs/server.go      ← 请求入口构造函数
chipper/pkg/wirepod/ttr/matchIntentSend.go ← 意图匹配总调度
chipper/pkg/vars/config.go               ← 配置结构
chipper/pkg/vars/vars.go                 ← 全局变量和路径
```

### 第 2 轮（按需读）
- 想改网页 → `config-ws/webserver.go` + `sdkapp/server.go`
- 想做问答 → `ttr/kgsim.go`（只看 `StreamingKGSim()` 的 Provider 分支部分即可）
- 想改安装流程 → `setup/certs.go` + `setup/ssh.go`
- 想理解向量云通信 → `vector-cloud/internal/voice/process.go:170` hotword 事件循环

---

## 6. 新手最容易踩的坑

### 坑 1：以为只有一个 HTTP 服务
实际上有三个并发监听：
- gRPC/TLS（芯片侧，端口由 `apiConfig.json` 的 `server.port` 决定）
- 配置页 Web（`vars.WebPort`，默认 `:8080`）
- SDK 控制页（`:80`）

它们共享同一个 `http.DefaultServeMux`，路由前缀不同所以不冲突。

### 坑 2：以为所有逻辑都在 `vector-cloud`
日常二开大部分逻辑在 `chipper/`。`vector-cloud` 只负责机器人侧的协议桥接，改动它的门槛明显更高。

### 坑 3：只改 `intent-data/*.json` 以为就够了
命中 intent 之后还要经过：参数提取（`ParamChecker`）→ 天气/KG 分支判断 → TTS 输出 → 回包。简单改 intent name 不会改变行为逻辑，要同时理解后续链路。

### 坑 4：乱跑 `update.sh`
这个脚本包含 `git reset --hard origin/main`，会强制覆盖本地所有改动。二开期间不要运行它。

### 坑 5：一开始就改 `vector-cloud`
机器人侧涉及 IPC、token 刷新、AudioDone 状态机、Unixgram socket 等底层机制，建议 chipper 完全熟悉后再考虑。

### 坑 6：不知道 IntentGraph 和 Intent 的区别
- `Intent` 只返回普通意图动作，失败时上层决定是否走 LLM
- `IntentGraph` 在 intent 未命中时可以**直接在 chipper 侧返回 KG 回答**，不再依赖外层 fallback
- 新版 Vector 固件（≥1.8.x / 2.0.x）默认使用 `IntentGraph` 流

---

## 7. 推荐的第一批二开任务

做完这 3 个任务，你对系统的核心链路就建立起来了：

1. **在 `apiConfig.json` 加一个新字段，并通过网页保存它**
   - 目标：理解 `config-ws` → `vars` → `WriteConfigToDisk()` 整条写盘链路
2. **在 `customIntents.json` 加一条"说出这句话就执行 `echo hello`"的规则**
   - 目标：理解 `customIntentHandler` 的执行流程和 `exec.Command` 的参数展开
3. **在 `sdkapp` 里加一个新按钮，点击后让机器人说一句话**
   - 目标：理解 `/api-sdk/*` → `vector-go-sdk` → 机器人的调用链

---

## 8. 定位问题时的思考框架

看到一个 bug 或不符合预期的行为，依次问自己：

| 问题 | 排查方向 |
| --- | --- |
| 是网页没显示正确，还是机器人声音不对？ | 先区分是配置问题还是语音链路问题 |
| 是匹配不到 intent，还是匹配到了但响应错了？ | 检查日志，看是否走到 `ttr.ProcessTextAll` |
| 是主机的问题还是机器人端的问题？ | 检查 chipper 日志 vs vector-cloud 日志 |
| 改配置能解决还是需要改代码？ | 优先尝试配置文件，不行再看源码 |

---

## 9. 正式开始改代码前的 checklist

```
□ 确认需求属于哪一层（页面 / 意图 / KG / 机器人协议）
□ 找到对应的入口文件（见 §2 表格）
□ 看该 handler 最终落到哪个函数
□ 定位相关的配置结构体或数据模型
□ 做最小改动（一次只动一处）
□ 确认是否影响网页展示 / 机器人行为 / 配置文件落盘
□ 再继续扩展
```

---

## 10. 一句话记忆

对这个项目做二开，最重要的不是先学会所有 Go 语法，而是建立下面这张稳定地图：

> **启动在 `initwirepod` → 请求先到 `preqs` → 业务主要在 `ttr` → 配置在 `vars` / `config-ws` → 页面在 `webroot` → 机器人协议在 `vector-cloud`**

---

## 11. 与"小智，小智"改造相关注意事项

本次改造涉及默认唤醒词从 "Hey, Vector" 切到中文场景，最容易踩的点是：**页面文案改了 ≠ 机器人真实热词已替换**。

改造落点优先级：
1. `chipper/webroot/setup.html` 和 `chipper/webroot/sdkapp/settings.html` — 页面默认体验
2. `chipper/pkg/wirepod/config-ws/webserver.go` — STT 语言写入（`zh-CN`）
3. `chipper/pkg/wirepod/sdkapp/` — locale / button_wakeword / `vic.RobotSettings` 缓存
4. `vector-cloud/internal/voice/process.go` — hotword 事件后的 `LanguageCode` 映射

注意：当前 `getLanguage()` 函数（`process.go:456`）的 switch 只有 en/fr/de 三个分支，`zh-CN` 会回退到英文。因此中文语音热词的真实识别能力取决于底层 STT 模型和固件热词配置，不在 `chipper` 代码本身内。
