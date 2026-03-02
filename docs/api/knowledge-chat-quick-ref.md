# /api/v1/knowledge-chat 快速参考卡

## 📋 接口概览

| 属性 | 值 |
|------|-----|
| **路径** | `/api/v1/knowledge-chat/{session_id}` |
| **方法** | `POST` |
| **认证** | Bearer Token 或 X-API-Key |
| **响应** | Server-Sent Events (SSE) 流 |
| **Content-Type** | `application/json` |

---

## 🔑 必填参数

```json
{
  "query": "用户的问题"
}
```

---

## ⚙️ 可选参数速查表

| 参数 | 类型 | 默认值 | 用途 |
|------|------|--------|------|
| `knowledge_base_ids` | `string[]` | 空 | 指定查询的知识库 |
| `knowledge_ids` | `string[]` | 空 | 指定查询的知识文件 |
| `summary_model_id` | `string` | 空 | 覆盖默认模型 |
| `web_search_enabled` | `boolean` | false | 启用网络搜索 |
| `agent_id` | `string` | 空 | 使用自定义Agent |
| `disable_title` | `boolean` | false | 禁用自动标题 |
| `mentioned_items` | `object[]` | 空 | @提及的项目 |

---

## 📤 请求示例汇总

### 基础请求
```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -d '{"query":"你的问题"}'
```

### 完整请求
```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "查询问题",
    "knowledge_base_ids": ["kb-001"],
    "knowledge_ids": [],
    "summary_model_id": "gpt-4",
    "web_search_enabled": false,
    "agent_id": "my-agent",
    "disable_title": false,
    "mentioned_items": []
  }'
```

---

## 📥 响应事件类型

### 1️⃣ references（知识引用）
**发送时机**: 知识检索完成后，第一个返回

```json
{
  "response_type": "references",
  "knowledge_references": [
    {
      "id": "chunk-001",
      "content": "相关内容...",
      "knowledge_id": "kb-001",
      "score": 3.45,
      "knowledge_title": "文件名.pdf",
      "metadata": {}
    }
  ]
}
```

### 2️⃣ answer（LLM回答）
**发送时机**: 流式返回答案（多次）

```json
{
  "response_type": "answer",
  "content": "答案文本片段",
  "done": false
}
```

**完成信号**:
```json
{
  "response_type": "answer",
  "content": "",
  "done": true
}
```

### 3️⃣ error（错误）
```json
{
  "response_type": "error",
  "content": "错误信息"
}
```

---

## 🎯 常见场景

### 场景1: 只用特定知识库
```bash
curl -X POST ... -d '{
  "query": "问题",
  "knowledge_base_ids": ["kb-medicine"]
}'
```

### 场景2: 禁用知识库检索（纯对话）
```bash
curl -X POST ... -d '{
  "query": "问题",
  "knowledge_base_ids": []
}'
```

### 场景3: 使用自定义Agent
```bash
curl -X POST ... -d '{
  "query": "问题",
  "agent_id": "medical-assistant"
}'
```

### 场景4: 多知识库查询
```bash
curl -X POST ... -d '{
  "query": "问题",
  "knowledge_base_ids": ["kb-001", "kb-002", "kb-003"]
}'
```

### 场景5: 查询特定文件
```bash
curl -X POST ... -d '{
  "query": "问题",
  "knowledge_ids": ["know-001"]
}'
```

### 场景6: 启用网络搜索
```bash
curl -X POST ... -d '{
  "query": "问题",
  "web_search_enabled": true
}'
```

---

## 🔄 配置优先级

### 知识库ID
```
请求KB ID > Agent配置 > config.yaml > 系统默认
```

### 聊天模型
```
请求model_id > Agent配置 > KB配置 > 系统默认
```

### 系统提示词
```
Agent配置 > config.yaml > 系统默认
```

### 其他参数（温度、TopK等）
```
请求 > Agent配置 > config.yaml > 系统默认
```

---

## 💾 客户端代码片段

### TypeScript
```typescript
const response = await fetch('/api/v1/knowledge-chat/session-id', {
  method: 'POST',
  headers: {
    'X-API-Key': 'sk-xxxxx',
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ query: '你的问题' }),
});

const reader = response.body.getReader();
while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  const text = new TextDecoder().decode(value);
  // 处理SSE数据...
}
```

### Python
```python
import requests

response = requests.post(
  'http://localhost:8080/api/v1/knowledge-chat/session-id',
  headers={'X-API-Key': 'sk-xxxxx'},
  json={'query': '你的问题'},
  stream=True
)

for line in response.iter_lines():
  if line.startswith(b'data:'):
    data = json.loads(line[5:])
    # 处理事件...
```

### Go
```go
err := apiClient.KnowledgeQAStream(ctx, sessionID, 
  &KnowledgeQARequest{Query: "你的问题"},
  func(response *StreamResponse) error {
    if response.ResponseType == ResponseTypeAnswer {
      fmt.Print(response.Content)
    }
    return nil
  })
```

---

## ⚠️ 常见错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 400 Bad Request | query 为空 | 必须提供 query 参数 |
| 404 Not Found | session 不存在 | 检查 session_id 是否正确 |
| 500 Internal Error | 模型调用失败 | 检查模型配置和网络 |
| SSE 连接断开 | 网络超时 | 增加超时配置 |
| 无知识引用 | 知识库无相关内容 | 降低 vector_threshold 或检查KB |

---

## 🔍 故障排查快速列表

### 收不到引用？
- [ ] 知识库中有相关文档？
- [ ] 向量模型配置正确？
- [ ] vector_threshold 是否过高？
- [ ] knowledge_base_ids 是否正确？

### 答案返回不完整？
- [ ] 监听 done=true 信号？
- [ ] 网络连接稳定？
- [ ] 客户端超时是否足够长？
- [ ] 服务日志有错误？

### Agent 配置未生效？
- [ ] agent_id 是否存在？
- [ ] Agent 配置是否正确保存？
- [ ] 查看日志"Using custom agent"？

### 标题未生成？
- [ ] session.title 初始为空？
- [ ] disable_title = true？
- [ ] 模型是否可用？

---

## 📊 性能建议

- 明确指定 knowledge_base_ids（避免全库检索）
- 问题要清晰具体（避免笼统查询）
- 设置合理的超时（30-60秒）
- 使用知识ID（比KB ID更快）
- 监控模型服务并发限制

---

## 🔗 相关文档

- 📖 [详细指南](./knowledge-chat-detailed.md)
- 🛠️ [实现参考](./knowledge-chat-implementation.md)
- 🧪 [测试示例](./knowledge-chat-testing.md)
- 📚 [API文档](./chat.md)

---

## 💡 提示

💡 **Tip 1**: 同一个 session 可以发送多个查询，系统会自动维护对话历史

💡 **Tip 2**: 使用 `mentioned_items` 为前端UI提供已提及项目的显示信息

💡 **Tip 3**: `knowledge_ids` 的优先级高于 `knowledge_base_ids`

💡 **Tip 4**: 请求参数的优先级永远高于 Agent 配置和全局配置

💡 **Tip 5**: 可以在同一请求中同时启用知识库检索和网络搜索

---

**最后更新**: 2024年
**维护者**: WiseDx Team
