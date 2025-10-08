<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
  search: [query: string]
}>()

const searchQuery = ref('')

function handleSearch() {
  if (searchQuery.value.trim()) {
    emit('search', searchQuery.value)
    searchQuery.value = ''
    emit('close')
  }
}

function handleClose() {
  searchQuery.value = ''
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="search-modal-overlay" @click="handleClose">
    <div class="search-modal" @click.stop>
      <div class="search-header">
        <h3>🔍 搜索问题</h3>
        <button @click="handleClose" class="btn-close">✕</button>
      </div>
      <div class="search-body">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="输入关键词搜索问题..."
          @keyup.enter="handleSearch"
          class="search-input"
          autofocus
        />
        <div class="search-tips">
          <p>💡 搜索提示：</p>
          <ul>
            <li>输入问题的关键词或标题</li>
            <li>支持模糊搜索</li>
            <li>按 Enter 键快速搜索</li>
          </ul>
        </div>
      </div>
      <div class="search-footer">
        <button @click="handleClose" class="btn-cancel">取消</button>
        <button @click="handleSearch" class="btn-search" :disabled="!searchQuery.trim()">
          搜索
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 10vh;
  z-index: 2000;
}

.search-modal {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 600px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.search-header {
  padding: 24px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-header h3 {
  margin: 0;
  font-size: 20px;
  color: #333;
}

.btn-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #999;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.3s;
}

.btn-close:hover {
  background: #f0f0f0;
  color: #333;
}

.search-body {
  padding: 24px;
}

.search-input {
  width: 100%;
  padding: 14px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 0.3s;
  box-sizing: border-box;
}

.search-input:focus {
  outline: none;
  border-color: #667eea;
}

.search-tips {
  margin-top: 20px;
  padding: 16px;
  background: #f9f9f9;
  border-radius: 8px;
  font-size: 14px;
  color: #666;
}

.search-tips p {
  margin: 0 0 8px 0;
  font-weight: 600;
  color: #333;
}

.search-tips ul {
  margin: 0;
  padding-left: 20px;
}

.search-tips li {
  margin: 4px 0;
}

.search-footer {
  padding: 20px 24px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn-cancel {
  padding: 10px 24px;
  background: #e0e0e0;
  color: #333;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-cancel:hover {
  background: #d0d0d0;
}

.btn-search {
  padding: 10px 24px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-search:hover:not(:disabled) {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-search:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>
