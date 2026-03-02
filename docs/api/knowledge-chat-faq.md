# /api/v1/knowledge-chat 常见问题解答

## 基础问题

### Q1: knowledge-chat 和 agent-chat 有什么区别？

**A**: 
| 方面 | knowledge-chat | agent-chat |
|------|----------------|-----------|
| 用途 | 基于知识库的简单问答 | 支持工具调用的智能体问答 |
| 工具调用 | ❌ 不支持 | ✅ 支持 |
| 多轮推理 | ❌ 基础对话 | ✅ 完整推理循环 |
| 思考过程 | ❌ 隐藏 | ✅ 可见 |
| 速度 | ✅ 快（一次总结） | ❌ 慢（多次调用） |
| 使用场景 | FAQ、文档查询 | 复杂问题、需要多步骤 |

**选择建议**:
- 用户只是查询信息 → knowledge-chat
- 用户需要完整分析、多工具协作 → agent-chat

---

### Q2: 为什么响应是SSE格式而不是JSON？

**A**: SSE（Server-Sent Events）的优势：

1. **流式传输**: 可以逐步接收答案，而不是等待所有内容
2. **更好的UX**: 用户能立即看到内容，不用等待完成
3. **处理长答案**: 适合处理很长的回复（几千字符）
4. **标准化**: SSE是Web标准，浏览器原生支持
5. **单向通信**: 服务器主动推送，不需要轮询

**对比方案**:
```
JSON 轮询:   客户端 → 服务器 → 客户端 → 服务器...（延迟大）
WebSocket:   双向通信（过度工程）
SSE:        单向推送（完美适配）✅
```

---

### Q3: query 参数有什么限制吗？

**A**: 
| 方面 | 限制 |
|------|------|
| **最小长度** | 1 个字符 |
| **最大长度** | 通常 4096 个字符（可配置） |
| **特殊字符** | 支持所有UTF-8字符 |
| **换行符** | 支持 |
| **编码** | UTF-8 |

**最佳实践**:
```json
// ✅ 好的query
{
  "query": "如何治疗高血压？"  // 清晰、具体
}

// ❌ 不好的query
{
  "query": "治疗"  // 太笼统，召回率低
}

{
  "query": "aaaaaaaaaa..."  // 垃圾输入，浪费资源
}
```

---

### Q4: 一次最多可以查询多少个知识库？

**A**: 理论上没有硬限制，但实际受以下因素影响：

```
系统推荐: 最多 3-5 个知识库
原因:
  - 检索时间随知识库数量线性增加
  - 结果合并和重排序成本增加
  - LLM 上下文窗口有限制

实际建议:
  ≤ 3 KB   → 响应 < 5 秒
  3-10 KB  → 响应 5-15 秒
  > 10 KB  → 可能超时
```

**优化策略**:
```json
// ❌ 不优化：同时查询10个KB
{
  "knowledge_base_ids": ["kb-1", "kb-2", ..., "kb-10"]
}

// ✅ 优化：在应用层过滤，只查询相关KB
{
  "knowledge_base_ids": ["kb-medicine"]  // 医学问题只查医学KB
}
```

---

## 功能问题

### Q5: 可以在不使用知识库的情况下使用knowledge-chat吗？

**A**: 完全可以！这相当于一个普通的LLM对话接口。

```bash
curl -X POST 'http://localhost:8080/api/v1/knowledge-chat/session-123' \
  -H 'X-API-Key: sk-xxxxx' \
  -d '{
    "query": "什么是人工智能？",
    "knowledge_base_ids": []  # ← 空数组表示不使用KB
  }'
```

**处理流程**:
```
知识库ID为空
  ├─ 不执行RAG检索
  └─ 直接调用LLM
       └─ 返回answer事件（无references）
```

---

### Q6: 如何保证多轮对话的上下文连贯？

**A**: 通过 session 机制实现：

```
Session 概念:
  session = 一个对话会话
  
  第1条消息: query1 → session_id
  第2条消息: query2 → 同一个session_id
  第3条消息: query3 → 同一个session_id
  
系统自动:
  ✅ 维护消息历史
  ✅ 传递上下文给LLM
  ✅ 理解指代（如"它"、"这个"）
```

**最佳实践**:
```typescript
// 创建session一次
const session = await createSession();

// 在同一session中进行多轮对话
await knowledgeChat(session.id, "什么是机器学习？");
await knowledgeChat(session.id, "它的应用有哪些？");     // 理解"它"=机器学习
await knowledgeChat(session.id, "还有其他吗？");         // 继续前面的话题
```

---

### Q7: agent_id 如果不存在会怎样？

**A**: 系统会记录警告但继续处理：

```go
// 源代码行为
agent, err := h.customAgentService.GetAgentByID(ctx, request.AgentID)
if err != nil {
  logger.Warnf(ctx, "Failed to get custom agent... using default config")
  // 继续执行，不抛出错误
  // ↓
  customAgent = nil  // 降级为无Agent模式
}
```

**日志输出**:
```
WARN: Failed to get custom agent, agent ID: xxx, error: ..., using default config
```

**客户端表现**: 看不出区别，使用了默认配置

---

### Q8: disable_title 有什么作用？

**A**: 控制是否自动生成会话标题。

```
场景1: 第1条消息，session.title为空
  disable_title=false (默认)
    └─ 异步生成标题（使用用户问题 + LLM生成）
       └─ 更新session.title
       └─ 发送title_updated事件给前端

场景2: 第2条及以后的消息
  disable_title=任何值
    └─ 忽略（session已有标题，不再生成）

场景3: 开发/测试
  disable_title=true
    └─ 完全禁用标题生成（加快响应）
```

**何时使用**:
```json
// ✅ 第一条消息，让系统生成标题
{
  "query": "第一个问题",
  "disable_title": false
}

// ✅ 后续消息，标题已存在，加快响应
{
  "query": "后续问题",
  "disable_title": true
}

// ✅ 测试环境
{
  "query": "测试",
  "disable_title": true
}
```

---

## 配置问题

### Q9: 如何覆盖全局配置中的vector_threshold？

**A**: 有两种方式：

**方式1: 通过Agent配置**（推荐）
```bash
# 创建或更新Agent时设置
PUT /api/v1/agents/my-agent
{
  "config": {
    "vector_threshold": 0.5  # 覆盖全局的0.7
  }
}

# 然后查询时指定
POST /api/v1/knowledge-chat/session-123
{
  "query": "...",
  "agent_id": "my-agent"
}
```

**方式2: 通过请求参数**（目前不直接支持）
```json
// ❌ 目前API不支持直接传threshold参数
{
  "query": "...",
  "vector_threshold": 0.5  // 这个参数不存在
}
```

**优先级说明**:
```
config.yaml 全局设置: vector_threshold: 0.7
  ↑ 被覆盖
Agent配置: vector_threshold: 0.5
  ↓ 最终使用 0.5
```

---

### Q10: web_search_enabled 和 knowledge_base_ids 可以同时使用吗？

**A**: 完全可以！系统会融合两种搜索结果。

```json
{
  "query": "近期AI发展有什么新闻？",
  "knowledge_base_ids": ["kb-ai"],      // 查知识库
  "web_search_enabled": true             // + 网络搜索
}
```

**处理流程**:
```
┌─ 知识库检索 → 结果A
│
├─ 网络搜索   → 结果B
│
└─ 融合 A+B  → LLM总结 → 最终答案
```

**好处**: 
- 知识库提供体系化知识
- 网络搜索提供最新信息
- 融合给出最完整的答案

---

## 错误处理

### Q11: 为什么我得到"No chat model ID available"错误？

**A**: 模型选择链路中没有找到可用模型。

**排查步骤**:
```
1. 检查请求中是否指定了summary_model_id
   ✓ 指定 → 检查该模型是否存在
   ✗ 未指定 → 继续

2. 检查Agent配置是否指定了model_id
   ✓ 指定 → 检查该模型是否存在
   ✗ 未指定 → 继续

3. 检查知识库是否配置了SummaryModelID
   ✓ 配置 → 检查该模型是否存在
   ✗ 未配置 → 继续

4. 系统中是否有ModelTypeKnowledgeQA的模型？
   ✓ 有 → 应该成功
   ✗ 没有 → 这是根本原因！
```

**解决方案**:
```bash
# 1. 检查可用模型
GET /api/v1/models
# 查看是否有type=knowledge_qa的模型

# 2. 创建模型
POST /api/v1/models
{
  "name": "gpt-4",
  "type": "knowledge_qa"
}

# 3. 关联知识库
PUT /api/v1/knowledge-bases/kb-001
{
  "summary_model_id": "gpt-4"
}
```

---

### Q12: SSE连接经常中断怎么办？

**A**: 多个可能的原因和解决方案：

**原因1: 网络超时**
```yaml
# config.yaml - 增加超时
server:
  read_timeout: 120s     # 默认可能是30s
  write_timeout: 120s
```

**原因2: 反向代理配置**
```nginx
# nginx 配置
location /api/v1/knowledge-chat/ {
  proxy_buffering off;        # 关键！禁用缓冲
  proxy_cache off;            # 禁用缓存
  proxy_read_timeout 120s;    # 增加超时
  proxy_send_timeout 120s;
  proxy_http_version 1.1;     # 使用HTTP/1.1
  proxy_set_header Connection "";  # 保持连接
}
```

**原因3: 客户端未正确处理SSE**
```typescript
// ❌ 错误的处理
const response = await fetch(...);
const text = await response.text();  // 等待全部内容！

// ✅ 正确的处理
const response = await fetch(...);
const reader = response.body.getReader();
while (true) {
  const { done, value } = await reader.read();  // 逐步读取
  if (done) break;
}
```

---

### Q13: 为什么有时knowledge_references为空？

**A**: 多个原因导致无检索结果：

**原因1: 知识库没有相关文档**
```
解决: 添加更多知识库内容
```

**原因2: 向量阈值过高**
```yaml
# config.yaml
conversation:
  vector_threshold: 0.7  # 太高，无相关结果

# 改为
  vector_threshold: 0.3  # 更宽松
```

**原因3: 向量模型不匹配**
```
问题: 知识库用模型A索引，查询用模型B编码
解决: 确保使用同一个向量模型
```

**原因4: 知识库未建立索引**
```bash
# 检查知识库状态
GET /api/v1/knowledge-bases/kb-001
# 查看chunks数量，是否为0
```

**排查顺序**:
```
1. 验证知识库有内容: GET /api/v1/knowledge-bases/{id}/knowledge
2. 检查向量阈值配置
3. 测试知识库搜索: POST /api/v1/knowledge-search
4. 检查向量模型配置
```

---

## 性能问题

### Q14: 为什么响应很慢？

**A**: 可能的性能瓶颈及优化：

**瓶颈1: 知识库检索慢**
```
原因: 
  - 知识库太大（几百万chunks）
  - 没有建立合适的索引
  
优化:
  1. 限制检索的知识库数量
     "knowledge_base_ids": ["kb-relevant"]  // 不要查所有
  
  2. 使用知识ID而非知识库ID
     "knowledge_ids": ["know-001"]  // 直接查特定文件
  
  3. 调整TopK参数（通过Agent配置）
     "embedding_top_k": 5  // 不要太大
```

**瓶颈2: LLM处理慢**
```
原因:
  - 模型响应慢（网络、CPU）
  - 上下文窗口太大
  
优化:
  1. 使用更快的模型
     "summary_model_id": "gpt-3.5"  // 而非gpt-4
  
  2. 限制历史轮数
     Agent配置: history_turns: 3  // 不要保留太多
```

**瓶颈3: 网络延迟**
```
原因:
  - 客户端和服务器距离远
  - 网络质量差
  
优化:
  1. 部署在更近的位置
  2. 增加超时时间
  3. 使用CDN加速
```

**性能基准**:
```
场景: 查1个KB，5个chunks，GPT3.5模型
  ├─ 知识检索: 300-500ms
  ├─ LLM生成: 1-3秒
  └─ 总时间: 1.3-3.5秒

优化后:
  ├─ 知识ID查询: 100-200ms
  ├─ 精简模型: 0.5-1秒
  └─ 总时间: 0.6-1.2秒
```

---

### Q15: 并发请求有限制吗？

**A**: 没有硬限制，但受实际资源约束：

```
理论无限制，实际受以下限制:

1. 数据库连接池
   └─ 默认: 10-20 连接
   └─ 建议: 同时处理 10-20 个请求
   
2. LLM服务限制
   └─ 取决于外部模型服务的QPS限制
   └─ 如调用OpenAI: 通常 100-500 req/s
   
3. 服务器资源
   └─ CPU: 16核服务器可处理 50-100 req/s
   └─ 内存: 16GB 可缓存 1-2 个并发回话

实际建议:
  小型部署: ≤ 10 req/s
  中型部署: 10-50 req/s
  大型部署: 需要负载均衡 + 多实例
```

---

## Agent相关问题

### Q16: Agent 会一直使用agent-chat还是可能回落到knowledge-chat？

**A**: Agent 绑定到 agent-chat 路由后，不会回落到 knowledge-chat。

```
// 如果在 agent-chat 端点调用
POST /api/v1/agent-chat/session-123
{
  "query": "...",
  "agent_id": "my-agent"
}
└─ 总是使用Agent模式

// 但可以在 knowledge-chat 端点调用不指定agent
POST /api/v1/knowledge-chat/session-123
{
  "query": "..."
  // 不指定agent_id
}
└─ 使用normal模式，不使用Agent
```

**选择哪个端点**:
```
knowledge-chat:     用于普通知识库查询
agent-chat:         用于需要Agent能力的查询
```

---

### Q17: 如何为不同的问题类型使用不同的Agent？

**A**: 在应用层或前端实现路由逻辑：

```typescript
// 应用层路由
function selectAgent(query: string): string {
  if (query.includes("医学") || query.includes("病")) {
    return "medical-agent";
  } else if (query.includes("法律") || query.includes("合同")) {
    return "legal-agent";
  } else if (query.includes("代码") || query.includes("编程")) {
    return "code-agent";
  }
  return null;  // 不指定，使用默认
}

// 使用
const agentId = selectAgent(userQuery);
await knowledgeChat(sessionId, {
  query: userQuery,
  agent_id: agentId
});
```

---

## 数据和安全

### Q18: 用户问题和答案是否被保存？

**A**: 是的，所有消息都被保存到数据库：

```sql
-- 保存的数据
INSERT INTO messages (
  session_id,         -- 会话ID
  role,               -- 'user' 或 'assistant'
  content,            -- 完整的问题或答案
  request_id,         -- 请求追踪ID
  created_at          -- 时间戳
) VALUES (...)
```

**数据链**:
```
用户问题 → API → 创建user消息 → DB
回答生成 → API → 创建assistant消息 → DB
前端查询历史 → 从DB读取消息
```

**隐私考虑**:
- 所有消息都被存储
- 敏感信息应该在应用层加密
- 可以实现定期清理策略（如保留30天）

---

### Q19: 如何实现多租户隔离？

**A**: 通过 TenantID 在API级别实现：

```go
// 从请求头或认证信息提取TenantID
tenantID := extractTenantFromAuth(c)

// 传递给所有操作
session, err := h.sessionService.GetSession(ctx, sessionID)
// 内部检查: session.TenantID == tenantID

// 防止跨租户访问
if session.TenantID != tenantID {
  return errors.NewForbiddenError("Access denied")
}
```

**实现要点**:
```
✅ API认证时提取TenantID
✅ 所有DB操作带上TenantID过滤
✅ 知识库/Agent等资源也关联TenantID
✅ 定期审计确保隔离完整
```

---

## 整合和部署

### Q20: 如何在Docker中部署并保证SSE正常工作？

**A**: Docker 部署时需要特殊配置：

```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .

RUN go build -o wisedx ./cmd/server

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./wisedx"]
```

```yaml
# docker-compose.yml
services:
  wisedx:
    build: .
    ports:
      - "8080:8080"
    environment:
      - WISEDX_LOG_LEVEL=info
    volumes:
      - ./config:/app/config
      - ./data:/app/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

```nginx
# nginx 配置关键部分
location /api/v1/knowledge-chat/ {
  proxy_pass http://wisedx:8080;
  proxy_buffering off;           # 关键！
  proxy_cache off;
  proxy_http_version 1.1;
  proxy_set_header Connection "";
  proxy_read_timeout 120s;
  proxy_send_timeout 120s;
}
```

---

### Q21: 如何监控knowledge-chat的性能？

**A**: 实现多层面监控：

**应用层指标**:
```go
// 记录请求延迟
start := time.Now()
err := h.sessionService.KnowledgeQA(...)
duration := time.Since(start)
metrics.ObserveKnowledgeQADuration(duration)
```

**关键指标**:
```
1. 响应时间分布
   - p50: < 2s
   - p95: < 5s
   - p99: < 10s

2. 成功率
   - 目标: > 99%
   
3. 引用覆盖率
   - KB有结果的百分比
   
4. 资源使用
   - CPU: < 80%
   - 内存: < 75%
   - DB连接: < 80%
```

**Prometheus 导出示例**:
```yaml
# metrics
wisedx_knowledge_qa_duration_seconds{percentile="p50"} 1.2
wisedx_knowledge_qa_duration_seconds{percentile="p95"} 4.5
wisedx_knowledge_qa_duration_seconds{percentile="p99"} 8.9
wisedx_knowledge_qa_errors_total 5
wisedx_knowledge_qa_requests_total 1000
```

---

## 故障排除快速索引

| 症状 | 可能原因 | 快速修复 |
|------|--------|--------|
| 400错误 | query为空 | 检查请求body |
| 404错误 | session不存在 | 先创建session |
| 无references | KB无相关内容 | 降低vector_threshold |
| 答案不完整 | SSE断开 | 检查超时配置 |
| 响应慢 | KB查询多 | 指定单个KB |
| Agent未生效 | agent_id不存在 | 检查Agent存在 |
| 内存泄漏 | Context未cancel | 检查defer cancel() |
| 断开连接 | nginx缓冲 | 设置proxy_buffering off |

---

**有其他问题？**
请查阅 [详细指南](./knowledge-chat-detailed.md) 或 [实现参考](./knowledge-chat-implementation.md)
