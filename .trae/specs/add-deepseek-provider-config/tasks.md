# Tasks
- [x] Task 1: 梳理并补齐 DeepSeek 配置数据结构
  - [x] SubTask 1.1: 为知识问答配置补充 `deepseek` provider 的保存约定与默认模型策略
  - [x] SubTask 1.2: 明确 `enable`、`provider`、`key`、`model` 与现有高级开关在 DeepSeek 场景下的兼容关系
  - [x] SubTask 1.3: 确认旧配置文件在未包含 DeepSeek 字段时仍可正常读取

- [x] Task 2: 更新设置页面，增加 DeepSeek 独立选型与表单
  - [x] SubTask 2.1: 在 Knowledge Graph provider 下拉框中新增 `DeepSeek`
  - [x] SubTask 2.2: 为 DeepSeek 增加专属表单区，至少包含启用状态、API Key、模型选择/填写、Prompt、Intent Graph、Save Chat、Commands Enable
  - [x] SubTask 2.3: 调整前端显隐逻辑、保存逻辑、回填逻辑，确保 DeepSeek 不再复用 Ollama 的 custom endpoint 输入

- [x] Task 3: 更新后端知识问答调用链，按官方文档接入 DeepSeek
  - [x] SubTask 3.1: 在 LLM 客户端初始化逻辑中新增 `provider=deepseek` 分支
  - [x] SubTask 3.2: 将 DeepSeek `base_url` 固定为 `https://api.deepseek.com`
  - [x] SubTask 3.3: 将默认模型设为 `deepseek-v4-flash`，并允许配置 `deepseek-v4-pro`
  - [x] SubTask 3.4: 确保启用 DeepSeek 后知识问答请求不再走本地 Ollama endpoint

- [x] Task 4: 做好兼容与错误处理
  - [x] SubTask 4.1: 保持 `custom` provider 对 Ollama 和其他 OpenAI 兼容接口的现有支持
  - [x] SubTask 4.2: 为 DeepSeek 缺失 API Key、模型为空、接口调用失败等情况提供清晰日志或提示
  - [x] SubTask 4.3: 确认 DeepSeek 接入不会误影响 OpenAI 专属 TTS 配置分支

- [x] Task 5: 完成验证与回归
  - [x] SubTask 5.1: 验证 `/api/set_kg_api` 与 `/api/get_kg_api` 在 DeepSeek 场景下的保存与回填
  - [x] SubTask 5.2: 验证设置页切换 `DeepSeek`、`Custom`、`None` 时表单行为正确
  - [x] SubTask 5.3: 验证知识问答请求在 DeepSeek 启用后实际命中官方 `chat/completions` 路径
  - [x] SubTask 5.4: 运行项目已有相关测试或最小可行手工验证流程

- [x] Task 6: 补齐 DeepSeek 最小回归证据
  - [x] SubTask 6.1: 增加可重复执行的 DeepSeek 配置与客户端回归测试
  - [x] SubTask 6.2: 运行回归测试并保存通过依据
  - [x] SubTask 6.3: 用回归结果重新核对 checklist 第 10 项

- [ ] Task 7: 修复中文问答链路的 DeepSeek 实际可用性
  - [ ] SubTask 7.1: 让流式分句支持中文 `。！？`，避免 DeepSeek 中文回复被误判为无响应
  - [ ] SubTask 7.2: 降低中文短句被本地 intent 误匹配的概率，让问答更稳定进入 DeepSeek
  - [ ] SubTask 7.3: 为上述修复补充回归测试并验证真实日志行为

# Task Dependencies
- Task 2 依赖 Task 1
- Task 3 依赖 Task 1
- Task 4 依赖 Task 3
- Task 5 依赖 Task 2、Task 3、Task 4
- Task 6 依赖 Task 2、Task 3
- Task 7 依赖 Task 3、Task 5
