# Contract: Wakeword Surface

## 目的

描述本次需求中会触达的用户表面与配置接口，明确哪些是“页面/配置层契约”，哪些不等同于真实热词引擎能力。

## 边界锁定

以下内容属于本次实现必须明确遵守的契约边界：

1. 页面与接口层可以承诺“默认展示与默认推荐入口迁移到 小智，小智”
2. 页面与接口层不能承诺“机器人一定已经把 小智，小智 识别成原生热词”
3. 若底层语言映射或固件热词能力不足，接口和文档必须显式暴露限制，而不是静默回退

## 1. STT 语言设置接口

**Path**: `/api/set_stt_info`

**Request**:

```json
{
  "language": "zh-CN"
}
```

**Behavior Contract**:

- 当 STT 服务为 `vosk` 或 `whisper.cpp` 时，接口接受合法语言码
- 如果中文模型未下载，允许先进入下载流程
- 成功后应把 `apiConfig.json` 中的 `STT.language` 更新为 `zh-CN`

## 2. SDK Locale 设置接口

**Path**: `/api-sdk/locale?serial={esn}&locale={locale}`

**Behavior Contract**:

- 接口负责将 `locale` 通过机器人 `/v1/update_settings` 下发
- 刷新 SDK 设置后，应能从 `get_sdk_settings` 读取到更新后的 `locale`
- 本接口属于机器人设置层，不直接等同于 STT 主配置

## 3. SDK 设置读取接口

**Path**: `/api-sdk/get_sdk_settings?serial={esn}`

**Behavior Contract**:

- 接口应从机器人拉取 `ROBOT_SETTINGS`
- 返回结果中可包含 `locale`、`button_wakeword` 等字段
- 返回值会被缓存到 `jdocs/jdocs.json`

## 4. 远程触发接口

**Path**: `/api-sdk/trigger_wake_word?serial={esn}`

**Behavior Contract**:

- 本接口用于模拟一次按钮触发
- 它只能证明“远程触发流程存在”，不能证明语音热词文本已经被替换成“小智，小智”
- 旧接口路径可以因兼容性继续保留，但不得继续作为默认推荐入口

## 5. 页面展示契约

**涉及页面**:

- `chipper/webroot/sdkapp/settings.html`
- `chipper/webroot/sdkapp/js/main.js`
- `chipper/webroot/sdkapp/js/heyvector.js`

**Behavior Contract**:

- 页面中默认展示的唤醒文案应与本次需求一致，即“小智，小智”
- 页面中的旧词 `Hey Vector` 不应继续作为默认推荐入口
- 若旧路径 `/api-sdk/button_hey_vector` 或 `/api-sdk/trigger_wake_word` 继续保留，它们只表示兼容接口，不代表默认入口仍是 `Hey Vector`
- 页面回显必须与接口读取到的当前设置一致

## 6. 非契约范围

以下内容不应在本契约中被误解为“仓库已直接控制”：

1. 机器人真实热词文本识别逻辑
2. 固件侧是否支持 `"小智，小智"` 作为原生热词
3. 上游热词检测模型是否可替换

这些内容需要在实现或验证阶段另行确认，不能仅凭页面/API 改动直接推断成功。
