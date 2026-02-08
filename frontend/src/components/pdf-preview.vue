// @ts-nocheck
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import VuePdfEmbed from 'vue-pdf-embed';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

interface ChunkItem {
  content: string;
  chunk_index: number;
  start_at?: number;
  end_at?: number;
  char_count?: number;
  token_count?: number;
  metadata?: any;
}

const props = defineProps<{
  pdfSource: string;
  chunks: ChunkItem[];
  totalChars?: number;
}>();

const emit = defineEmits<{
  (e: 'chunk-click', chunk: ChunkItem, pageNum: number): void;
  (e: 'page-change', pageNum: number): void;
  (e: 'close'): void;
}>();

// PDF 状态
const currentPage = ref(1);
const totalPages = ref(0);
const scale = ref(1.0);
const isLoading = ref(true);
const loadError = ref<string | null>(null);

// 当前高亮的分块索引
const highlightedChunkIndex = ref<number | null>(null);
const pageInput = ref('1');

const scaleOptions = [
  { label: '50%', value: 0.5 },
  { label: '75%', value: 0.75 },
  { label: '100%', value: 1.0 },
  { label: '125%', value: 1.25 },
  { label: '150%', value: 1.5 },
  { label: '200%', value: 2.0 },
];

// vue-pdf-embed v2.x 使用 rendered 事件
const handlePdfRendered = (pdfProxy: any) => {
  console.log('[PDF Preview] Rendered event fired:', pdfProxy);
  if (pdfProxy?.numPages) {
    totalPages.value = pdfProxy.numPages;
  }
  isLoading.value = false;
  loadError.value = null;
};

// 渲染失败回调
const handlePdfError = (error: any) => {
  console.error('[PDF Preview] Rendering failed:', error);
  isLoading.value = false;
  loadError.value = t('pdfPreview.loadError') || '加载 PDF 失败';
};

// 页面跳转
const goToPage = (page: number) => {
  if (page < 1) page = 1;
  if (totalPages.value > 0 && page > totalPages.value) page = totalPages.value;
  currentPage.value = page;
  pageInput.value = String(page);
  emit('page-change', page);
};

const handlePageInputConfirm = () => {
  const page = parseInt(pageInput.value, 10);
  if (!isNaN(page)) {
    goToPage(page);
  } else {
    pageInput.value = String(currentPage.value);
  }
};

const handleZoomIn = () => {
  const idx = scaleOptions.findIndex(o => o.value === scale.value);
  if (idx < scaleOptions.length - 1) {
    scale.value = scaleOptions[idx + 1].value;
  }
};

const handleZoomOut = () => {
  const idx = scaleOptions.findIndex(o => o.value === scale.value);
  if (idx > 0) {
    scale.value = scaleOptions[idx - 1].value;
  }
};

// 根据分块的字符位置估算页码
const estimatePageForChunk = (chunk: ChunkItem): number => {
  if (!props.totalChars || props.totalChars === 0 || totalPages.value === 0) {
    const avgChunksPerPage = Math.ceil(props.chunks.length / Math.max(totalPages.value, 1));
    return Math.min(Math.ceil((chunk.chunk_index + 1) / avgChunksPerPage), Math.max(totalPages.value, 1));
  }
  const startAt = chunk.start_at ?? 0;
  const ratio = startAt / props.totalChars;
  const estimatedPage = Math.ceil(ratio * totalPages.value);
  return Math.max(1, Math.min(estimatedPage, totalPages.value));
};

const handleChunkClick = (chunk: ChunkItem, index: number) => {
  highlightedChunkIndex.value = index;
  const estimatedPage = estimatePageForChunk(chunk);
  goToPage(estimatedPage);
  emit('chunk-click', chunk, estimatedPage);
};

const getChunkMeta = (chunk: ChunkItem): string => {
  const parts: string[] = [];
  if (chunk.char_count) parts.push(`${chunk.char_count} ${t('pdfPreview.chars') || '字符'}`);
  if (chunk.token_count) parts.push(`${chunk.token_count} tokens`);
  return parts.join(' · ');
};

const getChunkPreview = (chunk: ChunkItem): string => {
  const maxLen = 100;
  const content = chunk.content || '';
  if (content.length <= maxLen) return content;
  return content.slice(0, maxLen) + '...';
};

watch(currentPage, (newVal) => {
  pageInput.value = String(newVal);
});

onMounted(() => {
  console.log('[PDF Preview] Component mounted, source:', props.pdfSource?.substring(0, 50));
});
</script>

<template>
  <div class="pdf-preview-container">
    <!-- 工具栏 -->
    <div class="pdf-toolbar">
      <div class="toolbar-left">
        <t-button theme="default" variant="text" :disabled="currentPage <= 1" @click="goToPage(currentPage - 1)">
          <template #icon><t-icon name="chevron-left" /></template>
        </t-button>
        <div class="page-nav">
          <t-input v-model="pageInput" class="page-input" size="small" @blur="handlePageInputConfirm" @keyup.enter="handlePageInputConfirm" />
          <span class="page-total">/ {{ totalPages || '?' }}</span>
        </div>
        <t-button theme="default" variant="text" :disabled="totalPages > 0 && currentPage >= totalPages" @click="goToPage(currentPage + 1)">
          <template #icon><t-icon name="chevron-right" /></template>
        </t-button>
      </div>
      <div class="toolbar-center">
        <t-button theme="default" variant="text" @click="handleZoomOut" :disabled="scale <= 0.5">
          <template #icon><t-icon name="minus-circle" /></template>
        </t-button>
        <t-select v-model="scale" class="zoom-select" size="small" :borderless="true">
          <t-option v-for="opt in scaleOptions" :key="opt.value" :value="opt.value" :label="opt.label" />
        </t-select>
        <t-button theme="default" variant="text" @click="handleZoomIn" :disabled="scale >= 2.0">
          <template #icon><t-icon name="add-circle" /></template>
        </t-button>
      </div>
      <div class="toolbar-right">
        <t-button theme="default" variant="text" @click="emit('close')">
          <template #icon><t-icon name="close" /></template>
        </t-button>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="pdf-main">
      <!-- PDF 渲染区 -->
      <div class="pdf-viewer-wrapper">
        <!-- 加载提示（覆盖层） -->
        <div v-if="isLoading" class="pdf-loading-overlay">
          <t-loading size="large" :text="t('pdfPreview.loading') || '加载中...'" />
        </div>
        <!-- 错误提示 -->
        <div v-if="loadError" class="pdf-error">
          <t-icon name="error-circle" size="48px" />
          <p>{{ loadError }}</p>
        </div>
        <!-- PDF 组件（始终渲染，否则事件不会触发） -->
        <div v-show="!loadError" class="pdf-viewer" :style="{ transform: `scale(${scale})`, transformOrigin: 'top center' }">
          <VuePdfEmbed
            :source="pdfSource"
            :page="currentPage"
            @rendered="handlePdfRendered"
            @rendering-failed="handlePdfError"
          />
        </div>
      </div>

      <!-- 分块侧边栏 -->
      <div class="chunk-sidebar" v-if="chunks && chunks.length > 0">
        <div class="sidebar-header">
          <t-icon name="view-list" size="16px" />
          <span>{{ t('pdfPreview.chunkList') || '分块列表' }}</span>
          <span class="chunk-count">({{ chunks.length }})</span>
        </div>
        <div class="chunk-list">
          <div
            v-for="(chunk, index) in chunks"
            :key="index"
            class="chunk-item"
            :class="{ 'is-highlighted': highlightedChunkIndex === index }"
            @click="handleChunkClick(chunk, index)"
          >
            <div class="chunk-header">
              <span class="chunk-index">{{ t('pdfPreview.segment') || '片段' }} {{ index + 1 }}</span>
              <span class="chunk-meta">{{ getChunkMeta(chunk) }}</span>
            </div>
            <div class="chunk-content">{{ getChunkPreview(chunk) }}</div>
            <div class="chunk-page-hint" v-if="totalPages > 0">
              <t-icon name="file" size="12px" />
              <span>{{ t('pdfPreview.estimatedPage') || '约第' }} {{ estimatePageForChunk(chunk) }} {{ t('pdfPreview.page') || '页' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.pdf-preview-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f5f5f5;
}

.pdf-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: #fff;
  border-bottom: 1px solid #e7e7e7;
  flex-shrink: 0;

  .toolbar-left, .toolbar-center, .toolbar-right {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .page-nav {
    display: flex;
    align-items: center;
    gap: 4px;
    .page-input {
      width: 50px;
      :deep(.t-input__inner) { text-align: center; }
    }
    .page-total { color: #666; font-size: 14px; }
  }
  .zoom-select { width: 80px; }
}

.pdf-main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.pdf-viewer-wrapper {
  flex: 1;
  overflow: auto;
  display: flex;
  justify-content: center;
  padding: 20px;
  background: #e0e0e0;
  position: relative;
}

.pdf-loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.9);
  z-index: 10;
}

.pdf-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #666;
  p { margin-top: 16px; }
}

.pdf-viewer {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: transform 0.2s ease;
}

.chunk-sidebar {
  width: 320px;
  background: #fff;
  border-left: 1px solid #e7e7e7;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    border-bottom: 1px solid #e7e7e7;
    font-weight: 500;
    color: #333;
    .chunk-count { color: #999; font-weight: 400; }
  }
  .chunk-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }
}

.chunk-item {
  padding: 12px;
  margin-bottom: 8px;
  border-radius: 6px;
  background: #f9f9f9;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;

  &:hover {
    background: #f0f0f0;
    border-color: #d9d9d9;
  }
  &.is-highlighted {
    background: #FEF3C7;
    border-color: #F59E0B;
    .chunk-index { color: #B45309; }
  }

  .chunk-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 6px;
    .chunk-index { font-size: 12px; font-weight: 600; color: #666; }
    .chunk-meta { font-size: 11px; color: #999; }
  }
  .chunk-content {
    font-size: 13px;
    color: #333;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .chunk-page-hint {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 8px;
    font-size: 11px;
    color: #B45309;
  }
}
</style>
