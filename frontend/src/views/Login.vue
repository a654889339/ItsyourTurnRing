<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>ItsyourTurnRing</h1>
        <p>首饰商城管理系统</p>
      </div>

      <form @submit.prevent="handleSubmit" class="login-form">
        <div class="form-group">
          <label class="form-label">用户名</label>
          <input
            v-model="form.username"
            type="text"
            class="form-input"
            placeholder="请输入用户名"
            required
          />
        </div>

        <div class="form-group" v-if="isRegister">
          <label class="form-label">邮箱</label>
          <input
            v-model="form.email"
            type="email"
            class="form-input"
            placeholder="请输入邮箱"
          />
        </div>

        <div class="form-group">
          <label class="form-label">密码</label>
          <input
            v-model="form.password"
            type="password"
            class="form-input"
            placeholder="请输入密码"
            required
          />
        </div>

        <div class="form-group" v-if="isRegister">
          <label class="form-label">确认密码</label>
          <input
            v-model="form.confirmPassword"
            type="password"
            class="form-input"
            placeholder="请再次输入密码"
            required
          />
        </div>

        <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
          {{ loading ? '处理中...' : (isRegister ? '注册' : '登录') }}
        </button>
      </form>

      <div class="login-footer">
        <span v-if="!isRegister">
          还没有账号？
          <a href="#" @click.prevent="isRegister = true">立即注册</a>
        </span>
        <span v-else>
          已有账号？
          <a href="#" @click.prevent="isRegister = false">立即登录</a>
        </span>
      </div>

      <div v-if="error" class="error-message">{{ error }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'

const router = useRouter()
const authStore = useAuthStore()

const isRegister = ref(false)
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const handleSubmit = async () => {
  error.value = ''

  if (isRegister.value && form.password !== form.confirmPassword) {
    error.value = '两次密码输入不一致'
    return
  }

  loading.value = true

  try {
    if (isRegister.value) {
      await authStore.register({
        username: form.username,
        email: form.email,
        password: form.password
      })
    } else {
      await authStore.login(form.username, form.password)
    }
    router.push('/')
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #d4a574 0%, #8b7355 100%);
  padding: 20px;
}

.login-box {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  font-size: 28px;
  color: var(--primary-color);
  margin-bottom: 8px;
}

.login-header p {
  color: var(--text-secondary);
}

.login-form {
  margin-bottom: 20px;
}

.btn-block {
  width: 100%;
  padding: 12px;
  font-size: 16px;
}

.login-footer {
  text-align: center;
  color: var(--text-secondary);
}

.error-message {
  margin-top: 16px;
  padding: 12px;
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 4px;
  color: var(--error-color);
  text-align: center;
}
</style>
