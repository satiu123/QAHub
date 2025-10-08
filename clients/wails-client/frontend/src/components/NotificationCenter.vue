<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { GetNotifications, GetUnreadCount, MarkAsRead, DeleteNotification, StartNotificationStream, StopNotificationStream, IsNotificationStreamConnected } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const props = defineProps<{
  username: string
}>()

const emit = defineEmits<{
  back: []
  viewQuestion: [questionId: number, highlightId?: string, highlightType?: string]
}>()

const notifications = ref<any[]>([])
const loading = ref(false)
const unreadCount = ref(0)
const showOnlyUnread = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const streamConnected = ref(false)

// 解析通知的 target_url
function parseTargetUrl(url: string): { questionId: number, highlightId?: string, highlightType?: string } | null {
  // target_url 格式示例: 
  // /questions/123#answer-456
  // /questions/123#comment-789
  if (!url) return null

  const match = url.match(/\/questions\/(\d+)(?:#(answer|comment)-(\w+))?/)
  if (!match) return null

  const questionId = parseInt(match[1])
  const highlightType = match[2]
  const highlightId = match[3]

  return { questionId, highlightId, highlightType }
}

// 处理通知点击
function handleNotificationClick(notification: any) {
  console.log('Notification clicked:', notification)
  console.log('Target URL:', notification.target_url)

  const parsedUrl = parseTargetUrl(notification.target_url)
  console.log('Parsed URL:', parsedUrl)

  if (!parsedUrl) {
    console.warn('Invalid target_url:', notification.target_url)
    return
  }

  // 如果未读，先标记为已读
  if (!notification.is_read) {
    handleMarkAsRead(notification.id)
  }

  // 跳转到问题详情
  console.log('Emitting viewQuestion:', parsedUrl.questionId, parsedUrl.highlightId, parsedUrl.highlightType)
  emit('viewQuestion', parsedUrl.questionId, parsedUrl.highlightId, parsedUrl.highlightType)
}

// 加载通知列表
async function loadNotifications() {
  try {
    loading.value = true
    // 参数顺序: page, pageSize, unreadOnly
    const result = await GetNotifications(currentPage.value, pageSize.value, showOnlyUnread.value)
    notifications.value = result.notifications || []
    total.value = result.total || 0
    unreadCount.value = result.unread_count || 0
  } catch (error: any) {
    console.error('加载通知失败:', error)
    alert('加载通知失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 加载未读数量
async function loadUnreadCount() {
  try {
    unreadCount.value = await GetUnreadCount()
  } catch (error: any) {
    console.error('加载未读数量失败:', error)
  }
}

// 启动通知流
async function startStream() {
  try {
    await StartNotificationStream()
    streamConnected.value = await IsNotificationStreamConnected()
    console.log('✅ 通知流已启动')
  } catch (error: any) {
    console.error('启动通知流失败:', error)
  }
}

// 处理接收到的实时通知
function handleRealtimeNotification(notification: any) {
  console.log('📨 收到实时通知:', notification)

  // 如果当前显示全部或未读通知，添加到列表顶部
  if (!showOnlyUnread.value || !notification.is_read) {
    notifications.value.unshift(notification)
    total.value++
  }

  // 更新未读数量
  if (!notification.is_read) {
    unreadCount.value++
  }

  // 显示桌面通知（可选）
  showDesktopNotification(notification)
}

// 显示桌面通知
function showDesktopNotification(notification: any) {
  const title = `来自 ${notification.sender_name || '系统'} 的通知`
  const body = notification.content

  // 这里可以用 Wails 的通知 API 或浏览器通知
  console.log(`🔔 桌面通知: ${title} - ${body}`)
}

// 标记单个通知为已读
async function handleMarkAsRead(notificationId: string) {
  try {
    await MarkAsRead([notificationId], false)
    await loadNotifications()
  } catch (error: any) {
    alert('标记失败: ' + error.toString())
  }
}

// 标记全部为已读
async function handleMarkAllAsRead() {
  if (!confirm('确定要标记全部通知为已读吗？')) {
    return
  }

  try {
    loading.value = true
    await MarkAsRead([], true)
    await loadNotifications()
    alert('已标记全部为已读')
  } catch (error: any) {
    alert('标记失败: ' + error.toString())
  } finally {
    loading.value = false
  }
}

// 删除通知
async function handleDelete(notificationId: string) {
  if (!confirm('确定要删除这条通知吗？')) {
    return
  }

  try {
    await DeleteNotification(notificationId)
    await loadNotifications()
  } catch (error: any) {
    alert('删除失败: ' + error.toString())
  }
}

// 切换筛选状态
function toggleFilter() {
  showOnlyUnread.value = !showOnlyUnread.value
  currentPage.value = 1
  loadNotifications()
}

// 获取通知图标
function getNotificationIcon(type: string): string {
  const icons: { [key: string]: string } = {
    'answer': '💬',
    'comment': '💭',
    'upvote': '👍',
    'mention': '@',
    'system': '📢',
  }
  return icons[type] || '🔔'
}

// 获取通知颜色
function getNotificationColor(type: string): string {
  const colors: { [key: string]: string } = {
    'answer': '#3498db',
    'comment': '#9b59b6',
    'upvote': '#e74c3c',
    'mention': '#f39c12',
    'system': '#95a5a6',
  }
  return colors[type] || '#34495e'
}

// 页面加载
onMounted(async () => {
  loadNotifications()
  loadUnreadCount()

  // 启动通知流
  await startStream()

  // 监听实时通知事件（使用 Wails 事件系统）
  // 注意：我们需要在后端通过 Wails runtime 发送事件
  EventsOn('notification:received', handleRealtimeNotification)
})

// 页面卸载
onUnmounted(() => {
  EventsOff('notification:received')
})
</script>

<template>
  <div class="notifications">
    <!-- 顶部导航 -->
    <div class="notifications-header">
      <button @click="emit('back')" class="btn-back">
        ← 返回
      </button>
      <h2>通知中心</h2>
    </div>

    <!-- 操作栏 -->
    <div class="actions-bar">
      <div class="stats">
        <span class="total-count">共 {{ total }} 条通知</span>
        <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount }} 条未读</span>
      </div>
      <div class="actions">
        <button @click="toggleFilter" class="btn-filter" :class="{ active: showOnlyUnread }">
          {{ showOnlyUnread ? '📋 显示全部' : '📬 仅显示未读' }}
        </button>
        <button @click="handleMarkAllAsRead" class="btn-mark-all" :disabled="unreadCount === 0">
          ✓ 全部已读
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- 通知列表 -->
    <div v-else-if="notifications.length > 0" class="notifications-list">
      <div v-for="notification in notifications" :key="notification.id" class="notification-item"
        :class="{ unread: !notification.is_read, clickable: notification.target_url }"
        @click="notification.target_url && handleNotificationClick(notification)">
        <div class="notification-icon" :style="{ backgroundColor: getNotificationColor(notification.type) }">
          {{ getNotificationIcon(notification.type) }}
        </div>
        <div class="notification-content">
          <div class="notification-header">
            <span class="sender-name">{{ notification.sender_name || '系统' }}</span>
            <span class="notification-time">{{ notification.created_at }}</span>
          </div>
          <p class="notification-text">{{ notification.content }}</p>
          <div class="notification-footer">
            <button v-if="!notification.is_read" @click.stop="handleMarkAsRead(notification.id)" class="btn-mark-read">
              标记已读
            </button>
            <button @click.stop="handleDelete(notification.id)" class="btn-delete">
              删除
            </button>
          </div>
        </div>
        <div v-if="!notification.is_read" class="unread-indicator"></div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <div class="empty-icon">🔔</div>
      <p>{{ showOnlyUnread ? '没有未读通知' : '还没有通知' }}</p>
    </div>
  </div>
</template>

<style scoped>
.notifications {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.notifications-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 24px;
}

.notifications-header h2 {
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

.actions-bar {
  background: white;
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats {
  display: flex;
  gap: 12px;
  align-items: center;
}

.total-count {
  color: #666;
  font-size: 14px;
}

.unread-badge {
  padding: 4px 12px;
  background: #e74c3c;
  color: white;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
}

.actions {
  display: flex;
  gap: 12px;
}

.btn-filter {
  padding: 8px 16px;
  background: white;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
  color: #666;
}

.btn-filter.active {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

.btn-filter:hover:not(.active) {
  border-color: #667eea;
  color: #667eea;
}

.btn-mark-all {
  padding: 8px 16px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-mark-all:hover:not(:disabled) {
  background: #5568d3;
  transform: translateY(-2px);
}

.btn-mark-all:disabled {
  background: #ccc;
  cursor: not-allowed;
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

.notifications-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.notification-item {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  gap: 16px;
  position: relative;
  transition: all 0.3s;
}

.notification-item.clickable {
  cursor: pointer;
}

.notification-item.clickable:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.notification-item.unread {
  background: #f0f7ff;
  border-left: 4px solid #667eea;
}

.notification-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.notification-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.notification-content {
  flex: 1;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.sender-name {
  font-weight: 600;
  color: #333;
  font-size: 15px;
}

.notification-time {
  font-size: 13px;
  color: #999;
}

.notification-text {
  margin: 0 0 12px 0;
  color: #555;
  line-height: 1.6;
  font-size: 14px;
}

.notification-footer {
  display: flex;
  gap: 8px;
}

.btn-mark-read,
.btn-delete {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.3s;
}

.btn-mark-read {
  background: #667eea;
  color: white;
}

.btn-mark-read:hover {
  background: #5568d3;
}

.btn-delete {
  background: #e0e0e0;
  color: #666;
}

.btn-delete:hover {
  background: #e74c3c;
  color: white;
}

.unread-indicator {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 10px;
  height: 10px;
  background: #e74c3c;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {

  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.5;
  }
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-state p {
  font-size: 16px;
  color: #999;
  margin: 0;
}
</style>
