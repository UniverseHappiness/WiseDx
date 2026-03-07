# WiseDx 聊天流水线插件系统详解

## 一、架构概述

聊天流水线采用**事件驱动 + 插件化**的架构模式，将 RAG 流程的各个阶段抽象为独立的插件，通过事件触发机制串联执行。

### 核心数据结构

```go
// 插件接口
type Plugin interface {
    OnEvent(ctx, eventType, chatManage, next) *PluginError  // 事件处理
    ActivationEvents() []EventType                          // 声明监听哪些事件
}

// 事件管理器
type EventManager struct {
    listeners map[EventType][]Plugin      // 事件 → 插件列表
    handlers  map[EventType]func(...)     // 事件 → 处理链函数
}
```

---

## 二、全部插件列表（16个）

| 插件名称 | 事件类型 | 文件 | 功能说明 |
|---------|---------|------|---------|
| **PluginLoadHistory** | LOAD_HISTORY | load_history.go | 加载对话历史，支持多轮对话 |
| **PluginRewrite** | REWRITE_QUERY | rewrite.go | 使用LLM结合对话历史重写用户查询 |
| **PluginExtractEntity** | REWRITE_QUERY | extract_entity.go | 从查询中提取实体（配合Neo4j） |
| **PluginSearch** | CHUNK_SEARCH | search.go | 混合检索（向量+关键词）+ 网络搜索 |
| **PluginSearchParallel** | CHUNK_SEARCH_PARALLEL | search_parallel.go | 并行执行 Chunk 搜索和实体搜索 |
| **PluginSearchEntity** | ENTITY_SEARCH | search_entity.go | 图数据库实体关系搜索 |
| **PluginRerank** | CHUNK_RERANK | rerank.go | 使用重排模型对搜索结果重新排序 |
| **PluginMerge** | CHUNK_MERGE | merge.go | 合并重叠Chunk、扩展短上下文、FAQ增强 |
| **PluginFilterTopK** | FILTER_TOP_K | filter_top_k.go | 保留 Top K 结果 |
| **PluginIntoChatMessage** | INTO_CHAT_MESSAGE | into_chat_message.go | 将搜索结果转换为聊天消息格式 |
| **PluginDataAnalysis** | DATA_ANALYSIS | data_analysis.go | CSV/Excel 文件分析（生成DuckDB SQL） |
| **PluginChatCompletion** | CHAT_COMPLETION | chat_completion.go | 同步调用 LLM 完成对话 |
| **PluginChatCompletionStream** | CHAT_COMPLETION_STREAM | chat_completion_stream.go | 流式调用 LLM，实时返回响应 |
| **PluginStreamFilter** | STREAM_FILTER | stream_filter.go | 流式输出过滤，处理无匹配场景 |
| **PluginTracing** | 多种事件 | tracing.go | OpenTelemetry 链路追踪和可观测性 |

---

## 三、预定义流水线（5种模式）

```go
var Pipline = map[string][]EventType{
    "chat":               {CHAT_COMPLETION},
    "chat_stream":        {CHAT_COMPLETION_STREAM, STREAM_FILTER},
    "chat_history_stream": {LOAD_HISTORY, CHAT_COMPLETION_STREAM, STREAM_FILTER},
    "rag":                {CHUNK_SEARCH, CHUNK_RERANK, CHUNK_MERGE, INTO_CHAT_MESSAGE, CHAT_COMPLETION},
    "rag_stream":         {REWRITE_QUERY, CHUNK_SEARCH_PARALLEL, CHUNK_RERANK, CHUNK_MERGE, 
                           FILTER_TOP_K, DATA_ANALYSIS, INTO_CHAT_MESSAGE, 
                           CHAT_COMPLETION_STREAM, STREAM_FILTER},
}
```

---

## 四、插件注册流程

### Step 1: 创建插件并注册

```go
// rerank.go
func NewPluginRerank(eventManager *EventManager, modelService interfaces.ModelService) *PluginRerank {
    res := &PluginRerank{modelService: modelService}
    eventManager.Register(res)  // 关键：注册到事件管理器
    return res
}
```

### Step 2: 插件声明监听的事件类型

```go
func (p *PluginRerank) ActivationEvents() []types.EventType {
    return []types.EventType{types.CHUNK_RERANK}  // 监听 CHUNK_RERANK 事件
}
```

### Step 3: EventManager.Register() 内部处理

```go
func (e *EventManager) Register(plugin Plugin) {
    for _, eventType := range plugin.ActivationEvents() {
        // 1. 将插件添加到 listeners 映射
        e.listeners[eventType] = append(e.listeners[eventType], plugin)
        
        // 2. 重新构建该事件的处理链
        e.handlers[eventType] = e.buildHandler(e.listeners[eventType])
    }
}
```

---

## 五、洋葱模型处理链构建

### buildHandler 实现

```go
func (e *EventManager) buildHandler(plugins []Plugin) func(...) *PluginError {
    // 最内层：空函数
    next := func(...) *PluginError { return nil }
    
    // 从后向前包裹，形成洋葱结构
    for i := len(plugins) - 1; i >= 0; i-- {
        current := plugins[i]
        prevNext := next      // ⭐ 关键：保存当前 next 的值
        next = func(ctx, eventType, chatManage) *PluginError {
            return current.OnEvent(ctx, eventType, chatManage, func() *PluginError {
                return prevNext(ctx, eventType, chatManage)
            })
        }
    }
    return next
}
```

### 洋葱结构图示

```
假设 CHUNK_RERANK 有 2 个插件：[PluginRerank, PluginTracing]

调用 Trigger(CHUNK_RERANK) 时：

┌──────────────────────────────────────────────┐
│  最外层 next 函数                              │
│  ┌────────────────────────────────────────┐  │
│  │  PluginRerank.OnEvent()                │  │
│  │    ... 执行重排逻辑 ...                  │  │
│  │    调用 next() ─────────────────┐      │  │
│  │  ┌──────────────────────────────│────┐ │  │
│  │  │  PluginTracing.OnEvent()     ↓    │ │  │
│  │  │    ... 执行追踪逻辑 ...             │ │  │
│  │  │    调用 next() ──────────┐        │ │  │
│  │  │  ┌───────────────────────│──────┐ │ │  │
│  │  │  │   return nil          ↓      │ │ │  │
│  │  │  └──────────────────────────────┘ │ │  │
│  │  └───────────────────────────────────┘ │  │
│  └────────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

### 为什么要用 prevNext？

**闭包捕获变量问题**：

```go
// ❌ 错误写法：闭包捕获同一个 next，导致无限递归
next = func() { return next() }

// ✅ 正确写法：每轮创建新变量保存当前值
prevNext := next
next = func() { return prevNext() }
```

`prevNext := next` 在每轮循环创建**新变量**，"冻结"当前值，避免被后续覆盖。

---

## 六、事件触发流程

### Trigger 方法

```go
func (e *EventManager) Trigger(ctx context.Context,
    eventType types.EventType, chatManage *types.ChatManage,
) *PluginError {
    if handler, ok := e.handlers[eventType]; ok {
        return handler(ctx, eventType, chatManage)  // 执行整个处理链
    }
    return nil
}
```

---

## 七、Pipeline 执行（顺序执行，非洋葱）

```go
func (s *sessionService) KnowledgeQAByEvent(ctx context.Context,
    chatManage *types.ChatManage, eventList []types.EventType,
) error {
    // 顺序执行每个事件
    for _, eventType := range eventList {
        err := s.eventManager.Trigger(ctx, eventType, chatManage)
        if err != nil {
            return err  // 出错就停止
        }
    }
    return nil
}
```

### 两层架构对比

| 层级 | 执行方式 | 特点 |
|------|---------|------|
| **Pipeline（事件序列）** | 顺序 for 循环 | 一个完成才执行下一个，出错立即停止 |
| **Event 内部（多插件）** | 洋葱模型 | 嵌套调用，通过 next() 传递 |

---

## 八、数据流转

所有插件通过 `ChatManage` 对象共享数据：

```
REWRITE_QUERY      写入: chatManage.RewriteQuery, chatManage.Entity
        ↓
CHUNK_SEARCH       读取: RewriteQuery  写入: chatManage.SearchResult
        ↓
CHUNK_RERANK       读取: SearchResult  写入: chatManage.RerankResult
        ↓
CHUNK_MERGE        读取: RerankResult  写入: chatManage.MergeResult
        ↓
INTO_CHAT_MESSAGE  读取: MergeResult   写入: chatManage.UserContent
        ↓
CHAT_COMPLETION    读取: UserContent   写入: chatManage.ChatResponse
```

---

## 九、Go 语言知识点

### 1. 隐式接口实现（鸭子类型）

Go 语言中，**不需要显式声明** `implements`：

```go
// Go 的方式：只要实现了接口的所有方法，自动满足接口
type PluginRerank struct { ... }

func (p *PluginRerank) OnEvent(...) *PluginError { ... }      // ✅ 实现
func (p *PluginRerank) ActivationEvents() []EventType { ... } // ✅ 实现

// *PluginRerank 自动满足 Plugin 接口
```

### 2. append 函数

```go
e.listeners[eventType] = append(e.listeners[eventType], plugin)
```

- `append(slice, element)` 向切片末尾追加元素
- 返回新切片，必须重新赋值
- 原因：容量不足时会分配新底层数组

---

## 十、面试问答

### Q1: 为什么选择事件驱动+插件化架构？

**答**：
1. **解耦** - 各阶段独立开发，互不影响
2. **可扩展** - 新增功能只需添加插件（开闭原则）
3. **灵活组合** - 不同场景用不同 Pipeline

### Q2: `buildHandler` 中为什么要用 `prevNext := next`？

**答**：闭包捕获变量的引用，如果直接用 `next`，所有闭包都指向同一变量。`prevNext := next` 在每轮循环创建新变量，"冻结"当前值。

### Q3: 这个架构用了哪些设计模式？

**答**：
- **策略模式** - 不同 Pipeline 策略
- **责任链/中间件模式** - 插件链式调用
- **观察者模式** - EventManager 监听分发事件
- **依赖注入** - 插件构造时注入 Service

### Q4: 洋葱模型和普通责任链的区别？

**答**：
- **责任链**：单向传递，某节点处理后可终止
- **洋葱模型**：每个中间件有"进入"和"退出"两个时机，可做前置/后置处理

### Q5: 如何新增插件？

**答**：
1. 实现 `Plugin` 接口
2. `ActivationEvents()` 返回要监听的事件
3. 在 `OnEvent()` 中实现业务逻辑
4. 调用 `eventManager.Register()` 注册

不需要修改任何已有代码，符合开闭原则。

### Q6: 这个架构有什么缺点？

**答**：
1. **调试复杂** - 嵌套闭包，堆栈追踪不直观
2. **顺序敏感** - 插件注册顺序影响执行
3. **内存开销** - 每个事件都构建闭包链
