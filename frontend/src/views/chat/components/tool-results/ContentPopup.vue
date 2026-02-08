<template>
  <div class="popup-content">
    <div class="popup-content-wrapper">
      <div v-if="content" class="full-content" :class="{ 'html-content': isHtml }">
        <div v-if="isHtml" v-html="processedContent"></div>
        <template v-else>{{ content }}</template>
      </div>
    </div>
    <div class="popup-footer">
      <div v-if="hasInfo" class="info-section">
        <div v-if="chunkId" class="info-field">
          <span class="field-label">{{ $t('chat.chunkIdLabel') }}</span>
          <span class="field-value"><code>{{ chunkId }}</code></span>
        </div>
        <div v-if="knowledgeId" class="info-field">
          <span class="field-label">{{ $t('chat.documentIdLabel') }}</span>
          <span class="field-value"><code>{{ knowledgeId }}</code></span>
        </div>
      </div>
      <button 
        v-if="chunkId" 
        class="view-detail-btn" 
        @click="handleViewDetail"
        :disabled="isNavigating"
      >
        <t-icon name="jump" size="12px" />
        {{ isNavigating ? $t('common.loading') : ($t('chat.viewChunkDetail') || '查看详情') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { sanitizeHTML } from '@/utils/security';
import { getChunkByIdOnly } from '@/api/knowledge-base';
import { MessagePlugin } from 'tdesign-vue-next';

interface Props {
  content?: string;
  chunkId?: string;
  knowledgeId?: string;
  isHtml?: boolean; // 是否以 HTML 格式显示内容
}

const props = defineProps<Props>();
const router = useRouter();
const { t } = useI18n();

const isNavigating = ref(false);

const hasInfo = computed(() => {
  return !!(props.chunkId || props.knowledgeId);
});

// 处理 HTML 内容
const processedContent = computed(() => {
  if (!props.content) return '';
  if (props.isHtml) {
    return sanitizeHTML(props.content);
  }
  return props.content;
});

// 跳转到分块详情
const handleViewDetail = async () => {
  if (!props.chunkId || isNavigating.value) return;
  
  isNavigating.value = true;
  try {
    // 通过 API 获取 knowledge_base_id
    const response = await getChunkByIdOnly(props.chunkId);
    const kbId = response.data?.knowledge_base_id;
    
    if (!kbId) {
      MessagePlugin.warning(t('chat.noKnowledgeBaseId') || '无法获取知识库ID');
      return;
    }
    
    const query: Record<string, string> = {};
    if (props.chunkId) query.chunk = props.chunkId;
    if (props.knowledgeId) query.knowledge = props.knowledgeId;
    
    router.push({
      path: `/platform/knowledge-bases/${kbId}`,
      query
    });
  } catch (error) {
    console.error('Failed to get chunk info:', error);
    MessagePlugin.error(t('common.operationFailed') || '操作失败');
  } finally {
    isNavigating.value = false;
  }
};
</script>

<style lang="less" scoped>
.popup-content {
  display: flex;
  flex-direction: column;
  max-height: 400px;
  max-width: 500px;
  border: 1px solid #D97706;
  border-radius: 4px;
  word-wrap: break-word;
  word-break: break-word;
  overflow: hidden;
  
  .popup-content-wrapper {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 12px;
    min-height: 0;
  }
  
  .full-content {
    font-size: 13px;
    color: #333;
    line-height: 1.8;
    white-space: pre-wrap;
    word-break: break-word;
    
    &.html-content {
      white-space: normal;
      
      :deep(p) {
        margin: 8px 0;
        line-height: 1.8;
      }
      
      :deep(br) {
        line-height: 1.8;
      }
    }
  }
  
  .popup-footer {
    flex-shrink: 0;
    padding: 8px 12px;
    border-top: 1px solid #e7e7e7;
    background: #fafafa;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  
  .info-section {
    flex: 1;
  }
  
  .info-field {
    display: flex;
    gap: 8px;
    margin-bottom: 4px;
    font-size: 11px;
    
    &:last-child {
      margin-bottom: 0;
    }
    
    .field-label {
      color: #8b8b8b;
      min-width: 60px;
      flex-shrink: 0;
    }
    
    .field-value {
      color: #666;
      flex: 1;
      
      code {
        font-family: 'Monaco', 'Courier New', monospace;
        font-size: 10px;
        background: #f0f0f0;
        padding: 1px 4px;
        border-radius: 2px;
      }
    }
  }
  
  .view-detail-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    font-size: 12px;
    font-weight: 500;
    color: #D97706;
    background: #FFFBEB;
    border: 1px solid #FDE68A;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s ease;
    white-space: nowrap;
    flex-shrink: 0;
    
    &:hover:not(:disabled) {
      background: #FEF3C7;
      border-color: #D97706;
    }
    
    &:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }
  }
}
</style>

