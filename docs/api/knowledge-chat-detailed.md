# /api/v1/knowledge-chat/ 接口详细指南

## 快速开始

### 最简单的请求

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/your-session-id' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{"query": "你的问题"}'
```

### 完整参数请求

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "查询内容",
    "knowledge_base_ids": ["kb-001"],
    "knowledge_ids": [],
    "summary_model_id": "gpt-4",
    "web_search_enabled": false,
    "agent_id": "agent-001",
    "disable_title": false,
    "mentioned_items": []
  }'
```

---

## 请求参数详解

### query（必填）
- **类型**: `string`
- **约束**: 不能为空
- **说明**: 用户提出的问题或查询
- **示例**: "什么是机器学习？"

### knowledge_base_ids（可选）
- **类型**: `string[]`
- **默认值**: 空数组
- **说明**: 指定要查询的知识库ID列表。如果不指定，将根据自定义Agent配置决定
- **示例**: `["kb-medicine-001", "kb-medicine-002"]`
- **优先级**: 请求指定 > Agent配置 > 全局配置

### knowledge_ids（可选）
- **类型**: `string[]`
- **默认值**: 空数组
- **说明**: 指定要查询的具体知识文件ID列表。优先级高于knowledge_base_ids
- **示例**: `["know-001", "know-002"]`
- **场景**: 当用户在界面上@提及特定文件时使用

### summary_model_id（可选）
- **类型**: `string`
- **默认值**: 空
- **说明**: 覆盖默认的摘要/聊天模型ID
- **示例**: `"gpt-4"`, `"claude-3"`
- **优先级**: 最高 - 请求参数 > Agent配置 > 知识库配置 > 系统默认

### web_search_enabled（可选）
- **类型**: `boolean`
- **默认值**: `false`
- **说明**: 是否在知识库检索基础上启用网络搜索
- **示例**: `true`
- **注意**: 启用时会使用rag_stream管道，支持融合知识库和网络搜索结果

### agent_id（可选）
- **类型**: `string`
- **默认值**: 空
- **说明**: 绑定的自定义Agent ID，用于应用Agent特定的配置
- **示例**: `"medical-assistant-v1"`
- **效果**: 
  - 从自定义Agent加载系统提示词
  - 应用Agent的模型、温度、历史轮数等配置
  - 遵循Agent的知识库检索策略

### disable_title（可选）
- **类型**: `boolean`
- **默认值**: `false`
- **说明**: 是否禁用会话标题的自动生成
- **示例**: `true`
- **场景**: 当不需要自动生成标题时设置为true

### mentioned_items（可选）
- **类型**: `object[]`
- **默认值**: 空数组
- **说明**: 用户@提及的知识库和文件列表
- **示例**:
```json
[
  {
    "id": "kb-001",
    "name": "医学知识库",
    "type": "kb",
    "kb_type": "document"
  },
  {
    "id": "know-001",
    "name": "心脏病指南",
    "type": "knowledge",
    "kb_type": "document"
  }
]
```
- **字段说明**:
  - `id`: 知识库或知识的ID
  - `name`: 显示名称
  - `type`: "kb" 或 "knowledge"
  - `kb_type`: 知识库类型（"document", "faq"等）

---

## 响应格式详解

### 响应类型：references（知识引用）

首先返回检索到的相关文档块。

```json
{
  "id": "msg-12345",
  "response_type": "references",
  "content": "",
  "done": false,
  "knowledge_references": [
    {
      "id": "chunk-001",
      "content": "机器学习是人工智能的一个分支，它使计算机能够从数据中学习...",
      "knowledge_id": "kb-ai-001",
      "chunk_index": 0,
      "knowledge_title": "AI基础.pdf",
      "start_at": 0,
      "end_at": 150,
      "seq": 0,
      "score": 3.456,
      "match_type": 3,
      "sub_chunk_id": ["sub-001", "sub-002"],
      "metadata": {
        "page": 1,
        "section": "概述"
      },
      "chunk_type": "text",
      "parent_chunk_id": "",
      "image_info": "",
      "knowledge_filename": "AI基础.pdf",
      "knowledge_source": "upload"
    }
  ]
}
```

**knowledge_references数组中每个对象的字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 文档块的唯一ID |
| `content` | string | 文档块的完整内容 |
| `knowledge_id` | string | 所属知识库的ID |
| `chunk_index` | number | 块在文档中的索引 |
| `knowledge_title` | string | 原始知识文件的标题 |
| `score` | number | 相关性分数（越高越相关） |
| `match_type` | number | 匹配类型：1=关键词, 2=向量, 3=混合 |
| `metadata` | object | 元数据信息（如页码、章节等） |
| `chunk_type` | string | 块类型：text/summary/table/image等 |
| `knowledge_filename` | string | 原始文件名 |
| `knowledge_source` | string | 知识来源：upload/url/manual等 |

### 响应类型：answer（LLM回答）

然后流式返回LLM生成的答案。

```json
{
  "id": "msg-12345",
  "response_type": "answer",
  "content": "机器学习",
  "done": false,
  "knowledge_references": null
}
```

**多个answer事件会连续发送**，客户端需要拼接所有的content字段：

```
答案第1部分 → content: "机器学习"
答案第2部分 → content: "是"
答案第3部分 → content: "人工智能"
...
答案最后 → content: "", done: true (标志完成)
```

**最终拼接结果**: "机器学习是人工智能..."

---

## 完整请求-响应周期示例

### 场景：查询医学知识库

**请求**:
```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-abc123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "什么是高血压？",
    "knowledge_base_ids": ["kb-medicine"],
    "agent_id": "medical-agent"
  }'
```

**响应流程**:

```
# 1. 第一个事件：返回知识库引用（单次）
event: message
data: {"id":"msg-123","response_type":"references","done":false,"knowledge_references":[
  {"id":"chunk-001","content":"高血压是指血压长期...","knowledge_id":"kb-medicine","score":4.2,...},
  {"id":"chunk-002","content":"高血压的危害包括...","knowledge_id":"kb-medicine","score":3.8,...}
]}

# 2-N. 流式答案事件（多次）
event: message
data: {"id":"msg-123","response_type":"answer","content":"高血压","done":false}

event: message
data: {"id":"msg-123","response_type":"answer","content":"是指","done":false}

event: message
data: {"id":"msg-123","response_type":"answer","content":"血压","done":false}

event: message
data: {"id":"msg-123","response_type":"answer","content":"长期","done":false}

# N+1. 完成信号（最后一个answer事件）
event: message
data: {"id":"msg-123","response_type":"answer","content":"","done":true}
```

---

## 错误响应

### 请求解析错误（400）

```json
{
  "code": 400,
  "message": "Query content cannot be empty",
  "type": "BadRequest"
}
```

**可能的原因**:
- query为空
- session_id无效
- 请求体格式错误
- 自定义Agent不存在

### 执行错误（通过SSE返回）

```
event: message
data: {"response_type":"error","content":"Failed to retrieve from knowledge base","done":true}
```

**可能的原因**:
- 知识库检索失败
- LLM模型调用失败
- 网络连接问题
- 服务内部异常

---

## 客户端集成指南

### JavaScript/TypeScript（前端）

```typescript
// 使用前端API函数
import { knowledgeChat } from '@/api/chat';

async function askQuestion(sessionId: string, question: string) {
  const response = await knowledgeChat({
    session_id: sessionId,
    query: question
  });

  // 处理SSE流
  response.addEventListener('message', (event) => {
    const data = JSON.parse(event.data);
    
    if (data.response_type === 'references') {
      console.log('知识引用:', data.knowledge_references);
      displayReferences(data.knowledge_references);
    } else if (data.response_type === 'answer') {
      console.log('答案内容:', data.content);
      appendAnswerText(data.content);
      
      if (data.done) {
        console.log('答案生成完成');
        markAnswerComplete();
      }
    }
  });
}
```

### Go客户端

```go
import "github.com/UniverseHappiness/WiseDx/client"

func main() {
  apiClient := client.NewClient("http://localhost:8080", "sk-xxxxx")
  
  err := apiClient.KnowledgeQAStream(
    context.Background(),
    "session-123",
    &client.KnowledgeQARequest{
      Query: "什么是人工智能？",
    },
    func(response *client.StreamResponse) error {
      if response.ResponseType == client.ResponseTypeReferences {
        fmt.Printf("Found %d references\n", len(response.KnowledgeReferences))
      } else if response.ResponseType == client.ResponseTypeAnswer {
        fmt.Printf("Answer: %s", response.Content)
        if response.Done {
          fmt.Println("\nAnswer complete")
        }
      }
      return nil
    },
  )
}
```

### Python客户端

```python
import requests
import json

def knowledge_chat(session_id: str, query: str, api_key: str):
    url = f"http://localhost:8080/api/v1/knowledge-chat/{session_id}"
    headers = {
        "X-API-Key": api_key,
        "Content-Type": "application/json"
    }
    payload = {
        "query": query,
        "knowledge_base_ids": ["kb-001"]
    }
    
    response = requests.post(url, json=payload, headers=headers, stream=True)
    
    for line in response.iter_lines():
        if line.startswith(b'data:'):
            data = json.loads(line[5:])
            
            if data.get('response_type') == 'references':
                print("Knowledge References:")
                for ref in data.get('knowledge_references', []):
                    print(f"  - {ref['content'][:100]}... (score: {ref['score']})")
            
            elif data.get('response_type') == 'answer':
                print(data['content'], end='', flush=True)
                if data.get('done'):
                    print("\n[Complete]")

# 使用
knowledge_chat('session-123', '你的问题', 'sk-xxxxx')
```

---

## 常见场景和解决方案

### 场景1：指定特定知识库查询

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "查询医学信息",
    "knowledge_base_ids": ["kb-medicine-001", "kb-medicine-002"]
  }'
```

**处理流程**:
1. 系统只在指定的知识库中检索
2. 同时搜索两个知识库的文档
3. 结果合并后返回

### 场景2：使用自定义Agent配置

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "查询医学信息",
    "agent_id": "medical-assistant-v1"
  }'
```

**处理流程**:
1. 加载Agent的系统提示词
2. 应用Agent的模型选择和参数（温度、TopK等）
3. 使用Agent配置的知识库（如果未在请求中指定）
4. 应用Agent的其他特殊配置（如查询改写、重排序等）

### 场景3：禁用知识库检索（纯对话）

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "今天天气怎么样？",
    "knowledge_base_ids": []
  }'
```

**处理流程**:
1. 知识库ID列表为空
2. 不启用网络搜索
3. 使用chat_stream或chat_history_stream管道
4. 直接调用LLM进行对话（不进行RAG检索）

### 场景4：启用网络搜索

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "今天的新闻",
    "web_search_enabled": true,
    "knowledge_base_ids": []
  }'
```

**处理流程**:
1. 启用rag_stream管道
2. 执行网络搜索
3. 可选：同时检索知识库（如果指定了knowledge_base_ids）
4. 融合搜索结果后调用LLM总结

### 场景5：@提及特定文件

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "这个文件里讲了什么？",
    "knowledge_ids": ["know-001", "know-002"],
    "mentioned_items": [
      {
        "id": "know-001",
        "name": "心脏病指南",
        "type": "knowledge",
        "kb_type": "document"
      }
    ]
  }'
```

**处理流程**:
1. 优先使用knowledge_ids指定的文件
2. 只在这些特定文件中进行检索
3. mentioned_items用于前端UI展示（已@提及的项目）

### 场景6：覆盖默认模型

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "使用特定模型回答",
    "knowledge_base_ids": ["kb-001"],
    "summary_model_id": "gpt-4-turbo"
  }'
```

**优先级说明**:
1. 请求中的summary_model_id（**最高**）
2. Agent配置中的model_id
3. 知识库关联的SummaryModelID
4. 系统默认KnowledgeQA模型

---

## 性能和最佳实践

### 请求优化

1. **知识库选择**: 明确指定需要查询的知识库，避免全库检索
   ```json
   "knowledge_base_ids": ["kb-relevant"]  // ✓ 性能好
   // vs.
   // 不指定knowledge_base_ids  // ✗ 可能全库检索
   ```

2. **查询表述**: 使用清晰、具体的问题
   ```json
   "query": "高血压的治疗方案有哪些？"  // ✓ 清晰
   // vs.
   "query": "治疗"  // ✗ 太笼统
   ```

3. **知识ID优先**: 如果知道具体的知识文件ID，直接使用
   ```json
   "knowledge_ids": ["know-001"]  // ✓ 最快
   ```

### 超时和重试

- 建议设置HTTP超时：30-60秒
- SSE连接可能需要更长的等待时间
- 建议实现指数退避重试机制

### 并发处理

- 同一session可以处理多个并发请求
- 建议每个用户创建一个session，在session内维护对话历史
- 注意模型服务的并发限制

---

## 故障排除

### 问题1：收不到知识引用

**症状**: 答案返回了，但knowledge_references为空

**排查步骤**:
1. 检查知识库是否有相关文档：`GET /api/v1/knowledge-bases/{kb_id}/knowledge`
2. 查看向量模型配置是否正确
3. 尝试降低vector_threshold配置（允许更多相关性较低的结果）
4. 检查knowledge_base_ids是否正确

**解决方案**:
```bash
# 测试知识库搜索功能
curl -X POST 'http://localhost:8080/api/v1/knowledge-search' \
  -H 'X-API-Key: sk-xxxxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "你的问题",
    "knowledge_base_ids": ["kb-001"]
  }'
```

### 问题2：答案返回不完整或超时

**症状**: SSE连接中断，答案未完成

**排查步骤**:
1. 检查网络连接稳定性
2. 查看服务器日志是否有错误
3. 检查模型服务是否正常运行
4. 增加客户端超时时间

**解决方案**:
- 在客户端增加错误重试
- 监控event done=true标志确保完成
- 检查模型服务的日志

### 问题3：自定义Agent配置未生效

**症状**: agent_id指定了，但配置没有应用

**排查步骤**:
1. 确认agent_id是否存在：`GET /api/v1/agents/{agent_id}`
2. 查看Agent配置是否正确保存
3. 检查日志中是否输出"Using custom agent"

**解决方案**:
```bash
# 查询Agent详情
curl 'http://localhost:8080/api/v1/agents/agent-id' \
  -H 'X-API-Key: sk-xxxxx'
```

### 问题4：标题未自动生成

**症状**: session.title仍为空

**排查步骤**:
1. 确认session.title初始为空
2. 确认disable_title未设置为true
3. 检查聊天模型是否可用
4. 查看服务器日志中的标题生成过程

---

## 相关API文档

- [会话管理 API](/api/sessions.md)
- [知识库 API](/api/knowledge-bases.md)
- [自定义Agent API](/api/agents.md)
- [消息历史 API](/api/messages.md)
