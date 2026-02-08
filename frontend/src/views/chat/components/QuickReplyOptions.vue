<template>
  <div class="quick-reply-options" v-if="options && options.length > 0">
    <div class="quick-reply-question" v-if="question">
      {{ question }}
    </div>
    <div class="quick-reply-buttons">
      <t-button 
        v-for="option in options" 
        :key="option.value"
        size="small"
        variant="outline"
        shape="round"
        class="quick-reply-btn"
        @click="handleSelect(option)"
      >
        {{ option.label }}
      </t-button>
    </div>
  </div>
</template>

<script setup lang="ts">
interface QuickReplyOption {
  label: string;
  value: string;
}

const props = defineProps<{
  question?: string;
  options: QuickReplyOption[];
  multiSelect?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select', value: string): void;
}>();

const handleSelect = (option: QuickReplyOption) => {
  emit('select', option.value);
};
</script>

<style lang="less" scoped>
.quick-reply-options {
  margin: 12px 0;
  padding: 12px 16px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e6f7ff 100%);
  border: 1px solid #91d5ff;
  border-radius: 8px;
  
  .quick-reply-question {
    font-size: 14px;
    color: #333;
    margin-bottom: 12px;
    font-weight: 500;
  }
  
  .quick-reply-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    
    .quick-reply-btn {
      transition: all 0.2s ease;
      border-color: #1890ff;
      color: #1890ff;
      
      &:hover {
        background: #1890ff;
        color: #fff;
        border-color: #1890ff;
        transform: translateY(-1px);
      }
      
      &:active {
        transform: translateY(0);
      }
    }
  }
}

// Dark mode support
:global(.dark) {
  .quick-reply-options {
    background: linear-gradient(135deg, #1a2744 0%, #0d1b2a 100%);
    border-color: #177ddc;
    
    .quick-reply-question {
      color: #e6f7ff;
    }
    
    .quick-reply-buttons {
      .quick-reply-btn {
        border-color: #177ddc;
        color: #69c0ff;
        
        &:hover {
          background: #177ddc;
          color: #fff;
          border-color: #177ddc;
        }
      }
    }
  }
}
</style>
