<template>
    <div class="chat-container" :class="{ 'sidebar-open': showPatientSidebar && hasCollectedData }">
        <div class="chat">
            <div class="chat-header">
                <div class="chat-header-left">
                    <div class="chat-header-title" :title="currentSessionTitle">{{ currentSessionTitle }}</div>
                </div>
                <div class="chat-header-actions">
                    <t-dropdown :options="exportOptions" trigger="click" placement="bottom-right" @click="handleExportSelect">
                        <t-button variant="text" size="small" class="export-button">
                            <t-icon name="download" />
                            <span class="export-button-text">{{ t('chat.export') }}</span>
                        </t-button>
                    </t-dropdown>
                </div>
            </div>
            <div ref="scrollContainer" class="chat_scroll_box" @scroll="handleScroll">
                <div class="msg_list">
                    <div v-for="(session, id) in messagesList" :key='id'>
                        <div v-if="session.role == 'user'">
                            <usermsg :content="session.content" :mentioned_items="session.mentioned_items"></usermsg>
                        </div>
                        <div v-if="session.role == 'assistant'">
                            <botmsg :content="session.content" :session="session" :user-query="getUserQuery(id)" @scroll-bottom="scrollToBottom"
                                :isFirstEnter="isFirstEnter" @quick-reply-select="handleQuickReplySelectFromBot"></botmsg>
                        </div>
                    </div>
                    <div v-if="loading"
                        style="height: 41px;display: flex;align-items: center;padding-left: 4px;">
                        <div class="loading-typing">
                            <span></span>
                            <span></span>
                            <span></span>
                        </div>
                    </div>
                </div>
            </div>
            <div style="min-height: 115px; margin: 16px auto 4px;width: 100%;max-width: 800px;">
                <InputField 
                    @send-msg="(query, modelId, mentionedItems) => sendMsg(query, modelId, mentionedItems)" 
                    @stop-generation="handleStopGeneration"
                    :isReplying="isReplying" 
                    :sessionId="session_id"
                    :assistantMessageId="currentAssistantMessageId"
                ></InputField>
            </div>
        </div>
        
        <!-- Patient Info Sidebar -->
        <PatientInfoSidebar
            v-if="isMedicalAgent"
            :collected-data="aggregatedCollectedData"
            :visible="showPatientSidebar && hasCollectedData"
            @close="showPatientSidebar = false"
        />
        
        <!-- Sidebar Toggle Button -->
        <div 
            v-if="isMedicalAgent && hasCollectedData" 
            class="sidebar-toggle"
            :class="{ 'sidebar-open': showPatientSidebar }"
            @click="showPatientSidebar = !showPatientSidebar"
        >
            <t-icon :name="showPatientSidebar ? 'chevron-right' : 'chevron-left'" />
            <span v-if="!showPatientSidebar">{{ $t('agent.patientInfo') }}</span>
        </div>
    </div>
    <KnowledgeBaseEditorModal 
        :visible="uiStore.showKBEditorModal"
        :mode="uiStore.kbEditorMode"
        :kb-id="uiStore.currentKBId || undefined"
        :initial-type="uiStore.kbEditorType"
        @update:visible="(val) => val ? null : uiStore.closeKBEditor()"
        @success="handleKBEditorSuccess"
    />
</template>
<script setup>
import { storeToRefs } from 'pinia';
import { ref, onMounted, onUnmounted, nextTick, watch, reactive, onBeforeUnmount, computed } from 'vue';
import { useRoute, useRouter, onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router';
import InputField from '../../components/Input-field.vue';
import botmsg from './components/botmsg.vue';
import usermsg from './components/usermsg.vue';
import PatientInfoSidebar from './components/PatientInfoSidebar.vue';
import { getMessageList, generateSessionsTitle, getSession } from "@/api/chat/index";
import { useStream } from '../../api/chat/streame'
import { useMenuStore } from '@/stores/menu';
import { useSettingsStore } from '@/stores/settings';
import { MessagePlugin } from 'tdesign-vue-next';
import { useI18n } from 'vue-i18n';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { useUIStore } from '@/stores/ui';
import KnowledgeBaseEditorModal from '@/views/knowledge/KnowledgeBaseEditorModal.vue';
import { useKnowledgeBaseCreationNavigation } from '@/hooks/useKnowledgeBaseCreationNavigation';
const usemenuStore = useMenuStore();
const useSettingsStoreInstance = useSettingsStore();
const uiStore = useUIStore();
const { navigateToKnowledgeBaseList } = useKnowledgeBaseCreationNavigation();
const { t } = useI18n();

marked.use({
    mangle: false,
    headerIds: false,
    breaks: true
});
const { menuArr, isFirstSession, firstQuery, firstMentionedItems, firstModelId } = storeToRefs(usemenuStore);
const { output, onChunk, isStreaming, isLoading, error, startStream, stopStream } = useStream();
const route = useRoute();
const router = useRouter();
const session_id = ref(route.params.chatid);
const sessionData = ref(null);
const created_at = ref('');
const limit = ref(20);
const messagesList = reactive([]);
const isReplying = ref(false);
const currentAssistantMessageId = ref(''); // 当前正在生成的 assistant message ID
const scrollLock = ref(false);
const isNeedTitle = ref(false);
const isFirstEnter = ref(true);
const loading = ref(false);
let fullContent = ref('')
let userquery = ref('')

const currentSessionTitle = computed(() => {
    const chatMenu = menuArr.value?.[2];
    const item = chatMenu?.children?.find((c) => c.id === session_id.value);
    return item?.title || t('createChat.newSessionTitle');
});

const exportOptions = computed(() => [
    { content: t('chat.exportMarkdown'), value: 'markdown' },
    { content: t('chat.exportPDF'), value: 'pdf' }
]);
const scrollContainer = ref(null)
const handleKBEditorSuccess = (kbId) => {
    navigateToKnowledgeBaseList(kbId)
}

// Patient Info Sidebar state
const showPatientSidebar = ref(true);

// Check if current session is medical consultant agent
const isMedicalAgent = computed(() => {
    return sessionData.value?.agent_id === 'builtin-medical-consultant';
});

// Aggregate collected_data from all messages
const aggregatedCollectedData = computed(() => {
    const result = {};
    
    for (const msg of messagesList) {
        if (msg.role === 'assistant' && msg.agentEventStream) {
            // Find all todo_write events with collected_data
            const todoWriteEvents = msg.agentEventStream.filter(
                e => e.type === 'tool_call' && e.tool_name === 'todo_write' && e.tool_data?.collected_data
            );
            
            // Merge collected_data from each event
            for (const event of todoWriteEvents) {
                const data = event.tool_data.collected_data;
                if (data && typeof data === 'object') {
                    Object.assign(result, data);
                }
            }
        }
    }
    
    return result;
});

// Check if there is any collected data
const hasCollectedData = computed(() => {
    return Object.keys(aggregatedCollectedData.value).length > 0;
});

const getUserQuery = (index) => {
    if (index <= 0) {
        return '';
    }
    const previous = messagesList[index - 1];
    if (previous && previous.role === 'user') {
        return previous.content || '';
    }
    return '';
};

const sanitizeFileName = (name) => {
    return String(name || '')
        .trim()
        .replace(/[\\/:*?"<>|]/g, '_')
        .replace(/\s+/g, ' ')
        .slice(0, 80) || 'chat';
};

const downloadTextFile = (content, filename, mimeType) => {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
};

const getAssistantExportContent = (msg) => {
    if (msg?.isAgentMode && Array.isArray(msg.agentEventStream)) {
        const answerEvent = [...msg.agentEventStream].reverse().find((e) => e?.type === 'answer');
        const answer = answerEvent?.content;
        if (typeof answer === 'string') return answer;
    }
    if (typeof msg?.content === 'string') return msg.content;
    return '';
};

const formatMentionedItems = (mentionedItems) => {
    if (!Array.isArray(mentionedItems) || mentionedItems.length === 0) return '';
    return mentionedItems
        .map((it) => {
            const name = it?.name || it?.id || '';
            return name ? `@${name}` : '';
        })
        .filter(Boolean)
        .join(' ');
};

const formatKnowledgeReferences = (refs) => {
    if (!Array.isArray(refs) || refs.length === 0) return [];
    return refs.map((r) => {
        const title = r?.title || r?.document_title || r?.file_name || r?.name || r?.url || r?.id || '';
        const chunkId = r?.chunk_id || r?.chunkId;
        const docId = r?.document_id || r?.documentId;
        const metaParts = [
            chunkId ? `${t('chat.chunkIdLabel')} ${String(chunkId).slice(0, 12)}...` : '',
            docId ? `${t('chat.documentIdLabel')} ${String(docId).slice(0, 12)}...` : ''
        ].filter(Boolean);
        const meta = metaParts.length ? ` (${metaParts.join(', ')})` : '';
        return `- ${title || t('chat.otherSource')}${meta}`;
    });
};

const buildConversationMarkdown = () => {
    if (!Array.isArray(messagesList) || messagesList.length === 0) return '';

    const lines = [];
    const title = currentSessionTitle.value || t('chat.title');
    const now = new Date();

    lines.push(`# ${title}`);
    lines.push('');
    lines.push(`- ${t('chat.exportSessionId')}: ${session_id.value}`);
    lines.push(`- ${t('chat.exportGeneratedAt')}: ${now.toLocaleString()}`);
    lines.push('');

    for (const msg of messagesList) {
        if (!msg) continue;

        if (msg.role === 'user') {
            lines.push(`## ${t('chat.exportRoleUser')}`);
            lines.push('');
            const mentions = formatMentionedItems(msg.mentioned_items);
            if (mentions) {
                lines.push(`> ${t('chat.exportMentions')}: ${mentions}`);
                lines.push('');
            }
            lines.push(String(msg.content || '').trim());
            lines.push('');
            lines.push('---');
            lines.push('');
            continue;
        }

        if (msg.role === 'assistant') {
            lines.push(`## ${t('chat.exportRoleAssistant')}`);
            lines.push('');
            const content = getAssistantExportContent(msg).trim();
            lines.push(content || `_${t('chat.exportEmptyAssistant')}_`);
            lines.push('');

            const refLines = formatKnowledgeReferences(msg.knowledge_references);
            if (refLines.length) {
                lines.push(`### ${t('chat.reference')}`);
                lines.push(...refLines);
                lines.push('');
            }

            lines.push('---');
            lines.push('');
        }
    }

    return lines.join('\n');
};

const exportMarkdown = () => {
    const md = buildConversationMarkdown();
    if (!md) {
        MessagePlugin.warning(t('chat.exportEmpty'));
        return;
    }

    const base = sanitizeFileName(currentSessionTitle.value);
    const filename = `${base || 'chat'}-${String(session_id.value).slice(0, 12)}.md`;

    try {
        downloadTextFile(md, filename, 'text/markdown;charset=utf-8');
        MessagePlugin.success(t('chat.exportSuccess'));
    } catch (e) {
        console.error('[Export Markdown] failed:', e);
        MessagePlugin.error(t('chat.exportFailed'));
    }
};

const exportPDF = () => {
    const md = buildConversationMarkdown();
    if (!md) {
        MessagePlugin.warning(t('chat.exportEmpty'));
        return;
    }

    const base = sanitizeFileName(currentSessionTitle.value);
    const title = `${base || 'chat'}-${String(session_id.value).slice(0, 12)}`;

    // 采用浏览器打印为 PDF（更轻量、无需引入额外依赖）
    const rawHtml = marked.parse(md);
    const safeHtml = DOMPurify.sanitize(String(rawHtml || ''));

    const win = window.open('', '_blank');
    if (!win) {
        MessagePlugin.error(t('chat.exportPopupBlocked'));
        return;
    }

    win.document.open();
    win.document.write(`<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>${title}</title>
  <style>
    body { font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,"PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif; padding: 24px; color: #111; }
    h1 { font-size: 22px; margin: 0 0 16px; }
    h2 { font-size: 18px; margin: 20px 0 10px; }
    h3 { font-size: 16px; margin: 16px 0 8px; }
    hr { border: none; border-top: 1px solid #eee; margin: 18px 0; }
    pre { background: #f6f8fa; padding: 12px; overflow: auto; border-radius: 6px; }
    code { font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace; }
    blockquote { margin: 12px 0; padding: 8px 12px; border-left: 3px solid #ddd; color: #555; background: #fafafa; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border: 1px solid #eee; padding: 6px 8px; }
    img { max-width: 100%; }
    @media print { body { padding: 0; } }
  </style>
</head>
<body>
${safeHtml}
</body>
</html>`);
    win.document.close();

    // 给浏览器一点时间渲染
    setTimeout(() => {
        try {
            win.focus();
            win.print();
            MessagePlugin.success(t('chat.exportPDFTip'));
        } catch (e) {
            console.error('[Export PDF] failed:', e);
            MessagePlugin.error(t('chat.exportFailed'));
        }
    }, 300);
};

const handleExportSelect = (data) => {
    const value = data?.value;
    if (value === 'markdown') {
        exportMarkdown();
        return;
    }
    if (value === 'pdf') {
        exportPDF();
    }
};

// Handle quick reply selection from bot message
const handleQuickReplySelectFromBot = (value) => {
    console.log('[Chat] Quick reply selected:', value);
    // Send the selected value as a user message
    sendMsg(value);
};
watch([() => route.params], (newvalue) => {
    isFirstEnter.value = true;
    if (newvalue[0].chatid) {
        if (!firstQuery.value) {
            scrollLock.value = false;
        }
        messagesList.splice(0);
        session_id.value = newvalue[0].chatid;
        
        // 切换会话时，重置状态（加载历史消息不应显示loading）
        loading.value = false;
        isReplying.value = false;
        currentAssistantMessageId.value = '';
        
        checkmenuTitle(session_id.value)
        let data = {
            session_id: session_id.value,
            created_at: '',
            limit: limit.value
        }
        getmsgList(data);
    }
});
const scrollToBottom = () => {
    nextTick(() => {
        if (scrollContainer.value) {
            scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight;
        }
    })
}
const debounce = (fn, delay) => {
    let timer
    return (...args) => {
        clearTimeout(timer)
        timer = setTimeout(() => fn(...args), delay)
    }
}
const onChatScrollTop = () => {
    if (scrollLock.value) return;
    const { scrollTop, scrollHeight } = scrollContainer.value;
    isFirstEnter.value = false
    if (scrollTop == 0) {
        let data = {
            session_id: session_id.value,
            created_at: created_at.value,
            limit: limit.value
        }
        getmsgList(data, true, scrollHeight);
    }
}
const handleScroll = debounce(onChatScrollTop, 500);

const getmsgList = (data, isScrollType = false, scrollHeight) => {
    getMessageList(data).then(res => {
        if (res && res.data?.length) {
            created_at.value = res.data[0].created_at;
            handleMsgList(res.data, isScrollType, scrollHeight);
        }
    })
}

// Reconstruct agentEventStream from agent_steps stored in database
// This allows the frontend to restore the exact conversation state including all agent reasoning steps
const reconstructEventStreamFromSteps = (agentSteps, messageContent, isCompleted = false) => {
    const events = [];
    
    // Process agent steps if they exist
    if (agentSteps && Array.isArray(agentSteps) && agentSteps.length > 0) {
    agentSteps.forEach((step) => {
        // Add thinking event if thought content exists
        if (step.thought && step.thought.trim()) {
            events.push({
                type: 'thinking',
                event_id: `step-${step.iteration}-thought`,
                content: step.thought,
                done: true,
                thinking: false,
                // Extract duration from step if available
                duration_ms: step.duration || undefined,
            });
        }
        
        // Add tool call and result events
        if (step.tool_calls && Array.isArray(step.tool_calls)) {
            step.tool_calls.forEach((toolCall) => {
                events.push({
                    type: 'tool_call',
                    tool_call_id: toolCall.id,
                    tool_name: toolCall.name,
                    arguments: toolCall.args,
                    pending: false,
                    success: toolCall.result?.success !== false,
                    output: toolCall.result?.output || '',
                    error: toolCall.result?.error || undefined,
                    // Use both duration and duration_ms for compatibility
                    duration: toolCall.duration,
                    duration_ms: toolCall.duration,
                    display_type: toolCall.result?.data?.display_type,
                    tool_data: toolCall.result?.data,
                });
                
                // If this is a show_options tool, also add a quick_reply event
                if (toolCall.name === 'show_options' && toolCall.result?.data?.emit_quick_reply) {
                    events.push({
                        type: 'quick_reply',
                        event_id: `${toolCall.id}-quick-reply`,
                        question: toolCall.result.data.question || '',
                        options: toolCall.result.data.options || [],
                        multi_select: toolCall.result.data.multi_select || false,
                        timestamp: Date.now()
                    });
                }
            });
        }
    });
    }
    
    // 总是添加 answer 事件如果有内容（无论是否有 agent_steps）
    // 这样可以确保最终答案始终被渲染
    if (messageContent && messageContent.trim()) {
        events.push({
            type: 'answer',
            content: messageContent,
            done: true
        });
    } else if (isCompleted) {
        // 如果消息已完成但 content 为空（Agent 模式常见情况），添加一个空的 answer 事件标记完成
        // 这样可以确保 isConversationDone 返回 true，不显示 loading-indicator
        events.push({
            type: 'answer',
            content: '',
            done: true
        });
    }
    
    return events;
};
const handleMsgList = async (data, isScrollType = false, newScrollHeight) => {
    let chatlist = data.reverse()
    for (let i = 0, len = chatlist.length; i < len; i++) {
        let item = chatlist[i];
        item.isAgentMode = false; // Agent 模式标记
        item.agentEventStream = item.agentEventStream || [];
        item._eventMap = new Map();
        item._pendingToolCalls = new Map();
        
        // Check if this message has agent_steps from database (historical agent conversation)
        // If so, reconstruct the agentEventStream to restore the exact conversation state
        if (item.agent_steps && Array.isArray(item.agent_steps) && item.agent_steps.length > 0) {
            console.log('[Message Load] Reconstructing agent steps for message:', item.id, 'steps:', item.agent_steps.length);
            item.isAgentMode = true;
            item.agentEventStream = reconstructEventStreamFromSteps(item.agent_steps, item.content, item.is_completed);
            // 隐藏最终答案内容，因为它已经包含在 agentEventStream 的 answer 事件中
            item.hideContent = true;
            console.log('[Message Load] Reconstructed', item.agentEventStream.length, 'events from agent steps');
        }
        
        if (item.content) {
            if (!item.content.includes('<think>') && !item.content.includes('<\/think>')) {
                item.thinkContent = "";
                item.content = item.content;
                item.showThink = false;
                item.thinking = false;
            } else if (item.content.includes('<\/think>')) {
                // 历史消息中包含完整的 <think>...</think> 标签，说明 thinking 已完成
                const arr = item.content.trim().split('<\/think>');
                item.showThink = true;
                item.thinking = false;  // 关键：标记 thinking 已完成，使 deepThink 默认折叠
                item.thinkContent = arr[0].trim().replace('<think>', '');
                let index = item.content.trim().lastIndexOf('<\/think>')
                item.content = item.content.substring(index + 8);
            } else if (item.content.includes('<think>')) {
                // 内容包含 <think> 但没有 </think>，说明 thinking 还在进行中（不太可能出现在历史消息中）
                item.showThink = true;
                item.thinking = true;
                item.thinkContent = item.content.replace('<think>', '').trim();
                item.content = '';
            }
        }
        
        // 只给非Agent模式的空内容已完成消息设置默认错误消息
        // Agent模式的消息内容在agent_steps中，content为空是正常的
        if (item.is_completed && !item.content && !item.isAgentMode) {
            item.content = t('chat.cannotAnswer');
        }
        messagesList.unshift(item);
        if (isFirstEnter.value) {
            scrollToBottom();
        } else if (isScrollType) {
            nextTick(() => {
                const { scrollHeight } = scrollContainer.value;
                scrollContainer.value.scrollTop = scrollHeight - newScrollHeight
            })
        }
    }
    if (messagesList[messagesList.length - 1] && !messagesList[messagesList.length - 1].is_completed) {
        isReplying.value = true;
        // 保存正在 stream 的消息 ID，以便停止时使用
        const lastMessage = messagesList[messagesList.length - 1];
        if (lastMessage.role === 'assistant') {
            currentAssistantMessageId.value = lastMessage.id;
            console.log('[Continue Stream] Set assistant message ID:', lastMessage.id);
        }
        await startStream({ session_id: session_id.value, query: lastMessage.id, method: 'GET', url: '/api/v1/sessions/continue-stream' });
    }

}
const checkmenuTitle = (session_id) => {
    menuArr.value[1].children?.forEach(item => {
        if (item.id == session_id) {
            isNeedTitle.value = item.isNoTitle;
        }
    });
}
// 发送消息
// 处理停止生成事件 - 立即清除 loading 状态
const handleStopGeneration = () => {
    console.log('[Stop Generation] Immediately clearing loading state');
    loading.value = false;
    isReplying.value = false;
    // 注意：不在这里清空 currentAssistantMessageId，因为需要它来调用 API
    // API 调用成功后，后端的 stop 事件会清空它
};

const sendMsg = async (value, modelId = '', mentionedItems = []) => {
    userquery.value = value;
    isReplying.value = true;
    loading.value = true;
    // 将@提及的知识库和文件信息存入用户消息
    messagesList.push({ content: value, role: 'user', mentioned_items: mentionedItems });
    scrollToBottom();
    
    // Get agent mode status from settings store
    const agentEnabled = useSettingsStoreInstance.isAgentEnabled;
    
    // Get web search status from settings store
    const webSearchEnabled = useSettingsStoreInstance.isWebSearchEnabled;
    
    // Get knowledge_base_ids from settings store (selected by user via KnowledgeBaseSelector)
    const kbIds = useSettingsStoreInstance.settings.selectedKnowledgeBases || [];
    const knowledgeIds = useSettingsStoreInstance.settings.selectedFiles || [];
    
    // Get selected agent ID
    const selectedAgentId = useSettingsStoreInstance.selectedAgentId || '';

    
    // Use agent-chat endpoint when agent is enabled, otherwise use knowledge-chat
    const endpoint = agentEnabled ? '/api/v1/agent-chat' : '/api/v1/knowledge-chat';
    
    // Get selected MCP services from settings store (if available)
    const mcpServiceIds = useSettingsStoreInstance.settings.selectedMCPServices || [];
    
    await startStream({ 
        session_id: session_id.value, 
        knowledge_base_ids: kbIds,
        knowledge_ids: knowledgeIds,
        agent_enabled: agentEnabled,
        agent_id: selectedAgentId,  // 传递选中的智能体 ID
        web_search_enabled: webSearchEnabled,
        summary_model_id: modelId,
        mcp_service_ids: mcpServiceIds,
        mentioned_items: mentionedItems,
        query: value, 
        method: 'POST', 
        url: endpoint
    });
}

// Watch for stream errors and show message
watch(error, (newError) => {
    if (newError) {
        MessagePlugin.error(newError);
        isReplying.value = false;
        loading.value = false;
        // 清空当前 assistant message ID
        currentAssistantMessageId.value = '';
    }
});

// 处理流式数据
onChunk((data) => {
    // 日志：打印接收到的事件
    console.log('[Agent Event Received]', {
        response_type: data.response_type,
        id: data.id,
        done: data.done,
        content_length: data.content?.length || 0,
        content_preview: data.content ? data.content.substring(0, 50) : '',
        data: data.data,
        session_id: data.session_id,
        assistant_message_id: data.assistant_message_id
    });
    
    // 处理 agent query 事件 - 保存 assistant message ID 并保持 loading 状态
    if (data.response_type === 'agent_query') {
        if (data.assistant_message_id) {
            currentAssistantMessageId.value = data.assistant_message_id;
            console.log('[Agent Query] Saved assistant message ID:', data.assistant_message_id);
        }
        console.log('[Agent Query Event]', {
            session_id: data.session_id || data.data?.session_id,
            assistant_message_id: data.assistant_message_id,
            query: data.data?.query,
            request_id: data.data?.request_id
        });
        
        // 检查是否是继续流式传输（消息已存在）
        const existingMessage = messagesList.findLast((item) => item.id === data.id || item.request_id === data.id);
        if (!existingMessage) {
            // 新消息，设置 loading 状态
        loading.value = true;
            console.log('[Agent Query] New message, setting loading=true');
        } else {
            // 继续流式传输（刷新页面场景），不设置 loading，因为消息已经在列表中
            console.log('[Agent Query] Continuing stream for existing message, keeping current loading state');
        }
        return;
    }
    
    // 处理会话标题更新事件 - 不关闭 loading
    if (data.response_type === 'session_title') {
        const title = data.content || data.data?.title;
        if (title && data.data?.session_id) {
            console.log('[Session Title Update]', {
                session_id: data.data.session_id,
                title: title
            });
            usemenuStore.updatasessionTitle(data.data.session_id, title);
            usemenuStore.changeIsFirstSession(false);
            isNeedTitle.value = false;
        }
        // 不关闭 loading，等待实际内容
        return;
    }
    
    // 判断是否是 Agent 模式的响应
    // 注意：'answer', 'complete', 'references' 类型可能在两种模式下都存在
    // 只有 'thinking', 'tool_call', 'tool_result', 'reflection' 是 Agent 专有的
    const isAgentOnlyResponse = data.response_type === 'thinking' || 
                               data.response_type === 'tool_call' || 
                               data.response_type === 'tool_result' ||
                               data.response_type === 'reflection';
    
    // 检查当前消息是否已经是 Agent 模式
    const lastMessage = messagesList[messagesList.length - 1];
    const isCurrentlyAgentMode = lastMessage?.isAgentMode === true;
    
    // 如果是 Agent 专有的响应类型，或者当前消息已经是 Agent 模式，则走 Agent 处理
    const shouldHandleAsAgent = isAgentOnlyResponse || isCurrentlyAgentMode;
    
    // 处理 references 事件 - 在两种模式下都需要处理，但不改变模式
    if (data.response_type === 'references') {
        // 如果当前是 Agent 模式，走 Agent 处理
        if (isCurrentlyAgentMode) {
            handleAgentChunk(data);
            return;
        }
        // 非 Agent 模式：将 references 保存到消息中供 botmsg 使用
        let existingMessage = messagesList.findLast((item) => item.request_id === data.id || item.id === data.id);
        
        // 如果消息还不存在，先创建一个空的 assistant 消息
        if (!existingMessage) {
            existingMessage = {
                id: data.id,
                request_id: data.id,
                role: 'assistant',
                content: '',
                showThink: false,
                thinkContent: '',
                thinking: false,
                is_completed: false,
                knowledge_references: []
            };
            messagesList.push(existingMessage);
            loading.value = false; // 消息已创建，关闭 loading
            scrollToBottom();
        }
        
        existingMessage.knowledge_references = data.knowledge_references || data.data?.references || [];
        console.log('[References] Saved to message, count:', existingMessage.knowledge_references.length);
        return;
    }
    
    // Agent 模式处理（包括 stop 事件）
    if (shouldHandleAsAgent) {
        // 在 handleAgentChunk 中处理 loading 状态
        handleAgentChunk(data);
        
        // 对于 stop 事件，额外处理全局状态
        if (data.response_type === 'stop') {
            console.log('[Stop Event] Generation stopped');
            loading.value = false;
            isReplying.value = false;
            // 清空当前 assistant message ID
            currentAssistantMessageId.value = '';
        }
        return;
    }
    
    // 原有的知识库 QA 处理逻辑（非 Agent 模式）
    // answer 内容中可能包含 <think>...</think> 标签
    
    // 检查消息是否已经完成，如果已完成则忽略后续的完成事件（防止空内容覆盖）
    const existingMessage = messagesList.findLast((item) => {
        if (item.request_id === data.id) {
            return true
        }
        return item.id === data.id;
    });
    
    // 如果消息已完成且当前事件是完成事件（done=true 且无内容），直接忽略
    if (existingMessage?.is_completed && data.done && !data.content) {
        console.log('[Non-Agent] Ignoring duplicate completion event for completed message');
        return;
    }
    
    fullContent.value += data.content;
    let obj = { ...data, content: '', role: 'assistant', showThink: false, is_completed: false };

    if (fullContent.value.includes('<think>') && !fullContent.value.includes('<\/think>')) {
        obj.thinking = true;
        obj.showThink = true;
        obj.content = '';
        obj.thinkContent = fullContent.value.replace('<think>', '').trim();
    } else if (fullContent.value.includes('<think>') && fullContent.value.includes('<\/think>')) {
        obj.thinking = false;
        obj.showThink = true;
        const index = fullContent.value.indexOf('<\/think>');
        obj.thinkContent = fullContent.value.substring(0, index).replace('<think>', '').trim();
        obj.content = fullContent.value.substring(index + 8).trim();
    } else {
        obj.content = fullContent.value;
    }
    
    if (!existingMessage) {
        loading.value = false; // 消息即将创建，关闭 loading
    }
    
    if (data.done) {
        // 标记消息已完成
        obj.is_completed = true;
        // 标题生成已改为异步事件推送，不再需要在这里手动调用
        // 如果标题还未生成，前端会通过 SSE 事件接收
        isReplying.value = false;
        fullContent.value = "";
        // 清空当前 assistant message ID
        currentAssistantMessageId.value = '';
    }
    updateAssistantSession(obj);
})
// 处理 Agent 流式数据 (Cursor-style UI)
const handleAgentChunk = (data) => {
    console.log('[Agent Chunk] Processing event:', {
        response_type: data.response_type,
        id: data.id,
        done: data.done,
        has_data: !!data.data,
        data_keys: data.data ? Object.keys(data.data) : [],
        agent_event_stream_length: messagesList.findLast((item) => item.request_id === data.id || item.id === data.id)?.agentEventStream?.length || 0
    });
    
    let message = messagesList.findLast((item) => item.request_id === data.id || item.id === data.id);
    
    if (!message) {
        // 创建新的 Assistant 消息 - 此时开始显示内容，关闭 loading
        const newMsg = {
            id: data.id,
            request_id: data.id,
            role: 'assistant',
            content: '',
            isAgentMode: true,
            // Event stream: ordered list of all agent events (thinking, tool calls, etc)
            agentEventStream: [],
            // Map to track event by event_id for quick lookup
            _eventMap: new Map(),
            knowledge_references: []
        };
        messagesList.push(newMsg);
        loading.value = false; // 消息已创建，关闭 loading
        scrollToBottom();
        // Don't return - continue to process the current event data
        message = newMsg;
    }
    
    message.isAgentMode = true;
    
    // 确保在继续流式传输时（刷新页面场景），一旦接收到实际内容就关闭 loading
    // 这是一个保护措施，防止任何边缘情况导致 loading 残留
    if (loading.value && (data.response_type === 'thinking' || data.response_type === 'answer' || data.response_type === 'tool_call')) {
        console.log('[Agent Chunk] Closing loading for continued stream');
        loading.value = false;
    }
    
    switch(data.response_type) {
        case 'thinking':
            {
                const eventId = data.data?.event_id;
                console.log('[Thinking Event]', {
                    event_id: eventId,
                    done: data.done,
                    content_length: data.content?.length || 0
                });
                
                // Initialize structures
                if (!message.agentEventStream) message.agentEventStream = [];
                if (!message._eventMap) message._eventMap = new Map();
                
                if (!data.done) {
                    // Check if this thinking event already exists
                    let thinkingEvent = message._eventMap.get(eventId);
                    
                    if (!thinkingEvent) {
                        // Create new thinking event
                        console.log('[Thinking] Creating new thinking event, event_id:', eventId);
                        thinkingEvent = {
                            type: 'thinking',
                            event_id: eventId,
                            content: '',
                            done: false,
                            startTime: Date.now(),
                            thinking: true
                        };
                        
                        // Add to event stream
                        message.agentEventStream.push(thinkingEvent);
                        message._eventMap.set(eventId, thinkingEvent);
                    }
                    
                    // Accumulate content
                    if (data.content) {
                        thinkingEvent.content += data.content;
                        console.log('[Thinking] Event', eventId, 'accumulated:', thinkingEvent.content.length, 'chars');
                    }
                    
                } else {
                    // Thinking completed
                    const thinkingEvent = message._eventMap.get(eventId);
                    if (thinkingEvent) {
                        console.log('[Thinking] Completing event, event_id:', eventId, 'content length:', thinkingEvent.content.length);
                        
                        // Mark as done
                        thinkingEvent.done = true;
                        thinkingEvent.thinking = false;
                        thinkingEvent.duration_ms = data.data?.duration_ms || (Date.now() - thinkingEvent.startTime);
                        thinkingEvent.completed_at = data.data?.completed_at || Date.now();
                        
                        console.log('[Thinking] Event completed, duration:', thinkingEvent.duration_ms, 'ms');
                    } else {
                        console.warn('[Thinking] Received done for unknown event_id:', eventId);
                    }
                }
            }
            break;
            
        case 'tool_call':
            // Store or update pending tool call to pair with result later
            if (data.data && (data.data.tool_name || data.data.tool_call_id)) {
                const incomingToolName = data.data.tool_name;
                const incomingArguments = data.data.arguments;
                
                if (!message.agentEventStream) message.agentEventStream = [];
                if (!message._pendingToolCalls) message._pendingToolCalls = new Map();
                
                const toolCallId = data.data.tool_call_id || (incomingToolName ? (incomingToolName + '_' + Date.now()) : null);
                if (!toolCallId) {
                    console.warn('[Tool Call] Received event without identifiable tool_call_id:', data.data);
                    break;
                }
                
                console.log('[Tool Call]', {
                    tool_call_id: toolCallId,
                    tool_name: incomingToolName,
                    has_arguments: Boolean(incomingArguments)
                });
                
                let toolCallEvent = message._pendingToolCalls.get(toolCallId);
                if (!toolCallEvent) {
                    toolCallEvent = message.agentEventStream.find(
                        (event) => event.type === 'tool_call' && event.tool_call_id === toolCallId
                    );
                }
                
                if (toolCallEvent) {
                    if (incomingToolName) toolCallEvent.tool_name = incomingToolName;
                    if (incomingArguments) toolCallEvent.arguments = incomingArguments;
                    toolCallEvent.pending = true;
                    if (!toolCallEvent.timestamp) {
                        toolCallEvent.timestamp = Date.now();
                    }
                    message._pendingToolCalls.set(toolCallId, toolCallEvent);
                } else {
                    const newToolCallEvent = {
                        type: 'tool_call',
                        tool_call_id: toolCallId,
                        tool_name: incomingToolName,
                        arguments: incomingArguments,
                        timestamp: Date.now(),
                        pending: true
                    };
                    message.agentEventStream.push(newToolCallEvent);
                    message._pendingToolCalls.set(toolCallId, newToolCallEvent);
                }
            }
            break;
            
        case 'tool_result':
        case 'error':
            // Tool result - update the corresponding tool call event
            if (data.data) {
                const toolCallId = data.data.tool_call_id;
                const toolName = data.data.tool_name;
                const success = data.response_type !== 'error' && data.data.success !== false;
                
                console.log('[Tool Result]', {
                    tool_call_id: toolCallId,
                    tool_name: toolName,
                    success: success
                });
                
                // Find and update the pending tool call event
                let toolCallEvent = null;
                if (message._pendingToolCalls) {
                    if (toolCallId && message._pendingToolCalls.has(toolCallId)) {
                        toolCallEvent = message._pendingToolCalls.get(toolCallId);
                        message._pendingToolCalls.delete(toolCallId);
                    } else {
                        // Try to find by tool_name if no tool_call_id match
                        for (const [key, value] of message._pendingToolCalls.entries()) {
                            if (value.tool_name === toolName) {
                                toolCallEvent = value;
                                message._pendingToolCalls.delete(key);
                                break;
                            }
                        }
                    }
                }
                
                if (toolCallEvent) {
                    // Update the existing event with result
                    toolCallEvent.pending = false;
                    toolCallEvent.success = success;
                    toolCallEvent.output = success ? (data.data.output || data.content) : (data.data.error || data.content);
                    toolCallEvent.error = !success ? (data.data.error || data.content) : undefined;
                    // Set both duration and duration_ms for compatibility
                    const duration = data.data.duration_ms !== undefined ? data.data.duration_ms : data.data.duration;
                    toolCallEvent.duration = duration;
                    toolCallEvent.duration_ms = duration;
                    toolCallEvent.display_type = data.data.display_type;
                    toolCallEvent.tool_data = data.data;
                    
                    console.log('[Tool Result] Updated event in stream');
                } else {
                    console.warn('[Tool Result] No pending tool call found for', toolCallId || toolName);
                }
                
                // If this is an error response without tool data, handle it
                if (data.response_type === 'error' && !toolName) {
                    const errorMsg = data.content || t('chat.processError');
                    message.content = errorMsg;
                    isReplying.value = false;
                    loading.value = false;
                    MessagePlugin.error(errorMsg);
                    console.error('[Chat Error]', errorMsg);
                }
            } else if (data.response_type === 'error') {
                // Generic error without tool context
                const errorMsg = data.content || t('chat.processError');
                message.content = errorMsg;
                isReplying.value = false;
                loading.value = false;
                MessagePlugin.error(errorMsg);
                console.error('[Chat Error]', errorMsg);
            }
            break;
            

        case 'references':
            // 知识引用
            if (data.data?.references) {
                message.knowledge_references = data.data.references;
            } else if (data.knowledge_references) {
                // 兼容旧格式
                message.knowledge_references = data.knowledge_references;
            }
            break;
            
        case 'answer':
            // 最终答案
            message.thinking = false;
            
            console.log('[Answer Event] Received:', {
                has_content: !!data.content,
                content_length: data.content?.length || 0,
                done: data.done,
                current_message_content_length: message.content?.length || 0
            });
            
            // 只有当有实际内容时才追加，避免空内容覆盖
            if (data.content) {
                message.content = (message.content || '') + data.content;
                fullContent.value += data.content;
                console.log('[Answer] Content appended, new length:', message.content.length);
            }
            
            // Add or update answer event in agentEventStream
            if (!message.agentEventStream) message.agentEventStream = [];
            
            let answerEvent = message.agentEventStream.find((e) => e.type === 'answer');
            if (!answerEvent) {
                answerEvent = {
                    type: 'answer',
                    content: '',
                    done: false
                };
                message.agentEventStream.push(answerEvent);
                console.log('[Answer] Created new answer event in stream');
            }
            
            // 只有当有实际内容时才更新 answerEvent.content
            if (data.content) {
                answerEvent.content = message.content;
                console.log('[Answer] answerEvent.content updated, length:', answerEvent.content.length);
            }
            
            // 只在第一次收到 done:true 时标记完成，忽略后续重复的完成事件
            if (data.done && !answerEvent.done) {
                answerEvent.done = true;
                console.log('[Agent] Answer done, content length:', message.content?.length || 0, 'answerEvent.content length:', answerEvent.content?.length || 0);
                
                // 完成 - 关闭所有状态
                loading.value = false;
                isReplying.value = false;
                fullContent.value = '';
                // 清空当前 assistant message ID
                currentAssistantMessageId.value = '';
                
                // 标题生成已改为异步事件推送，不再需要在这里手动调用
                // 如果标题还未生成，前端会通过 SSE 事件接收
            } else if (data.done && answerEvent.done) {
                console.log('[Answer] Ignoring duplicate done event, current content preserved:', answerEvent.content?.length || 0);
            }
            break;
            
        case 'complete':
            // 整个流式响应完成事件 - 确保状态正确关闭
            console.log('[Agent] Complete event received');
            loading.value = false;
            isReplying.value = false;
            // 不清空 message.content，保持已有内容
            break;
            
        case 'stop':
            // 停止事件 - 添加到事件流并标记对话完成
            console.log('[Agent] Stop event received');
            if (!message.agentEventStream) message.agentEventStream = [];
            
            // Add stop event to stream
            message.agentEventStream.push({
                type: 'stop',
                timestamp: Date.now(),
                reason: data.data?.reason || 'user_requested'
            });
            
            // Mark conversation as stopped
            isReplying.value = false;
            fullContent.value = '';
            break;
            
        case 'quick_reply':
            // 快速回复选项事件 - 用于问诊等交互场景
            console.log('[Agent] Quick reply event received:', data.data);
            console.log('[Agent] Current message agentEventStream length before:', message.agentEventStream?.length || 0);
            if (!message.agentEventStream) message.agentEventStream = [];
            
            // Add quick_reply event to stream
            const quickReplyEvent = {
                type: 'quick_reply',
                event_id: data.data?.tool_call_id || `quick-reply-${Date.now()}`,
                question: data.data?.question || '',
                options: data.data?.options || [],
                multi_select: data.data?.multi_select || false,
                timestamp: Date.now()
            };
            message.agentEventStream.push(quickReplyEvent);
            console.log('[Agent] Added quick_reply event to stream, new length:', message.agentEventStream.length);
            console.log('[Agent] quick_reply event details:', quickReplyEvent);
            
            // Mark conversation as done - waiting for user selection
            console.log('[Agent] Setting loading=false and isReplying=false');
            loading.value = false;
            isReplying.value = false;
            console.log('[Agent] Loading state after:', { loading: loading.value, isReplying: isReplying.value });
            break;
    }
    
    scrollToBottom();
};

const updateAssistantSession = (payload) => {
    const message = messagesList.findLast((item) => {
        if (item.request_id === payload.id) {
            return true
        }
        return item.id === payload.id;
    });
    if (message) {
        message.content = payload.content;
        message.thinking = payload.thinking;
        message.thinkContent = payload.thinkContent;
        message.showThink = payload.showThink;
        message.knowledge_references = message.knowledge_references ? message.knowledge_references : payload.knowledge_references;
        // 更新完成状态
        if (payload.is_completed) {
            message.is_completed = true;
        }
    } else {
        messagesList.push(payload);
    }
    scrollToBottom();
}
onMounted(async () => {
    messagesList.splice(0);
    
    // 初始化状态：加载历史消息时不应显示loading
    loading.value = false;
    isReplying.value = false;
    
    // Load session data to get agent_config
    try {
        const sessionRes = await getSession(session_id.value);
        if (sessionRes?.data) {
            sessionData.value = sessionRes.data;
        }
    } catch (error) {
        console.error('Failed to load session data:', error);
    }
    
    checkmenuTitle(session_id.value)
    if (firstQuery.value) {
        scrollLock.value = true;
        sendMsg(firstQuery.value, firstModelId.value || '', firstMentionedItems.value || []);
        usemenuStore.changeFirstQuery('', [], '');
    } else {
        scrollLock.value = false;
        let data = {
            session_id: session_id.value,
            created_at: '',
            limit: limit.value
        }
        getmsgList(data)
    }
})
const clearData = () => {
    stopStream();
    isReplying.value = false;
    fullContent.value = '';
    userquery.value = '';

}
onBeforeRouteLeave((to, from, next) => {
    clearData()
    next()
})
onBeforeRouteUpdate((to, from, next) => {
    clearData()
    next()
})
</script>
<style lang="less" scoped>
.chat-container {
    display: flex;
    flex: 1;
    position: relative;
    overflow: hidden;
    
    &.sidebar-open {
        .chat {
            max-width: calc(100vw - 260px - 280px);
        }
    }
}

.chat {
    font-size: 20px;
    padding: 20px;
    box-sizing: border-box;
    flex: 1;
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    max-width: calc(100vw - 260px);
    min-width: 400px;

    :deep(.answers-input) {
        position: static;
        transform: translateX(0);

        .t-textarea__inner {
            width: 100% !important;
        }
    }
}

.chat-header {
    width: 100%;
    max-width: 800px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin: 6px 0 12px;

    .chat-header-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--td-text-color-primary);
        max-width: 520px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .export-button {
        display: inline-flex;
        align-items: center;
        gap: 6px;
    }

    .export-button-text {
        font-size: 12px;
    }
}

.chat_scroll_box {
    flex: 1;
    width: 100%;
    overflow-y: auto;

    &::-webkit-scrollbar {
        width: 0;
        height: 0;
        color: transparent;
    }
}


.agent-mode-indicator {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
    border: 1px solid #91d5ff;
    border-radius: 6px;
    margin-bottom: 12px;
    max-width: 800px;
    width: 100%;

    .agent-icon {
        font-size: 20px;
    }

    .agent-text {
        font-size: 14px;
        font-weight: 500;
        color: #0050b3;
        flex: 1;
    }
}

.msg_list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 800px;
    flex: 1;
    margin: 0 auto;
    width: 100%;

    .botanswer_laoding_gif {
        width: 24px;
        height: 18px;
        margin-left: 16px;
    }
    
    .loading-typing {
        display: flex;
        align-items: center;
        gap: 4px;
        
        span {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: #D97706;
            animation: typingBounce 1.4s ease-in-out infinite;
            
            &:nth-child(1) {
                animation-delay: 0s;
            }
            
            &:nth-child(2) {
                animation-delay: 0.2s;
            }
            
            &:nth-child(3) {
                animation-delay: 0.4s;
            }
        }
    }
}

@keyframes typingBounce {
    0%, 60%, 100% {
        transform: translateY(0);
    }
    30% {
        transform: translateY(-8px);
    }
}

.sidebar-toggle {
    position: fixed;
    right: 0;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 8px 12px;
    background: var(--td-brand-color);
    color: white;
    border-radius: 8px 0 0 8px;
    cursor: pointer;
    z-index: 100;
    font-size: 13px;
    transition: all 0.2s;
    
    &:hover {
        background: var(--td-brand-color-hover);
    }
    
    &.sidebar-open {
        right: 280px;
        padding: 8px;
    }
    
    :deep(.t-icon) {
        font-size: 16px;
    }
}
</style>