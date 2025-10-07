<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { Login, Register, IsLoggedIn, GetUsername, Logout } from '../wailsjs/go/main/App'
import QAHome from './components/QAHome.vue'

const isLoggedIn = ref(false)
const username = ref('')
const currentView = ref('login') // 'login' or 'register'

// 登录表单
const loginForm = ref({
  username: '',
  password: ''
})

// 注册表单
const registerForm = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const message = ref('')
const messageType = ref('') // 'success' or 'error'

// 检查登录状态
onMounted(async () => {
  try {
    isLoggedIn.value = await IsLoggedIn()
    if (isLoggedIn.value) {
      username.value = await GetUsername()
    }
  } catch (error) {
    console.error('检查登录状态失败:', error)
  }
})

// 登录
async function handleLogin() {
  try {
    message.value = ''
    const result = await Login(loginForm.value.username, loginForm.value.password)
    
    if (result.success) {
      messageType.value = 'success'
      message.value = result.message
      isLoggedIn.value = true
      username.value = loginForm.value.username
      loginForm.value = { username: '', password: '' }
    } else {
      messageType.value = 'error'
      message.value = result.message
    }
  } catch (error: any) {
    messageType.value = 'error'
    message.value = '登录失败: ' + error.toString()
  }
}

// 注册
async function handleRegister() {
  try {
    message.value = ''
    
    // 验证密码
    if (registerForm.value.password !== registerForm.value.confirmPassword) {
      messageType.value = 'error'
      message.value = '两次输入的密码不一致'
      return
    }
    
    const result = await Register(
      registerForm.value.username,
      registerForm.value.email,
      registerForm.value.password
    )
    
    if (result.success) {
      messageType.value = 'success'
      message.value = result.message
      registerForm.value = { username: '', email: '', password: '', confirmPassword: '' }
      // 注册成功后切换到登录
      setTimeout(() => {
        currentView.value = 'login'
        message.value = ''
      }, 2000)
    } else {
      messageType.value = 'error'
      message.value = result.message
    }
  } catch (error: any) {
    messageType.value = 'error'
    message.value = '注册失败: ' + error.toString()
  }
}

// 登出回调
function handleLogout() {
  isLoggedIn.value = false
  username.value = ''
  message.value = ''
}
</script>

<template>
  <div id="app">
    <!-- 已登录：显示问答主界面 -->
    <QAHome 
      v-if="isLoggedIn" 
      :username="username"
      @logout="handleLogout"
    />

    <!-- 未登录：显示登录/注册界面 -->
    <div v-else class="auth-container">
      <div class="auth-section">
        <img id="logo" alt="QAHub logo" src="./assets/images/logo-universal.png"/>
        <h1>QAHub 桌面客户端</h1>
        
        <!-- 切换登录/注册 -->
        <div class="tab-buttons">
          <button 
            @click="currentView = 'login'; message = ''" 
            :class="{ active: currentView === 'login' }"
            class="tab-btn"
          >
            登录
          </button>
          <button 
            @click="currentView = 'register'; message = ''" 
            :class="{ active: currentView === 'register' }"
            class="tab-btn"
          >
            注册
          </button>
        </div>

        <!-- 消息提示 -->
        <div v-if="message" :class="['message', messageType]">
          {{ message }}
        </div>

        <!-- 登录表单 -->
        <form v-if="currentView === 'login'" @submit.prevent="handleLogin" class="auth-form">
          <div class="form-group">
            <label>用户名</label>
            <input 
              v-model="loginForm.username" 
              type="text" 
              placeholder="请输入用户名" 
              required
            />
          </div>
          <div class="form-group">
            <label>密码</label>
            <input 
              v-model="loginForm.password" 
              type="password" 
              placeholder="请输入密码" 
              required
            />
          </div>
          <button type="submit" class="btn btn-primary">登录</button>
        </form>

        <!-- 注册表单 -->
        <form v-if="currentView === 'register'" @submit.prevent="handleRegister" class="auth-form">
          <div class="form-group">
            <label>用户名</label>
            <input 
              v-model="registerForm.username" 
              type="text" 
              placeholder="请输入用户名" 
              required
              minlength="3"
            />
          </div>
          <div class="form-group">
            <label>邮箱</label>
            <input 
              v-model="registerForm.email" 
              type="email" 
              placeholder="请输入邮箱" 
              required
            />
          </div>
          <div class="form-group">
            <label>密码</label>
            <input 
              v-model="registerForm.password" 
              type="password" 
              placeholder="请输入密码 (至少6位)" 
              required
              minlength="6"
            />
          </div>
          <div class="form-group">
            <label>确认密码</label>
            <input 
              v-model="registerForm.confirmPassword" 
              type="password" 
              placeholder="请再次输入密码" 
              required
            />
          </div>
          <button type="submit" class="btn btn-primary">注册</button>
        </form>
      </div>
    </div>
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

#app {
  min-height: 100vh;
}
</style>

<style scoped>
.auth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.auth-section {
  width: 100%;
  max-width: 450px;
  background: white;
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  margin: 20px;
}

#logo {
  display: block;
  width: 80px;
  height: 80px;
  margin: 0 auto 20px;
  border-radius: 50%;
}

h1 {
  text-align: center;
  color: #333;
  margin-bottom: 10px;
  font-size: 28px;
}

.tab-buttons {
  display: flex;
  gap: 10px;
  margin: 30px 0;
}

.tab-btn {
  flex: 1;
  padding: 12px;
  border: 2px solid #e0e0e0;
  background: white;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.3s;
  color: #666;
}

.tab-btn.active {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

.tab-btn:hover:not(.active) {
  border-color: #667eea;
  color: #667eea;
}

.message {
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 20px;
  text-align: center;
  font-size: 14px;
}

.message.success {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.message.error {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.form-group input {
  padding: 12px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 15px;
  transition: border-color 0.3s;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
}

.btn {
  padding: 14px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
}
</style>
