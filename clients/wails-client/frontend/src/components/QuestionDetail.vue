<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import {
  GetQuestion,
  ListAnswers,
  CreateAnswer,
  UpvoteAnswer,
  DownvoteAnswer,
  ListComments,
  CreateComment
} from '../../wailsjs/go/main/App'

const props = defineProps<{
  questionId: number
  username: string
  highlightId?: string
  highlightType?: string
}>()

const emit = defineEmits<{
  back: []
}>()

const question = ref<any>(null)
const answers = ref<any[]>([])
const loading = ref(false)
const answerContent = ref('')
const commentContent = ref<{ [key: number]: string }>({})
const showComments = ref<{ [key: number]: boolean }>({})
const comments = ref<{ [key: number]: any[] }>({})
const loadingComments = ref<{ [key: number]: boolean }>({})

// 添加滚动到高亮元素的函数
function scrollToHighlight(retry = 0) {
  if (!props.highlightId || !props.highlightType) {
    console.log('No highlight needed')
    return
  }

  console.log('Attempting to highlight (retry:', retry, '):', props.highlightType, props.highlightId)

  const elementId = `${props.highlightType}-${props.highlightId}`
  const element = document.getElementById(elementId) as HTMLElement

  if (element) {
    console.log('Element found, scrolling...')

    // 先添加高亮样式
    element.classList.add('highlight-flash')

    // 获取元素位置并滚动
    const rect = element.getBoundingClientRect()
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop
    const targetY = rect.top + scrollTop - (window.innerHeight / 2) + (rect.height / 2)

    // 立即滚动
    window.scrollTo({
      top: Math.max(0, targetY),
      behavior: 'smooth'
    })

    // 3秒后移除高亮
    setTimeout(() => element.classList.remove('highlight-flash'), 3000)
  } else if (retry < 5) {
    // 如果元素还没渲染，短时间后重试
    console.log('Element not found, will retry...')
    setTimeout(() => scrollToHighlight(retry + 1), 200)
  } else {
    console.warn('Element not found after retries:', elementId)
  }
}

// 页面加载
onMounted(async () => {
  console.log('QuestionDetail mounted with props:', {
    questionId: props.questionId,
    highlightId: props.highlightId,
    highlightType: props.highlightType
  })

  // 先启动数据加载（不等待）
  const loadPromise = Promise.all([loadQuestion(), loadAnswers()])

  // 如果是评论高亮，先展开评论
  if (props.highlightId && props.highlightType === 'comment') {
    const answerId = parseInt(props.highlightId.split('-')[0] || '0')
    console.log('Need to expand comments for answer:', answerId)

    if (answerId) {
      // 等待回答加载完成
      await loadPromise
      await nextTick()

      // 展开评论
      toggleComments(answerId)
      await nextTick()

      // 立即尝试滚动
      setTimeout(() => scrollToHighlight(), 100)
    }
  } else if (props.highlightId && props.highlightType === 'answer') {
    // 对于回答的高亮，数据加载时就开始尝试滚动
    await loadPromise
    await nextTick()

    // 立即尝试滚动
    setTimeout(() => scrollToHighlight(), 100)
  } else {
    // 没有高亮，正常等待加载完成
    await loadPromise
  }
})
// 加载问题详情
async function loadQuestion() {
  try {
    loading.value = true
    question.value = await GetQuestion(props.questionId)
  } catch (error: any) {
    console.error('加载问题详情失败:', error)
    alert('加载问题详情失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 加载回答列表
async function loadAnswers() {
  try {
    loading.value = true
    const result = await ListAnswers(props.questionId, 1, 50)
    answers.value = result || []
  } catch (error: any) {
    console.error('加载回答失败:', error)
  } finally {
    loading.value = false
  }
}

// 提交回答
async function handleSubmitAnswer() {
  if (!answerContent.value.trim()) {
    alert('请输入回答内容')
    return
  }

  try {
    loading.value = true
    await CreateAnswer(props.questionId, answerContent.value)
    alert('回答提交成功！')
    answerContent.value = ''
    await loadAnswers()
  } catch (error: any) {
    alert('提交回答失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 点赞回答
async function handleUpvote(answerId: number) {
  try {
    await UpvoteAnswer(answerId)
    await loadAnswers()
  } catch (error: any) {
    alert('点赞失败: ' + error.toString())
  }
}

// 取消点赞
async function handleDownvote(answerId: number) {
  try {
    await DownvoteAnswer(answerId)
    await loadAnswers()
  } catch (error: any) {
    alert('取消点赞失败: ' + error.toString())
  }
}

// 加载评论
async function loadComments(answerId: number) {
  try {
    loadingComments.value[answerId] = true
    const result = await ListComments(answerId, 1, 50)
    comments.value[answerId] = result || []
    showComments.value[answerId] = true
  } catch (error: any) {
    alert('加载评论失败: ' + error.toString())
  } finally {
    loadingComments.value[answerId] = false
  }
}

// 切换评论显示
function toggleComments(answerId: number) {
  if (showComments.value[answerId]) {
    showComments.value[answerId] = false
  } else {
    if (!comments.value[answerId]) {
      loadComments(answerId)
    } else {
      showComments.value[answerId] = true
    }
  }
}

// 提交评论
async function handleSubmitComment(answerId: number) {
  const content = commentContent.value[answerId]
  if (!content || !content.trim()) {
    alert('请输入评论内容')
    return
  }

  try {
    await CreateComment(answerId, content)
    commentContent.value[answerId] = ''
    await loadComments(answerId)
  } catch (error: any) {
    alert('提交评论失败: ' + error.toString())
  }
}
</script>

<template>
  <div class="question-detail">
    <!-- 顶部导航 -->
    <div class="detail-header">
      <button @click="emit('back')" class="btn-back">
        ← 返回列表
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && !question" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- 问题详情 -->
    <div v-else-if="question" class="detail-content">
      <!-- 问题卡片 -->
      <div class="question-card">
        <h1 class="question-title">{{ question.title }}</h1>
        <div class="question-meta">
          <span class="author">👤 {{ question.author_name }}</span>
          <span class="time">🕐 {{ question.created_at }}</span>
          <span class="answer-count">💬 {{ question.answer_count }} 个回答</span>
        </div>
        <div class="question-content">
          {{ question.content }}
        </div>
      </div>

      <!-- 回答区域 -->
      <div class="answers-section">
        <h2 class="section-title">全部回答 ({{ answers.length }})</h2>

        <!-- 回答列表 -->
        <div v-if="answers.length > 0" class="answers-list">
          <div v-for="answer in answers" :key="answer.id" :id="`answer-${answer.id}`" class="answer-card">
            <div class="answer-header">
              <span class="answer-author">👤 {{ answer.username }}</span>
              <span class="answer-time">{{ answer.created_at }}</span>
            </div>
            <div class="answer-content">
              {{ answer.content }}
            </div>
            <div class="answer-footer">
              <button @click="answer.is_upvoted ? handleDownvote(answer.id) : handleUpvote(answer.id)"
                :class="['btn-vote', { active: answer.is_upvoted }]">
                {{ answer.is_upvoted ? '👍 已赞' : '👍 点赞' }} ({{ answer.upvote_count }})
              </button>
              <button @click="toggleComments(answer.id)" class="btn-comment">
                💬 {{ showComments[answer.id] ? '收起评论' : '评论' }}
              </button>
            </div>

            <!-- 评论区 -->
            <div v-if="showComments[answer.id]" class="comments-section">
              <div v-if="loadingComments[answer.id]" class="loading-mini">
                加载评论中...
              </div>
              <div v-else>
                <!-- 评论列表 -->
                <div v-if="comments[answer.id]?.length > 0" class="comments-list">
                  <div v-for="comment in comments[answer.id]" :key="comment.id" :id="`comment-${comment.id}`"
                    class="comment-item">
                    <div class="comment-header">
                      <span class="comment-author">{{ comment.username }}</span>
                      <span class="comment-time">{{ comment.created_at }}</span>
                    </div>
                    <div class="comment-content">{{ comment.content }}</div>
                  </div>
                </div>
                <div v-else class="no-comments">
                  暂无评论
                </div>

                <!-- 添加评论 -->
                <div class="comment-input">
                  <input v-model="commentContent[answer.id]" type="text" placeholder="写下你的评论..."
                    @keyup.enter="handleSubmitComment(answer.id)" />
                  <button @click="handleSubmitComment(answer.id)" class="btn-submit-comment">
                    发送
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-else class="empty-answers">
          <p>还没有回答，来抢沙发吧！</p>
        </div>
      </div>

      <!-- 回答输入区 -->
      <div class="answer-input-section">
        <h3>写下你的回答</h3>
        <textarea v-model="answerContent" placeholder="分享你的见解..." rows="6"></textarea>
        <div class="input-actions">
          <button @click="handleSubmitAnswer" class="btn-submit" :disabled="loading || !answerContent.trim()">
            {{ loading ? '提交中...' : '提交回答' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.question-detail {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.detail-header {
  margin-bottom: 20px;
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

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

@keyframes highlight {

  0%,
  100% {
    background-color: transparent;
  }

  50% {
    background-color: #fff3cd;
  }
}

.highlight-flash {
  animation: highlight 1s ease-in-out 3;
  border-left: 4px solid #ffc107 !important;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.question-card {
  background: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.question-title {
  margin: 0 0 16px 0;
  font-size: 28px;
  color: #333;
  line-height: 1.4;
}

.question-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e0e0e0;
  font-size: 14px;
  color: #666;
}

.question-content {
  font-size: 16px;
  line-height: 1.8;
  color: #333;
  white-space: pre-wrap;
}

.answers-section {
  background: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.section-title {
  margin: 0 0 24px 0;
  font-size: 20px;
  color: #333;
  padding-bottom: 12px;
  border-bottom: 2px solid #f0f0f0;
}

.answers-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.answer-card {
  padding: 20px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #fafafa;
}

.answer-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
  font-size: 14px;
  color: #666;
}

.answer-author {
  font-weight: 600;
  color: #333;
}

.answer-content {
  font-size: 15px;
  line-height: 1.7;
  color: #333;
  margin-bottom: 16px;
  white-space: pre-wrap;
}

.answer-footer {
  display: flex;
  gap: 12px;
}

.btn-vote {
  padding: 8px 16px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
  color: #666;
}

.btn-vote:hover {
  border-color: #667eea;
  color: #667eea;
}

.btn-vote.active {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

.btn-comment {
  padding: 8px 16px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
  color: #666;
}

.btn-comment:hover {
  border-color: #667eea;
  color: #667eea;
}

.comments-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed #e0e0e0;
}

.loading-mini {
  text-align: center;
  padding: 20px;
  color: #999;
  font-size: 14px;
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.comment-item {
  padding: 12px;
  background: white;
  border-radius: 6px;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
}

.comment-author {
  font-weight: 600;
  color: #333;
}

.comment-time {
  color: #999;
}

.comment-content {
  font-size: 14px;
  color: #555;
  line-height: 1.6;
}

.no-comments {
  text-align: center;
  padding: 20px;
  color: #999;
  font-size: 14px;
}

.comment-input {
  display: flex;
  gap: 8px;
}

.comment-input input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  font-size: 14px;
}

.comment-input input:focus {
  outline: none;
  border-color: #667eea;
}

.btn-submit-comment {
  padding: 10px 20px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-submit-comment:hover {
  background: #5568d3;
}

.empty-answers {
  text-align: center;
  padding: 40px 20px;
  color: #999;
}

.answer-input-section {
  background: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.answer-input-section h3 {
  margin: 0 0 16px 0;
  font-size: 18px;
  color: #333;
}

.answer-input-section textarea {
  width: 100%;
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 15px;
  font-family: inherit;
  resize: vertical;
  box-sizing: border-box;
  transition: border-color 0.3s;
}

.answer-input-section textarea:focus {
  outline: none;
  border-color: #667eea;
}

.input-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.btn-submit {
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

.btn-submit:hover:not(:disabled) {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-submit:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}
</style>
