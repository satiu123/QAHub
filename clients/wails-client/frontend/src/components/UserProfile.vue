<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { GetCurrentUser, ListQuestions } from '../../wailsjs/go/main/App'

const props = defineProps<{
  username: string
}>()

const emit = defineEmits<{
  back: []
  viewQuestion: [questionId: number]
}>()

const userProfile = ref<any>(null)
const myQuestions = ref<any[]>([])
const loading = ref(false)
const activeTab = ref('profile') // 'profile' or 'questions'

// 查看问题详情
function viewQuestion(questionId: number) {
  emit('viewQuestion', questionId)
}

// 加载用户信息
async function loadUserProfile() {
  try {
    loading.value = true
    const profile = await GetCurrentUser()
    userProfile.value = profile
  } catch (error: any) {
    console.error('加载用户信息失败:', error)
    alert('加载用户信息失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 加载我的问题
async function loadMyQuestions() {
  try {
    loading.value = true
    // 加载更多问题来获取当前用户的问题
    const questions = await ListQuestions(1, 100)
    // 过滤出当前用户的问题
    myQuestions.value = questions.filter((q: any) => q.author_name === props.username)
  } catch (error: any) {
    console.error('加载我的问题失败:', error)
  } finally {
    loading.value = false
  }
}

// 切换标签页
function switchTab(tab: string) {
  activeTab.value = tab
  if (tab === 'questions' && myQuestions.value.length === 0) {
    loadMyQuestions()
  }
}

// 页面加载
onMounted(() => {
  loadUserProfile()
})
</script>

<template>
  <div class="user-profile">
    <!-- 顶部导航 -->
    <div class="profile-header">
      <button @click="emit('back')" class="btn-back">
        ← 返回首页
      </button>
      <h2>个人中心</h2>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && !userProfile" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- 用户信息 -->
    <div v-else class="profile-content">
      <!-- 用户卡片 -->
      <div class="user-card">
        <div class="user-avatar">
          <div class="avatar-circle">
            {{ username.charAt(0).toUpperCase() }}
          </div>
        </div>
        <div class="user-info">
          <h2 class="user-name">{{ username }}</h2>
          <p v-if="userProfile?.email" class="user-email">📧 {{ userProfile.email }}</p>
          <div class="user-stats">
            <div class="stat-item">
              <span class="stat-value">{{ myQuestions.length }}</span>
              <span class="stat-label">问题</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">0</span>
              <span class="stat-label">回答</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">0</span>
              <span class="stat-label">点赞</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 标签页 -->
      <div class="tabs">
        <button @click="switchTab('profile')" :class="['tab-btn', { active: activeTab === 'profile' }]">
          📋 个人信息
        </button>
        <button @click="switchTab('questions')" :class="['tab-btn', { active: activeTab === 'questions' }]">
          ❓ 我的问题
        </button>
      </div>

      <!-- 个人信息标签页 -->
      <div v-if="activeTab === 'profile'" class="tab-content">
        <div class="info-section">
          <h3>基本信息</h3>
          <div class="info-grid">
            <div class="info-item">
              <label>用户名</label>
              <div class="info-value">{{ username }}</div>
            </div>
            <div class="info-item">
              <label>邮箱</label>
              <div class="info-value">{{ userProfile?.email || '未设置' }}</div>
            </div>
            <div class="info-item">
              <label>注册时间</label>
              <div class="info-value">{{ userProfile?.created_at || '-' }}</div>
            </div>
            <div class="info-item">
              <label>用户ID</label>
              <div class="info-value">{{ userProfile?.user_id || '-' }}</div>
            </div>
          </div>
        </div>

        <div class="info-section">
          <h3>账户统计</h3>
          <div class="stats-grid">
            <div class="stats-card">
              <div class="stats-icon">❓</div>
              <div class="stats-info">
                <div class="stats-number">{{ myQuestions.length }}</div>
                <div class="stats-text">提出的问题</div>
              </div>
            </div>
            <div class="stats-card">
              <div class="stats-icon">💬</div>
              <div class="stats-info">
                <div class="stats-number">0</div>
                <div class="stats-text">发布的回答</div>
              </div>
            </div>
            <div class="stats-card">
              <div class="stats-icon">👍</div>
              <div class="stats-info">
                <div class="stats-number">0</div>
                <div class="stats-text">收到的点赞</div>
              </div>
            </div>
            <div class="stats-card">
              <div class="stats-icon">⭐</div>
              <div class="stats-info">
                <div class="stats-number">0</div>
                <div class="stats-text">获得的采纳</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 我的问题标签页 -->
      <div v-if="activeTab === 'questions'" class="tab-content">
        <div v-if="loading" class="loading-mini">
          <div class="spinner-small"></div>
          <p>加载中...</p>
        </div>

        <div v-else-if="myQuestions.length > 0" class="questions-list">
          <div v-for="question in myQuestions" :key="question.id" class="question-item"
            @click="viewQuestion(question.id)">
            <div class="question-header">
              <h4 class="question-title">{{ question.title }}</h4>
              <span class="answer-badge">{{ question.answer_count }} 回答</span>
            </div>
            <p class="question-content">{{ question.content }}</p>
            <div class="question-footer">
              <span class="question-time">🕐 {{ question.created_at }}</span>
            </div>
          </div>
        </div>

        <div v-else class="empty-state">
          <div class="empty-icon">📝</div>
          <p>你还没有提出任何问题</p>
          <button @click="emit('back')" class="btn-primary">
            去提问
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-profile {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 32px;
}

.profile-header h2 {
  margin: 0;
  font-size: 24px;
  color: #333;
}

.btn-back {
  padding: 10px 20px;
  background: white;
  border: 2px solid #667eea;
  color: #667eea;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-back:hover {
  background: #667eea;
  color: white;
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

.spinner-small {
  width: 30px;
  height: 30px;
  margin: 0 auto 12px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #667eea;
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

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.user-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  padding: 40px;
  display: flex;
  gap: 32px;
  align-items: center;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.3);
}

.user-avatar {
  flex-shrink: 0;
}

.avatar-circle {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  font-weight: bold;
  color: #667eea;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.user-info {
  flex: 1;
  color: white;
}

.user-name {
  margin: 0 0 8px 0;
  font-size: 32px;
  font-weight: bold;
}

.user-email {
  margin: 0 0 24px 0;
  font-size: 16px;
  opacity: 0.9;
}

.user-stats {
  display: flex;
  gap: 32px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

.tabs {
  display: flex;
  gap: 8px;
  background: white;
  padding: 8px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.tab-btn {
  flex: 1;
  padding: 12px 24px;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  color: #666;
  transition: all 0.3s;
}

.tab-btn.active {
  background: #667eea;
  color: white;
}

.tab-btn:hover:not(.active) {
  background: #f5f5f5;
}

.tab-content {
  background: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.info-section {
  margin-bottom: 32px;
}

.info-section:last-child {
  margin-bottom: 0;
}

.info-section h3 {
  margin: 0 0 20px 0;
  font-size: 20px;
  color: #333;
  padding-bottom: 12px;
  border-bottom: 2px solid #f0f0f0;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item label {
  font-size: 14px;
  color: #999;
  font-weight: 600;
}

.info-value {
  font-size: 16px;
  color: #333;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 8px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.stats-card {
  display: flex;
  gap: 16px;
  padding: 20px;
  background: #f9f9f9;
  border-radius: 12px;
  align-items: center;
}

.stats-icon {
  font-size: 36px;
}

.stats-info {
  flex: 1;
}

.stats-number {
  font-size: 28px;
  font-weight: bold;
  color: #667eea;
  margin-bottom: 4px;
}

.stats-text {
  font-size: 14px;
  color: #666;
}

.loading-mini {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

.questions-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.question-item {
  padding: 20px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  transition: all 0.3s;
  cursor: pointer;
}

.question-item:hover {
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

.answer-badge {
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
  font-size: 13px;
  color: #999;
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 16px;
  color: #999;
  margin-bottom: 24px;
}

.btn-primary {
  padding: 12px 32px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-primary:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}
</style>
