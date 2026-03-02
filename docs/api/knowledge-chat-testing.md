# /api/v1/knowledge-chat 测试和集成示例

本文档提供各种测试场景和实现示例。

## 快速测试

### 使用curl测试基本功能

```bash
#!/bin/bash

# 配置
API_URL="http://localhost:8080"
API_KEY="sk-xxxxx"
SESSION_ID="your-session-id"

# 1. 测试基本查询
echo "=== 测试1: 基本知识库查询 ==="
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "什么是机器学习？"
  }' \
  -N  # 显示流式响应

# 2. 测试指定知识库
echo -e "\n\n=== 测试2: 指定知识库 ==="
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "查询医学内容",
    "knowledge_base_ids": ["kb-medicine"]
  }' \
  -N

# 3. 测试使用Agent
echo -e "\n\n=== 测试3: 使用自定义Agent ==="
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "查询医学内容",
    "agent_id": "medical-agent-v1"
  }' \
  -N

# 4. 测试禁用标题生成
echo -e "\n\n=== 测试4: 禁用标题生成 ==="
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "测试查询",
    "disable_title": true
  }' \
  -N

# 5. 测试网络搜索
echo -e "\n\n=== 测试5: 启用网络搜索 ==="
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "今天新闻",
    "web_search_enabled": true
  }' \
  -N
```

### 保存到文件并解析响应

```bash
#!/bin/bash

API_URL="http://localhost:8080"
API_KEY="sk-xxxxx"
SESSION_ID="your-session-id"

# 保存SSE响应到文件
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query": "测试"}' \
  -N > response.sse

# 使用Python脚本解析SSE
python3 << 'EOF'
import json
import re

with open('response.sse', 'r', encoding='utf-8') as f:
    content = f.read()

# 提取所有data行
data_lines = re.findall(r'data: ({.*})', content)

references_found = False
answer_text = ""

for line in data_lines:
    try:
        data = json.loads(line)
        
        if data.get('response_type') == 'references':
            print("=== 知识引用 ===")
            references_found = True
            for ref in data.get('knowledge_references', []):
                print(f"来源: {ref.get('knowledge_title')}")
                print(f"内容: {ref.get('content')[:100]}...")
                print(f"分数: {ref.get('score')}\n")
        
        elif data.get('response_type') == 'answer':
            print(data.get('content'), end='', flush=True)
            answer_text += data.get('content', '')
            if data.get('done'):
                print("\n\n=== 答案完成 ===")
    
    except json.JSONDecodeError:
        pass

print(f"\n总字数: {len(answer_text)}")
EOF
```

---

## 高级测试场景

### 场景1：多知识库联合查询

```bash
#!/bin/bash

API_URL="http://localhost:8080"
API_KEY="sk-xxxxx"
SESSION_ID="your-session-id"

# 同时查询多个知识库
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "请对比中西医治疗方法",
    "knowledge_base_ids": [
      "kb-chinese-medicine",
      "kb-western-medicine"
    ]
  }' \
  -N
```

### 场景2：特定文件查询

```bash
#!/bin/bash

API_URL="http://localhost:8080"
API_KEY="sk-xxxxx"
SESSION_ID="your-session-id"

# 只在指定的知识文件中查询
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "这份文档主要讲了什么？",
    "knowledge_ids": [
      "know-file-001",
      "know-file-002"
    ],
    "mentioned_items": [
      {
        "id": "know-file-001",
        "name": "2024年市场报告",
        "type": "knowledge",
        "kb_type": "document"
      }
    ]
  }' \
  -N
```

### 场景3：自定义模型和参数

```bash
#!/bin/bash

API_URL="http://localhost:8080"
API_KEY="sk-xxxxx"
SESSION_ID="your-session-id"

# 使用自定义模型和高级参数
curl -X POST "$API_URL/api/v1/knowledge-chat/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "详细解释",
    "knowledge_base_ids": ["kb-001"],
    "summary_model_id": "gpt-4-turbo",
    "agent_id": "advanced-agent"
  }' \
  -N
```

---

## 客户端集成示例

### TypeScript/JavaScript (前端)

```typescript
// frontend/src/api/chat/knowledge-chat.ts

import { API_BASE_URL, API_KEY } from '@/config';

interface KnowledgeChatRequest {
  query: string;
  knowledge_base_ids?: string[];
  knowledge_ids?: string[];
  summary_model_id?: string;
  web_search_enabled?: boolean;
  agent_id?: string;
  disable_title?: boolean;
  mentioned_items?: MentionedItem[];
}

interface StreamEvent {
  id: string;
  response_type: 'references' | 'answer' | 'error' | 'complete';
  content: string;
  done: boolean;
  knowledge_references?: KnowledgeReference[];
}

export class KnowledgeChatClient {
  private sessionId: string;

  constructor(sessionId: string) {
    this.sessionId = sessionId;
  }

  async chat(
    request: KnowledgeChatRequest,
    onEvent: (event: StreamEvent) => void,
    onError: (error: Error) => void
  ): Promise<void> {
    const url = `${API_BASE_URL}/api/v1/knowledge-chat/${this.sessionId}`;

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'X-API-Key': API_KEY,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('Response body is not readable');
      }

      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();

        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // 按行处理缓冲区
        const lines = buffer.split('\n');
        buffer = lines[lines.length - 1]; // 保留不完整的行

        for (let i = 0; i < lines.length - 1; i++) {
          const line = lines[i];

          // 解析SSE格式
          if (line.startsWith('data:')) {
            const dataStr = line.substring(5).trim();
            if (dataStr) {
              try {
                const event = JSON.parse(dataStr) as StreamEvent;
                onEvent(event);
              } catch (e) {
                console.error('Failed to parse event:', e);
              }
            }
          }
        }
      }

      // 处理缓冲区中的最后一行
      if (buffer.startsWith('data:')) {
        const dataStr = buffer.substring(5).trim();
        if (dataStr) {
          try {
            const event = JSON.parse(dataStr) as StreamEvent;
            onEvent(event);
          } catch (e) {
            console.error('Failed to parse last event:', e);
          }
        }
      }
    } catch (error) {
      onError(error as Error);
    }
  }
}

// 使用示例
async function main() {
  const client = new KnowledgeChatClient('session-123');
  let answerText = '';

  await client.chat(
    {
      query: '什么是机器学习？',
      knowledge_base_ids: ['kb-ai'],
    },
    (event) => {
      if (event.response_type === 'references') {
        console.log('Found references:');
        event.knowledge_references?.forEach((ref) => {
          console.log(`  - ${ref.knowledge_title}: ${ref.content.substring(0, 100)}`);
        });
      } else if (event.response_type === 'answer') {
        answerText += event.content;
        process.stdout.write(event.content);

        if (event.done) {
          console.log('\n\n[Complete]');
        }
      } else if (event.response_type === 'error') {
        console.error('Error:', event.content);
      }
    },
    (error) => {
      console.error('Client error:', error);
    }
  );
}
```

### Go 客户端

```go
// client/knowledge_chat_example.go

package client

import (
  "context"
  "fmt"
  "log"
)

func ExampleKnowledgeChat() {
  // 初始化客户端
  apiClient := NewClient("http://localhost:8080", "sk-xxxxx")

  sessionID := "session-123"
  question := "什么是机器学习？"

  // 创建请求
  request := &KnowledgeQARequest{
    Query:            question,
    KnowledgeBaseIDs: []string{"kb-ai"},
  }

  // 变量用于累积结果
  var answerText string
  var refCount int

  // 执行流式查询
  err := apiClient.KnowledgeQAStream(
    context.Background(),
    sessionID,
    request,
    func(response *StreamResponse) error {
      switch response.ResponseType {
      case ResponseTypeReferences:
        refCount = len(response.KnowledgeReferences)
        fmt.Printf("Found %d references:\n", refCount)
        for _, ref := range response.KnowledgeReferences {
          fmt.Printf("  - From: %s (score: %.2f)\n",
            ref.KnowledgeTitle, ref.Score)
        }

      case ResponseTypeAnswer:
        fmt.Print(response.Content)
        answerText += response.Content

        if response.Done {
          fmt.Printf("\n\n[Completed]\n")
        }

      case ResponseTypeError:
        fmt.Printf("Error: %s\n", response.Content)
        return fmt.Errorf(response.Content)
      }

      return nil
    },
  )

  if err != nil {
    log.Fatalf("Knowledge chat failed: %v", err)
  }

  // 输出统计
  fmt.Printf("\nStatistics:\n")
  fmt.Printf("  References: %d\n", refCount)
  fmt.Printf("  Answer length: %d chars\n", len(answerText))
}
```

### Python 客户端

```python
# scripts/knowledge_chat_client.py

import requests
import json
import sys
from typing import Callable, Optional

class KnowledgeChatClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.api_key = api_key

    def chat(
        self,
        session_id: str,
        query: str,
        on_event: Callable[[dict], None],
        knowledge_base_ids: Optional[list] = None,
        agent_id: Optional[str] = None,
        **kwargs
    ) -> None:
        """发送知识库查询请求"""
        url = f"{self.base_url}/api/v1/knowledge-chat/{session_id}"

        headers = {
            "X-API-Key": self.api_key,
            "Content-Type": "application/json",
        }

        payload = {
            "query": query,
        }
        if knowledge_base_ids:
            payload["knowledge_base_ids"] = knowledge_base_ids
        if agent_id:
            payload["agent_id"] = agent_id
        payload.update(kwargs)

        # 发送请求，启用流式响应
        response = requests.post(
            url,
            headers=headers,
            json=payload,
            stream=True,
        )

        if response.status_code != 200:
            raise Exception(f"HTTP {response.status_code}: {response.text}")

        # 逐行读取SSE流
        for line in response.iter_lines():
            if not line:
                continue

            line = line.decode('utf-8')

            # 解析SSE格式
            if line.startswith('data:'):
                data_str = line[5:].strip()
                if data_str:
                    try:
                        event = json.loads(data_str)
                        on_event(event)
                    except json.JSONDecodeError as e:
                        print(f"Failed to parse: {data_str}", file=sys.stderr)


def main():
    client = KnowledgeChatClient(
        "http://localhost:8080",
        "sk-xxxxx"
    )

    answer = ""
    ref_count = 0

    def on_event(event: dict):
        nonlocal answer, ref_count

        if event.get('response_type') == 'references':
            refs = event.get('knowledge_references', [])
            ref_count = len(refs)
            print(f"Found {len(refs)} references:")
            for ref in refs:
                print(f"  - {ref['knowledge_title']}: {ref['content'][:80]}...")

        elif event.get('response_type') == 'answer':
            content = event.get('content', '')
            answer += content
            sys.stdout.write(content)
            sys.stdout.flush()

            if event.get('done'):
                print("\n\n[Complete]")

        elif event.get('response_type') == 'error':
            print(f"Error: {event.get('content')}")

    try:
        client.chat(
            "session-123",
            "什么是机器学习？",
            on_event,
            knowledge_base_ids=["kb-ai"],
        )

        print(f"\nStatistics:")
        print(f"  References: {ref_count}")
        print(f"  Answer length: {len(answer)}")

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
```

---

## 单元测试示例

### Go单元测试

```go
// internal/handler/session/qa_test.go

package session

import (
  "context"
  "testing"

  "github.com/stretchr/testify/assert"
  "github.com/UniverseHappiness/WiseDx/internal/types"
)

func TestParseQARequest(t *testing.T) {
  // 准备test fixtures
  // ...

  testCases := []struct {
    name        string
    request     string
    expectedErr bool
  }{
    {
      name:        "valid request with query",
      request:     `{"query":"test question"}`,
      expectedErr: false,
    },
    {
      name:        "empty query",
      request:     `{"query":""}`,
      expectedErr: true,
    },
    {
      name:        "with knowledge_base_ids",
      request:     `{"query":"test","knowledge_base_ids":["kb-001"]}`,
      expectedErr: false,
    },
  }

  for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
      // 执行测试
      // ...
    })
  }
}

func TestKnowledgeQAIntegration(t *testing.T) {
  // 集成测试
  // 1. 创建session
  // 2. 调用KnowledgeQA
  // 3. 验证SSE响应
  // 4. 检查数据库记录
}
```

---

## 性能测试

### 使用Apache Bench进行负载测试

```bash
#!/bin/bash

# 需要先保存请求体到文件
cat > request.json << 'EOF'
{
  "query": "test question",
  "knowledge_base_ids": ["kb-001"]
}
EOF

# 运行100个请求，10个并发
ab -n 100 -c 10 \
  -H "X-API-Key: sk-xxxxx" \
  -H "Content-Type: application/json" \
  -p request.json \
  'http://localhost:8080/api/v1/knowledge-chat/session-123'
```

### 使用 wrk 进行性能测试

```bash
# 安装: brew install wrk

# 创建lua脚本
cat > test.lua << 'EOF'
request = function()
  wrk.method = "POST"
  wrk.headers["X-API-Key"] = "sk-xxxxx"
  wrk.headers["Content-Type"] = "application/json"
  wrk.body = '{"query":"test"}'
  return wrk.format(nil, "/api/v1/knowledge-chat/session-123")
end
EOF

# 运行测试
# 持续30秒，4个线程，10个连接
wrk -t4 -c10 -d30s -s test.lua 'http://localhost:8080'
```

---

## 监控和调试

### 使用tcpdump捕获网络流量

```bash
# 捕获所有HTTP请求
sudo tcpdump -i any 'tcp port 8080' -A | grep -E '(POST|query|response_type)'
```

### 启用详细日志

```yaml
# config.yaml
logging:
  level: debug
  format: json
  output: stdout

# 或保存到文件
logging:
  level: debug
  format: json
  output: logs/app.log
```

### 使用Jaeger追踪请求

```bash
# 需要在config.yaml中启用tracing
tracing:
  enabled: true
  jaeger:
    endpoint: http://localhost:14268/api/traces
```

### 监控EventBus

在服务中添加event监听以跟踪流程：

```go
eventBus.On(event.EventAgentQuery, func(ctx context.Context, evt event.Event) error {
  logger.Infof(ctx, "EventAgentQuery: %v", evt)
  return nil
})

eventBus.On(event.EventAgentReferences, func(ctx context.Context, evt event.Event) error {
  data := evt.Data.(event.AgentReferencesData)
  logger.Infof(ctx, "EventAgentReferences: %d results", len(data.References))
  return nil
})

eventBus.On(event.EventAgentFinalAnswer, func(ctx context.Context, evt event.Event) error {
  data := evt.Data.(event.AgentFinalAnswerData)
  logger.Infof(ctx, "EventAgentFinalAnswer: %d chars, done=%v", len(data.Content), data.Done)
  return nil
})
```

---

## 常见集成问题和解决方案

### 问题1：SSE连接断开

**症状**: 连接在发送几个事件后中断

**排查步骤**:
1. 检查Nginx/反向代理配置
2. 检查firewall/LB是否有超时设置
3. 增加服务端心跳

**解决方案**:
```nginx
# nginx配置
location /api/v1/knowledge-chat/ {
  proxy_pass http://backend;
  proxy_buffering off;  # 禁用缓冲
  proxy_cache off;
  proxy_set_header Connection "";
  proxy_http_version 1.1;
  
  # 增加超时
  proxy_read_timeout 120s;
  proxy_send_timeout 120s;
  
  # 禁用keepalive超时
  keepalive_timeout 0;
}
```

### 问题2：重复接收同一条事件

**症状**: 某些answer事件被重复处理

**原因**: 客户端SSE处理逻辑问题

**解决方案**:
```typescript
// 使用message_id跟踪已处理事件
const processedIds = new Set<string>();

eventSource.addEventListener('message', (event) => {
  const data = JSON.parse(event.data);
  const eventId = `${data.id}-${data.response_type}-${data.content.substring(0, 10)}`;
  
  if (processedIds.has(eventId)) {
    return; // 跳过重复
  }
  
  processedIds.add(eventId);
  // 处理事件...
});
```

### 问题3：内存泄漏

**症状**: 长时间运行后内存不断增加

**排查步骤**:
1. 检查EventBus监听器是否正确注销
2. 检查context是否正确cancel
3. 检查数据库连接是否关闭

**解决方案**:
```go
// 确保context cancel
defer cancel()

// 确保EventBus监听器注销
unsubscribe := eventBus.On(event.EventType, handler)
defer unsubscribe()

// 确保数据库资源释放
rows, err := db.Query(...)
defer rows.Close()
```
