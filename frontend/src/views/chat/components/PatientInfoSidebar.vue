<template>
  <div class="patient-info-sidebar" v-if="visible && hasData">
    <div class="sidebar-header">
      <div class="sidebar-title">
        <t-icon name="user-circle" />
        <span>{{ $t('agent.patientInfo') }}</span>
      </div>
      <t-button variant="text" size="small" @click="handleClose">
        <t-icon name="close" />
      </t-button>
    </div>
    
    <div class="sidebar-content">
      <!-- 基本信息 -->
      <div class="info-section" v-if="basicInfo.length > 0">
        <div class="section-title">{{ $t('agent.basicInfo') }}</div>
        <div class="info-list">
          <div class="info-item" v-for="item in basicInfo" :key="item.key">
            <span class="info-label">{{ item.key }}</span>
            <span class="info-value">{{ item.value }}</span>
          </div>
        </div>
      </div>
      
      <!-- 主诉信息 -->
      <div class="info-section" v-if="chiefComplaint.length > 0">
        <div class="section-title">{{ $t('agent.chiefComplaint') }}</div>
        <div class="info-list">
          <div class="info-item" v-for="item in chiefComplaint" :key="item.key">
            <span class="info-label">{{ item.key }}</span>
            <span class="info-value">{{ item.value }}</span>
          </div>
        </div>
      </div>
      
      <!-- 现病史 -->
      <div class="info-section" v-if="presentIllness.length > 0">
        <div class="section-title">{{ $t('agent.presentIllness') }}</div>
        <div class="info-list">
          <div class="info-item" v-for="item in presentIllness" :key="item.key">
            <span class="info-label">{{ item.key }}</span>
            <span class="info-value">{{ item.value }}</span>
          </div>
        </div>
      </div>
      
      <!-- 既往史 -->
      <div class="info-section" v-if="pastHistory.length > 0">
        <div class="section-title">{{ $t('agent.pastHistory') }}</div>
        <div class="info-list">
          <div class="info-item" v-for="item in pastHistory" :key="item.key">
            <span class="info-label">{{ item.key }}</span>
            <span class="info-value">{{ item.value }}</span>
          </div>
        </div>
      </div>
      
      <!-- 过敏史 -->
      <div class="info-section" v-if="allergyHistory.length > 0">
        <div class="section-title">{{ $t('agent.allergyHistory') }}</div>
        <div class="info-list">
          <div class="info-item" v-for="item in allergyHistory" :key="item.key">
            <span class="info-label">{{ item.key }}</span>
            <span class="info-value">{{ item.value }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

interface InfoItem {
  key: string;
  value: string;
}

const props = defineProps<{
  collectedData: Record<string, string>;
  visible: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

// Field category mapping
const basicFields = ['姓名', '年龄', '性别', 'name', 'age', 'gender'];
const chiefFields = ['主诉', '持续时间', '严重程度', 'chief_complaint', 'duration', 'severity'];
const presentFields = ['发病时间', '诱因', '症状特点', '伴随症状', '已有治疗', 'onset_time', 'cause', 'characteristics', 'accompanying_symptoms', 'treatment_taken'];
const pastFields = ['既往疾病', '手术史', '住院史', '长期用药', 'diseases', 'surgeries', 'hospitalizations', 'medications'];
const allergyFields = ['药物过敏', '食物过敏', '其他过敏', 'drug_allergy', 'food_allergy', 'other_allergy'];

const hasData = computed(() => {
  return props.collectedData && Object.keys(props.collectedData).length > 0;
});

const categorizeData = (fields: string[]): InfoItem[] => {
  if (!props.collectedData) return [];
  
  return Object.entries(props.collectedData)
    .filter(([key]) => fields.some(f => key.toLowerCase().includes(f.toLowerCase()) || f.toLowerCase().includes(key.toLowerCase())))
    .map(([key, value]) => ({ key, value }));
};

const basicInfo = computed(() => categorizeData(basicFields));
const chiefComplaint = computed(() => categorizeData(chiefFields));
const presentIllness = computed(() => categorizeData(presentFields));
const pastHistory = computed(() => categorizeData(pastFields));
const allergyHistory = computed(() => categorizeData(allergyFields));

const handleClose = () => {
  emit('close');
};
</script>

<style lang="less" scoped>
.patient-info-sidebar {
  width: 280px;
  height: 100%;
  background: var(--td-bg-color-container);
  border-left: 1px solid var(--td-border-level-1-color);
  display: flex;
  flex-direction: column;
  
  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--td-border-level-1-color);
    
    .sidebar-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 14px;
      font-weight: 500;
      color: var(--td-text-color-primary);
      
      :deep(.t-icon) {
        font-size: 18px;
        color: var(--td-brand-color);
      }
    }
  }
  
  .sidebar-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    
    .info-section {
      margin-bottom: 16px;
      
      &:last-child {
        margin-bottom: 0;
      }
      
      .section-title {
        font-size: 12px;
        font-weight: 500;
        color: var(--td-text-color-secondary);
        margin-bottom: 8px;
        padding-bottom: 4px;
        border-bottom: 1px solid var(--td-border-level-1-color);
      }
      
      .info-list {
        .info-item {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          padding: 6px 0;
          font-size: 13px;
          
          .info-label {
            color: var(--td-text-color-secondary);
            flex-shrink: 0;
            margin-right: 8px;
          }
          
          .info-value {
            color: var(--td-text-color-primary);
            text-align: right;
            word-break: break-word;
          }
        }
      }
    }
  }
}

// Dark mode
:global(.dark) {
  .patient-info-sidebar {
    background: var(--td-bg-color-container);
  }
}
</style>
