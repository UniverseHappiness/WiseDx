# /api/v1/knowledge-chat 实现参考

本文档详细说明后端如何处理knowledge-chat请求，以及数据如何在各个层级流转。

## 请求处理流程详解

### 1. HTTP请求到达

```
POST /api/v1/knowledge-chat/session-123
Headers:
  X-API-Key: sk-xxxxx
  Content-Type: application/json
Body:
{
  "query": "什么是机器学习？",
  "knowledge_base_ids": ["kb-ai"]
}
```

### 2. 路由层（router.go）

**文件**: `internal/router/router.go` 第271-275行

```go
func RegisterChatRoutes(r *gin.RouterGroup, handler *session.Handler) {
  knowledgeChat := r.Group("/knowledge-chat")
  {
    knowledgeChat.POST("/:session_id", handler.KnowledgeQA)
  }
}
```

**职责**: 
- 将请求路由到 `Handler.KnowledgeQA()` 方法
- 提取URL中的session_id参数

### 3. 处理器层（qa.go）

**文件**: `internal/handler/session/qa.go` 第250-260行

```go
func (h *Handler) KnowledgeQA(c *gin.Context) {
  // 1. 解析和验证请求
  reqCtx, request, err := h.parseQARequest(c, "KnowledgeQA")
  if err != nil {
    c.Error(err)
    return
  }

  // 2. 执行Normal模式QA（不是Agent模式）
  h.executeNormalModeQA(reqCtx, !request.DisableTitle)
}
```

#### 3.1 请求解析（parseQARequest）

**位置**: `qa.go` 第37-112行

处理步骤：

```
┌─────────────────────────────────────┐
│ 1. 提取session_id（从URL路径参数）   │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ 2. 解析JSON请求体                   │
│    - query                           │
│    - knowledge_base_ids             │
│    - knowledge_ids                  │
│    - summary_model_id              │
│    - web_search_enabled            │
│    - agent_id                       │
│    - disable_title                 │
│    - mentioned_items               │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ 3. 验证查询不为空                    │
│    if query == "" {                  │
│      return BadRequestError         │
│    }                                 │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ 4. 从DB获取Session对象              │
│    session, err := h.sessionService │
│      .GetSession(ctx, sessionID)    │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ 5. 如果指定了agent_id，获取Agent   │
│    agent, err := h.customAgentService│
│      .GetAgentByID(ctx, agentID)    │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ 6. 构建qaRequestContext对象         │
│    包含所有已验证和解析的数据        │
└─────────────────────────────────────┘
```

**关键代码**:

```go
type qaRequestContext struct {
  ctx              context.Context
  c                *gin.Context
  sessionID        string
  requestID        string
  query            string
  session          *types.Session
  customAgent      *types.CustomAgent
  assistantMessage *types.Message
  knowledgeBaseIDs []string
  knowledgeIDs     []string
  summaryModelID   string
  webSearchEnabled bool
  mentionedItems   types.MentionedItems
}
```

#### 3.2 执行Normal模式QA（executeNormalModeQA）

**位置**: `qa.go` 第304-398行

```
┌─────────────────────────────────┐
│ 1. 创建User消息                 │
│    保存用户提出的问题到数据库     │
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│ 2. 创建Assistant消息            │
│    准备用于存储AI回复的消息对象   │
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│ 3. 设置SSE流                    │
│    - 设置HTTP响应头              │
│    - 创建EventBus               │
│    - 写入初始agent_query事件    │
│    - 设置停止事件处理器          │
│    - 配置流处理器               │
│    - 启动标题异步生成（如需）    │
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│ 4. 注册EventBus监听器           │
│    监听EventAgentFinalAnswer事件 │
│    当done=true时标记完成         │
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│ 5. 异步执行KnowledgeQA服务      │
│    (在独立Goroutine中)          │
│    - 调用sessionService.KnowledgeQA │
│    - 处理错误事件               │
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│ 6. 处理SSE事件流（阻塞）        │
│    handleAgentEventsForSSE()    │
│    - 监听EventBus中的事件       │
│    - 转换为SSE格式写入响应      │
│    - 直到EventAgentComplete    │
└─────────────────────────────────┘
```

### 4. 服务层（sessionService.KnowledgeQA）

**文件**: `internal/application/service/session.go` 第398-681行

#### 4.1 知识库解析阶段

```go
// 确定最终使用哪些知识库
hasExplicitMention := len(knowledgeBaseIDs) > 0 || len(knowledgeIDs) > 0

if hasExplicitMention {
  // 使用请求明确指定的知识库（优先级最高）
  // keep knowledgeBaseIDs and knowledgeIDs as-is
} else if customAgent != nil && customAgent.Config.RetrieveKBOnlyWhenMentioned {
  // Agent配置要求仅在@提及时检索，清空知识库ID
  knowledgeBaseIDs = nil
  knowledgeIDs = nil
} else {
  // 使用Agent配置中的默认知识库
  knowledgeBaseIDs = s.resolveKnowledgeBasesFromAgent(ctx, customAgent)
}
```

**优先级树**:

```
┌─ 请求中有knowledge_base_ids/knowledge_ids？
│  ├─ YES → 使用请求的 (优先级最高)
│  └─ NO
│     ├─ Agent启用RetrieveKBOnlyWhenMentioned？
│     │  ├─ YES → 清空(禁用KB检索)
│     │  └─ NO → 使用Agent默认KB
│     └─ 无Agent → 使用config.yaml全局默认KB
```

#### 4.2 模型选择阶段

```go
chatModelID, err := s.selectChatModelIDWithOverride(
  ctx, session, knowledgeBaseIDs, knowledgeIDs, summaryModelID
)
```

**优先级顺序**:

```
1. 请求参数summaryModelID (最高)
   └─ 用户可以覆盖一切

2. Agent配置 (如果绑定了Agent)
   └─ customAgent.Config.ModelID

3. 知识库配置
   └─ 遍历knowledgeBaseIDs，查找Remote模型
   └─ 使用第一个KB的SummaryModelID

4. 系统默认 (最低)
   └─ 第一个ModelTypeKnowledgeQA的模型
```

#### 4.3 配置合并阶段

如果绑定了自定义Agent，合并其配置：

```go
if customAgent != nil {
  customAgent.EnsureDefaults()
  
  // 各种配置覆盖
  if summaryModelID == "" && customAgent.Config.ModelID != "" {
    chatModelID = customAgent.Config.ModelID
  }
  if customAgent.Config.SystemPrompt != "" {
    summaryConfig.Prompt = customAgent.Config.SystemPrompt
  }
  if customAgent.Config.ContextTemplate != "" {
    summaryConfig.ContextTemplate = customAgent.Config.ContextTemplate
  }
  // ... 更多配置 ...
}
```

**配置来源优先级**:

```
全局默认(config.yaml)
        ↑
      覆盖
        |
  自定义Agent配置
        ↑
      覆盖
        |
  请求参数
```

#### 4.4 检索目标构建

```go
searchTargets, err := s.buildSearchTargets(
  ctx, session.TenantID, knowledgeBaseIDs, knowledgeIDs
)
```

**构建逻辑**:

```
Input: knowledgeBaseIDs, knowledgeIDs
  ├─ 如果knowledgeIDs不为空
  │  └─ 直接使用knowledgeIDs
  │
  └─ 否则，遍历knowledgeBaseIDs
     └─ 获取每个KB下的所有Knowledge
     └─ 构建SearchTarget数组

Output: []SearchTarget
  ├─ 每个元素代表一个要检索的对象
  └─ 后续RAG管道使用这个数组
```

#### 4.5 管道选择

```go
var pipeline []types.EventType

if len(knowledgeBaseIDs) == 0 && len(knowledgeIDs) == 0 && !webSearchEnabled {
  // 无知识库、无网络搜索 → 纯对话模式
  if maxRounds > 0 {
    pipeline = types.Pipline["chat_history_stream"]  // 支持多轮
  } else {
    pipeline = types.Pipline["chat_stream"]          // 单轮
  }
} else {
  // 有知识库或网络搜索 → RAG模式
  pipeline = types.Pipline["rag_stream"]
}
```

**管道流程**:

```
chat_stream:
  └─ 直接调用LLM
  └─ 不进行知识检索

chat_history_stream:
  └─ 加载历史对话
  └─ 直接调用LLM（考虑历史上下文）
  └─ 不进行知识检索

rag_stream:
  ├─ 知识检索
  │  ├─ 向量相似度搜索
  │  └─ 关键词搜索
  ├─ 结果合并和重排
  ├─ (可选) 查询改写
  ├─ (可选) 网络搜索
  └─ LLM总结回答
```

#### 4.6 构建ChatManage对象

```go
chatManage := &types.ChatManage{
  Query:                query,
  RewriteQuery:         query,
  SessionID:            session.ID,
  MessageID:            assistantMessageID,
  KnowledgeBaseIDs:     knowledgeBaseIDs,
  KnowledgeIDs:         knowledgeIDs,
  SearchTargets:        searchTargets,
  VectorThreshold:      vectorThreshold,
  KeywordThreshold:     keywordThreshold,
  EmbeddingTopK:        embeddingTopK,
  RerankModelID:        rerankModelID,
  RerankTopK:           rerankTopK,
  ChatModelID:          chatModelID,
  SummaryConfig:        summaryConfig,
  EventBus:             eventBus,
  WebSearchEnabled:     webSearchEnabled,
  // ... 更多字段 ...
}
```

#### 4.7 触发管道执行

```go
err = s.KnowledgeQAByEvent(ctx, chatManage, pipeline)
```

**KnowledgeQAByEvent职责**:
1. 初始化事件系统
2. 按照pipeline顺序执行各个处理阶段
3. 使用EventBus在各阶段之间传递数据和事件
4. 最终将MergeResult保存到chatManage中

#### 4.8 发送引用事件

```go
if len(chatManage.MergeResult) > 0 {
  eventBus.Emit(ctx, event.Event{
    ID:        generateEventID("references"),
    Type:      event.EventAgentReferences,
    SessionID: session.ID,
    Data: event.AgentReferencesData{
      References: chatManage.MergeResult,
    },
  })
}
```

### 5. 事件处理层（handleAgentEventsForSSE）

**文件**: `internal/handler/session/event_handler.go`

**职责**: 监听EventBus事件，并将其转换为SSE格式写入HTTP响应

```
EventBus (来自服务层)
    ├─ EventAgentReferences
    │  └─ SSE handler 监听
    │     └─ 转换为references事件
    │        └─ 写入: {"response_type":"references","knowledge_references":[...]}
    │
    ├─ EventAgentFinalAnswer
    │  └─ SSE handler 监听
    │     └─ 转换为answer事件
    │        └─ 写入: {"response_type":"answer","content":"...","done":false}
    │
    ├─ EventError
    │  └─ SSE handler 监听
    │     └─ 转换为error事件
    │        └─ 写入: {"response_type":"error","content":"..."}
    │
    └─ EventAgentComplete
       └─ SSE handler 监听
          └─ 发送最终的done=true事件
          └─ 关闭响应流
```

**SSE响应格式**:

```
event: message
data: {JSON_DATA}

```

每个事件后跟一个空行，用于在客户端SSE解析器中标记事件边界。

### 6. 完整数据流图

```
┌─────────────────────────┐
│  HTTP POST 请求         │
│  /api/v1/knowledge-chat │
│  /{session_id}          │
└────────────┬────────────┘
             │
┌────────────▼────────────┐
│  Router                 │
│  - 提取session_id       │
└────────────┬────────────┘
             │
┌────────────▼──────────────────────┐
│  Handler.KnowledgeQA()            │
│  - parseQARequest()               │
│  - executeNormalModeQA()          │
│  - setupSSEStream()               │
└────────────┬──────────────────────┘
             │
┌────────────▼──────────────────────────────────────────────────┐
│  异步执行: sessionService.KnowledgeQA()                        │
│  ├─ 知识库解析                                                  │
│  ├─ 模型选择                                                    │
│  ├─ 配置合并                                                    │
│  ├─ 检索目标构建                                                │
│  ├─ 管道选择                                                    │
│  ├─ KnowledgeQAByEvent()                                       │
│  │  ├─ RAG管道/Chat管道执行                                    │
│  │  ├─ 知识检索 (if RAG)                                       │
│  │  ├─ 结果合并到MergeResult                                   │
│  │  └─ 发送EventAgentFinalAnswer                              │
│  └─ 发送EventAgentReferences                                  │
└────────────┬──────────────────────────────────────────────────┘
             │
┌────────────▼──────────────────────────────────────────────────┐
│  SSE Event Handler (handleAgentEventsForSSE)                   │
│  ├─ 监听EventAgentReferences                                   │
│  │  └─ 写入references SSE事件                                   │
│  ├─ 监听EventAgentFinalAnswer                                  │
│  │  └─ 流式写入answer SSE事件                                   │
│  ├─ 监听EventError                                              │
│  │  └─ 写入error SSE事件                                        │
│  └─ 监听EventAgentComplete                                     │
│     └─ 发送完成信号并关闭流                                      │
└────────────┬──────────────────────────────────────────────────┘
             │
┌────────────▼─────────────────────────┐
│  HTTP Response (SSE Stream)           │
│  event: message                       │
│  data: {...}                          │
│                                       │
│  event: message                       │
│  data: {...}                          │
│                                       │
│  ... (多个事件) ...                   │
└───────────────────────────────────────┘
```

## 数据库操作

### 会话操作

```
1. 读取会话信息
   └─ Query: SELECT * FROM sessions WHERE id = ?
   └─ 获取session.title, session.TenantID等

2. 创建用户消息
   └─ INSERT INTO messages (session_id, role, content, ...)
   └─ role = 'user'
   └─ content = query

3. 创建助手消息
   └─ INSERT INTO messages (session_id, role, content, ...)
   └─ role = 'assistant'
   └─ is_completed = false
   └─ 最初content为空

4. 更新助手消息
   └─ UPDATE messages SET content = ?, is_completed = true, updated_at = ?
   └─ 当回答完成时
```

### 知识库操作

```
1. 获取知识库信息
   └─ SELECT * FROM knowledge_bases WHERE id IN (?)
   └─ 获取SummaryModelID等配置

2. 获取知识文件列表
   └─ SELECT * FROM knowledge WHERE kb_id = ?
   └─ 构建SearchTargets

3. 向量检索
   └─ 调用embedding_engine执行向量相似度搜索
   └─ 获取Top-K相似的chunks

4. 关键词检索
   └─ 使用倒排索引查询关键词匹配
   └─ 获取匹配的chunks

5. 结果合并
   └─ 将向量和关键词结果合并
   └─ 按分数排序
   └─ 重排序（如配置了rerank模型）
```

## 配置覆盖优先级总结

### 知识库ID选择

```
1. 请求参数knowledge_base_ids ────────┐
                                       ├─→ 最终使用的KB ID列表
2. (仅当请求无KB时) Agent配置KB ──────┘
```

### 知识文件ID选择

```
如果knowledge_ids非空
  └─ 优先使用knowledge_ids
  └─ 忽略knowledge_base_ids
否则
  └─ 使用knowledge_base_ids
```

### 聊天模型选择

```
1. 请求参数summary_model_id (最高优先级)
   │
2. (仅当请求未指定时) Agent配置model_id
   │
3. (仅当无Agent时) 知识库配置SummaryModelID
   │
4. 系统中第一个KnowledgeQA类型模型 (最低优先级)
```

### 系统提示词选择

```
Agent配置systemPrompt (如果Agent绑定)
   │
config.yaml中的conversation.summary.prompt
```

### 其他参数选择

```
参数 | 请求 | Agent配置 | config.yaml | 系统默认
----|------|---------|-----------|--------
temperature | 优先1 | 优先2 | 优先3 | 优先4
max_completion_tokens | 优先1 | 优先2 | 优先3 | 优先4
embedding_top_k | 优先1 | 优先2 | 优先3 | 优先4
rerank_top_k | 优先1 | 优先2 | 优先3 | 优先4
vector_threshold | 优先1 | 优先2 | 优先3 | 优先4
enable_rewrite | 优先1 | 优先2 | 优先3 | 优先4
```

## 错误处理流程

### 解析阶段错误

发生在Handler.KnowledgeQA()中，直接返回HTTP 400错误：

```
Request Parsing Error
    └─ Session not found → BadRequest (400)
    └─ Query empty → BadRequest (400)
    └─ Invalid JSON → BadRequest (400)
    └─ Agent not found → Warning (but continue) → Log and use default
```

### 执行阶段错误

发生在sessionService.KnowledgeQA()执行中，通过SSE返回：

```
Execution Error
    └─ Knowledge retrieval failed
    │  └─ Emit EventError
    │     └─ SSE: {"response_type":"error",...}
    │
    └─ LLM call failed
    │  └─ Emit EventError
    │     └─ SSE: {"response_type":"error",...}
    │
    └─ Model not found
    │  └─ Emit EventError
    │     └─ SSE: {"response_type":"error",...}
    │
    └─ Panic recovery
       └─ Stack trace logged
       └─ Emit EventError
          └─ SSE: {"response_type":"error",...}
```

## 性能考虑

### 并发处理

```
1. SSE连接持久化
   └─ 单个请求的生命周期内保持连接
   └─ 允许多次EventEmit写入

2. Goroutine隔离
   └─ 服务层在独立Goroutine执行
   └─ Handler主线程处理SSE事件和写入
   └─ 不阻塞其他请求

3. EventBus解耦
   └─ 服务和SSE处理通过EventBus通信
   └─ 异步处理，不需要直接等待
```

### 资源管理

```
1. Context生命周期
   └─ Request context 用于请求级操作
   └─ Async context 用于后台处理
   └─ Cancel context 支持停止操作

2. 数据库连接
   └─ 连接池管理
   └─ 及时关闭Reader
   └─ 事务管理

3. 内存管理
   └─ MergeResult中的chunk缓存
   └─ EventBus事件快速传递和清理
   └─ 避免大规模数据在内存中堆积
```

## 调试建议

### 启用详细日志

在config.yaml中：

```yaml
logging:
  level: debug
  format: json
```

### 关键日志点

监控这些日志输出以跟踪请求流程：

```
[KnowledgeQA] Start processing request
[KnowledgeQA] Request: session_id=xxx, request={...}
Knowledge base question answering parameters, session ID: xxx
KB resolution: hasExplicitMention=true
Using knowledge bases: [kb-001]
Creating chat manage object, knowledge base IDs: [kb-001]
Triggering question answering event
Knowledge base question answering initiated
Emitting references event with N results
Knowledge QA service completed for session: xxx
```

### 测试建议

```bash
# 1. 测试简单查询
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-id' \
  -H 'X-API-Key: sk-xxxxx' \
  -d '{"query": "test"}'

# 2. 使用工具监控SSE
# 安装: npm install -g sse-tool
sse-tool 'http://localhost:8080/api/v1/knowledge-chat/session-id' -X POST -d '{"query":"test"}'

# 3. 检查服务器日志
tail -f logs/app.log | grep KnowledgeQA
```
