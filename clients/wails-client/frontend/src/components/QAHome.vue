<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ListQuestions, CreateQuestion, Logout, GetUsername, SearchQuestions, IndexAllQuestions, DeleteIndexAllQuestions, GetUnreadCount } from '../../wailsjs/go/main/App'
import QuestionDetail from './QuestionDetail.vue'
import UserProfile from './UserProfile.vue'
import NotificationCenter from './NotificationCenter.vue'

const props = defineProps<{
  username: string
}>()

const emit = defineEmits<{
  logout: []
}>()

const currentView = ref<'list' | 'detail' | 'profile' | 'notifications'>('list')
const previousView = ref<'list' | 'profile' | 'notifications'>('list') // 记录进入详情页前的视图
const selectedQuestionId = ref<number>(0)
const highlightId = ref<string | undefined>(undefined)
const highlightType = ref<string | undefined>(undefined)
const questions = ref<any[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')
const isSearchMode = ref(false)
const showAdminPanel = ref(false) // 管理面板显示状态
const unreadNotificationCount = ref(0) // 未读通知数量

// 新建问题表单
const newQuestion = ref({
  title: '',
  content: ''
})

// 加载问题列表
async function loadQuestions() {
  try {
    loading.value = true
    const result = await ListQuestions(currentPage.value, pageSize.value)
    questions.value = result || []
    isSearchMode.value = false
  } catch (error: any) {
    console.error('加载问题失败:', error)
    alert('加载问题失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 搜索问题
async function handleSearch() {
  if (!searchQuery.value.trim()) {
    loadQuestions()
    return
  }

  try {
    loading.value = true
    const result = await SearchQuestions(searchQuery.value, 50, 0)
    questions.value = result || []
    isSearchMode.value = true
  } catch (error: any) {
    console.error('搜索失败:', error)
    alert('搜索失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 清除搜索
function clearSearch() {
  searchQuery.value = ''
  loadQuestions()
}

// 创建问题
async function handleCreateQuestion() {
  if (!newQuestion.value.title || !newQuestion.value.content) {
    alert('请填写标题和内容')
    return
  }

  try {
    loading.value = true
    await CreateQuestion(newQuestion.value.title, newQuestion.value.content)
    alert('问题创建成功！')
    showCreateDialog.value = false
    newQuestion.value = { title: '', content: '' }
    // 重新加载问题列表
    await loadQuestions()
  } catch (error: any) {
    alert('创建问题失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 查看问题详情
function viewQuestion(question: any) {
  previousView.value = currentView.value as 'list' | 'profile' | 'notifications'
  selectedQuestionId.value = question.id
  highlightId.value = undefined
  highlightType.value = undefined
  currentView.value = 'detail'
}

// 从通知跳转到问题详情
function viewQuestionFromNotification(questionId: number, hId?: string, hType?: string) {
  previousView.value = 'notifications' // 从通知跳转，记录前一个视图
  selectedQuestionId.value = questionId
  highlightId.value = hId
  highlightType.value = hType
  currentView.value = 'detail'
}

// 返回上一个视图
function backToPrevious() {
  currentView.value = previousView.value
  // 清除高亮状态
  highlightId.value = undefined
  highlightType.value = undefined
  // 如果返回列表且不是搜索模式，重新加载问题列表
  if (currentView.value === 'list' && !isSearchMode.value) {
    loadQuestions()
  }
}

// 返回列表（保留用于兼容性）
function backToList() {
  previousView.value = 'list'
  currentView.value = 'list'
  // 如果是搜索模式，保持搜索结果；否则重新加载问题列表
  if (!isSearchMode.value) {
    loadQuestions()
  }
}

// 打开个人中心
function openProfile() {
  currentView.value = 'profile'
}

// 打开通知中心
function openNotifications() {
  currentView.value = 'notifications'
}

// 加载未读通知数量
async function loadUnreadCount() {
  try {
    unreadNotificationCount.value = await GetUnreadCount()
  } catch (error: any) {
    console.error('加载未读通知数量失败:', error)
  }
}

// 登出
async function handleLogout() {
  try {
    await Logout()
    emit('logout')
  } catch (error: any) {
    alert('登出失败: ' + error.toString())
  }
}

// 索引所有问题（测试用）
async function handleIndexAll() {
  if (!confirm('确定要索引所有问题吗？这将从 QA 服务获取所有问题并建立搜索索引。')) {
    return
  }

  try {
    loading.value = true
    const message = await IndexAllQuestions()
    alert(message || '索引创建成功！')
  } catch (error: any) {
    alert('索引创建失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 删除所有索引（测试用）
async function handleDeleteIndexAll() {
  if (!confirm('确定要删除所有问题索引吗？这将清空搜索索引！')) {
    return
  }

  try {
    loading.value = true
    const message = await DeleteIndexAllQuestions()
    alert(message || '索引删除成功！')
  } catch (error: any) {
    alert('索引删除失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 页面加载时获取问题列表
onMounted(() => {
  loadQuestions()
  loadUnreadCount()
  // 定期刷新未读通知数量
  setInterval(loadUnreadCount, 30000) // 每30秒刷新一次
})
</script>

<template>
  <div class="qa-home">
    <!-- 问题详情页 -->
    <QuestionDetail v-if="currentView === 'detail'" :question-id="selectedQuestionId" :username="props.username"
      :highlight-id="highlightId" :highlight-type="highlightType" @back="backToPrevious" />

    <!-- 个人中心 -->
    <UserProfile v-else-if="currentView === 'profile'" :username="props.username" @back="backToList" />

    <!-- 通知中心 -->
    <NotificationCenter v-else-if="currentView === 'notifications'" :username="props.username" @back="backToList"
      @viewQuestion="viewQuestionFromNotification" />

    <!-- 问题列表页 -->
    <div v-else>
      <!-- 顶部导航栏 -->
      <header class="header">
        <div class="header-content">
          <h1 class="logo">🎓 QAHub</h1>
          <div class="header-right">
            <button @click="openNotifications" class="btn-notifications" title="通知中心">
              🔔
              <span v-if="unreadNotificationCount > 0" class="notification-badge">
                {{ unreadNotificationCount > 99 ? '99+' : unreadNotificationCount }}
              </span>
            </button>
            <button @click="openProfile" class="btn-profile">
              👤 {{ props.username }}
            </button>
            <button @click="showAdminPanel = !showAdminPanel" class="btn-admin" title="管理面板">
              ⚙️
            </button>
            <button @click="handleLogout" class="btn-logout">登出</button>
          </div>
        </div>
      </header>

      <!-- 管理面板 -->
      <div v-if="showAdminPanel" class="admin-panel">
        <div class="admin-content">
          <h3>🔧 索引管理 (测试功能)</h3>
          <p class="admin-desc">用于管理 Elasticsearch 搜索索引</p>
          <div class="admin-actions">
            <button @click="handleIndexAll" class="btn-admin-action btn-index" :disabled="loading">
              📥 索引所有问题
            </button>
            <button @click="handleDeleteIndexAll" class="btn-admin-action btn-delete" :disabled="loading">
              🗑️ 删除所有索引
            </button>
          </div>
        </div>
      </div>

      <!-- 主内容区 -->
      <main class="main-content">
        <div class="container">
          <!-- 搜索栏 -->
          <div class="search-bar">
            <div class="search-input-wrapper">
              <input v-model="searchQuery" type="text" placeholder="🔍 搜索问题..." @keyup.enter="handleSearch"
                class="search-input" />
              <button v-if="searchQuery" @click="clearSearch" class="btn-clear" title="清除搜索">
                ✕
              </button>
              <button @click="handleSearch" class="btn-search">
                搜索
              </button>
            </div>
          </div>

          <!-- 操作栏 -->
          <div class="action-bar">
            <h2>{{ isSearchMode ? `搜索结果 (${questions.length})` : '问题列表' }}</h2>
            <button @click="showCreateDialog = true" class="btn-primary">
              ➕ 提问
            </button>
          </div>

          <!-- 加载状态 -->
          <div v-if="loading" class="loading">
            <div class="spinner"></div>
            <p>加载中...</p>
          </div>

          <!-- 问题列表 -->
          <div v-else-if="questions.length > 0" class="question-list">
            <div v-for="question in questions" :key="question.id" class="question-card" @click="viewQuestion(question)">
              <div class="question-header">
                <h3 class="question-title">{{ question.title }}</h3>
                <span v-if="!isSearchMode" class="answer-count">{{ question.answer_count }} 回答</span>
              </div>
              <p class="question-content">{{ question.content }}</p>
              <div class="question-footer">
                <span class="author">👤 {{ question.author_name }}</span>
                <span class="time">🕐 {{ question.created_at }}</span>
              </div>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-else class="empty-state">
            <p>📝 还没有问题，来提第一个问题吧！</p>
            <button @click="showCreateDialog = true" class="btn-primary">
              立即提问
            </button>
          </div>
        </div>
      </main>

      <!-- 创建问题对话框 -->
      <div v-if="showCreateDialog" class="modal-overlay" @click="showCreateDialog = false">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3>提出问题</h3>
            <button @click="showCreateDialog = false" class="btn-close">✕</button>
          </div>
          <form @submit.prevent="handleCreateQuestion" class="question-form">
            <div class="form-group">
              <label>标题</label>
              <input v-model="newQuestion.title" type="text" placeholder="请输入问题标题" required maxlength="200" />
            </div>
            <div class="form-group">
              <label>详细描述</label>
              <textarea v-model="newQuestion.content" placeholder="请详细描述你的问题..." required rows="8"></textarea>
            </div>
            <div class="form-actions">
              <button type="button" @click="showCreateDialog = false" class="btn-secondary">
                取消
              </button>
              <button type="submit" class="btn-primary" :disabled="loading">
                {{ loading ? '提交中...' : '提交问题' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.qa-home {
  min-height: 100vh;
  background: #f5f5f5;
}

.header {
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-size: 24px;
  font-weight: bold;
  color: #667eea;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.btn-profile {
  padding: 8px 16px;
  background: #f0f0f0;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s;
  color: #333;
}

.btn-profile:hover {
  background: #e0e0e0;
}

.btn-admin {
  padding: 8px 12px;
  background: #f0f0f0;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.3s;
}

.btn-admin:hover {
  background: #667eea;
  transform: scale(1.1);
}

.username {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.btn-logout {
  padding: 8px 16px;
  background-color: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.btn-notifications {
  position: relative;
  padding: 8px 12px;
  background-color: #6c757d;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
  margin-right: 10px;
  font-size: 18px;
}

.btn-notifications:hover {
  background-color: #5a6268;
}

.notification-badge {
  position: absolute;
  top: -5px;
  right: -5px;
  background-color: #dc3545;
  color: white;
  border-radius: 10px;
  padding: 2px 6px;
  font-size: 12px;
  font-weight: bold;
  min-width: 20px;
  text-align: center;
}

.btn-logout:hover {
  background-color: #c82333;
}

.admin-panel {
  max-width: 1200px;
  margin: -12px auto 20px;
  padding: 0 20px;
}

.admin-content {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
  color: white;
}

.admin-content h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
}

.admin-desc {
  margin: 0 0 16px 0;
  font-size: 14px;
  opacity: 0.9;
}

.admin-actions {
  display: flex;
  gap: 12px;
}

.btn-admin-action {
  padding: 10px 20px;
  border: 2px solid white;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s;
  background: transparent;
  color: white;
}

.btn-admin-action:hover:not(:disabled) {
  background: white;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.btn-index:hover:not(:disabled) {
  color: #667eea;
}

.btn-delete:hover:not(:disabled) {
  color: #e74c3c;
}

.btn-admin-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.main-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 32px 20px;
}

.container {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.search-bar {
  margin-bottom: 24px;
}

.search-input-wrapper {
  display: flex;
  gap: 8px;
  align-items: center;
}

.search-input {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 15px;
  transition: border-color 0.3s;
}

.search-input:focus {
  outline: none;
  border-color: #667eea;
}

.btn-clear {
  padding: 8px 12px;
  background: #f0f0f0;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.3s;
}

.btn-clear:hover {
  background: #e0e0e0;
  color: #333;
}

.btn-search {
  padding: 12px 24px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-search:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid #f0f0f0;
}

.action-bar h2 {
  margin: 0;
  color: #333;
  font-size: 20px;
}

.btn-primary {
  padding: 10px 20px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-primary:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}

.loading {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 16px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

.question-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.question-card {
  padding: 20px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.question-card:hover {
  border-color: #667eea;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
  transform: translateY(-2px);
}

.question-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.question-title {
  margin: 0;
  font-size: 18px;
  color: #333;
  flex: 1;
}

.answer-count {
  padding: 4px 12px;
  background: #e8f4f8;
  color: #0288d1;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 500;
}

.question-content {
  color: #666;
  margin: 0 0 12px 0;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.question-footer {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #999;
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
}

.empty-state p {
  font-size: 16px;
  color: #999;
  margin-bottom: 24px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.modal-header {
  padding: 24px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
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

.question-form {
  padding: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 15px;
  font-family: inherit;
  transition: border-color 0.3s;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #667eea;
}

.form-group textarea {
  resize: vertical;
  min-height: 120px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.btn-secondary {
  padding: 10px 20px;
  background: #e0e0e0;
  color: #333;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-secondary:hover {
  background: #d0d0d0;
}
</style>
