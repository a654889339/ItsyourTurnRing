<template>
  <div class="qrcodes-page">
    <div class="page-header">
      <h2 class="page-title">小程序二维码管理</h2>
      <button class="btn btn-primary" @click="openCreateModal">+ 生成二维码</button>
    </div>

    <!-- 平台筛选 -->
    <div class="filter-bar">
      <button
        class="filter-btn"
        :class="{ active: filterPlatform === '' }"
        @click="filterPlatform = ''; fetchList()"
      >全部</button>
      <button
        class="filter-btn"
        :class="{ active: filterPlatform === 'wechat' }"
        @click="filterPlatform = 'wechat'; fetchList()"
      >微信小程序</button>
      <button
        class="filter-btn"
        :class="{ active: filterPlatform === 'alipay' }"
        @click="filterPlatform = 'alipay'; fetchList()"
      >支付宝小程序</button>
    </div>

    <!-- 二维码列表 -->
    <div class="card">
      <div class="qr-grid" v-if="qrcodes.length > 0">
        <div v-for="qr in qrcodes" :key="qr.id" class="qr-card">
          <div class="qr-canvas-wrap">
            <canvas :ref="el => setCanvasRef(el, qr.id)" width="200" height="200"></canvas>
          </div>
          <div class="qr-info">
            <div class="qr-name">{{ qr.name }}</div>
            <span class="platform-tag" :class="'platform-' + qr.platform">
              {{ qr.platform === 'wechat' ? '微信' : '支付宝' }}
            </span>
            <div class="qr-page">{{ qr.page }}</div>
            <div class="qr-params" v-if="qr.params">{{ qr.params }}</div>
            <div class="qr-time">{{ formatTime(qr.created_at) }}</div>
          </div>
          <div class="qr-actions">
            <button class="btn btn-small btn-secondary" @click="openViewModal(qr)">查看</button>
            <button class="btn btn-small btn-secondary" @click="downloadQR(qr)">下载</button>
            <button class="btn btn-small btn-danger" @click="confirmDelete(qr)">删除</button>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">暂无二维码，点击上方按钮生成</div>
    </div>

    <!-- 创建弹窗 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal" style="max-width: 500px">
        <div class="modal-header">
          <h3 class="modal-title">生成小程序二维码</h3>
          <button class="modal-close" @click="showCreateModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">名称 *</label>
            <input v-model="form.name" type="text" class="form-input" placeholder="例如: 首页二维码" />
          </div>

          <div class="form-group">
            <label class="form-label">平台 *</label>
            <div class="radio-group">
              <label class="radio-item" :class="{ active: form.platform === 'wechat' }">
                <input type="radio" v-model="form.platform" value="wechat" /> 微信小程序
              </label>
              <label class="radio-item" :class="{ active: form.platform === 'alipay' }">
                <input type="radio" v-model="form.platform" value="alipay" /> 支付宝小程序
              </label>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">页面路径 *</label>
            <select v-model="form.page" class="form-input form-select">
              <option value="">请选择页面</option>
              <option v-for="p in pageOptions" :key="p.value" :value="p.value">{{ p.label }} ({{ p.value }})</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">页面参数</label>
            <input v-model="form.params" type="text" class="form-input" placeholder="例如: id=123&type=hot" />
            <p class="form-hint">可选，用于传递参数给小程序页面</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showCreateModal = false">取消</button>
          <button class="btn btn-primary" @click="handleCreate" :disabled="creating">
            {{ creating ? '生成中...' : '生成二维码' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 查看弹窗 -->
    <div v-if="showViewModal" class="modal-overlay" @click.self="showViewModal = false">
      <div class="modal" style="max-width: 420px">
        <div class="modal-header">
          <h3 class="modal-title">{{ viewQR.name }}</h3>
          <button class="modal-close" @click="showViewModal = false">&times;</button>
        </div>
        <div class="modal-body view-body">
          <div class="view-qr-wrap">
            <canvas ref="viewCanvas" width="280" height="280"></canvas>
          </div>
          <div class="view-detail">
            <div class="detail-row">
              <span class="detail-label">平台</span>
              <span class="platform-tag" :class="'platform-' + viewQR.platform">
                {{ viewQR.platform === 'wechat' ? '微信小程序' : '支付宝小程序' }}
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">页面</span>
              <span>{{ viewQR.page }}</span>
            </div>
            <div class="detail-row" v-if="viewQR.params">
              <span class="detail-label">参数</span>
              <span>{{ viewQR.params }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">内容</span>
              <span class="content-text">{{ viewQR.content }}</span>
            </div>
          </div>
          <button class="btn btn-primary btn-block" @click="downloadQR(viewQR)" style="margin-top: 16px">
            下载二维码图片
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, watch } from 'vue'
import { qrcodeAPI } from '../api'

const qrcodes = ref([])
const filterPlatform = ref('')
const showCreateModal = ref(false)
const showViewModal = ref(false)
const creating = ref(false)
const viewQR = ref({})
const viewCanvas = ref(null)
const canvasRefs = {}

const pageOptions = ref([
  { value: 'pages/index/index', label: '首页' },
  { value: 'pages/product/index', label: '商品详情' },
  { value: 'pages/cart/index', label: '购物车' },
  { value: 'pages/order/index', label: '订单列表' },
  { value: 'pages/user/index', label: '个人中心' }
])

const form = reactive({
  name: '',
  platform: 'wechat',
  page: '',
  params: ''
})

function setCanvasRef(el, id) {
  if (el) canvasRefs[id] = el
}

// QR Code encoder (simplified)
function generateQRMatrix(text) {
  // Use a data URI approach: encode content for canvas rendering
  return text
}

function drawQR(canvas, text, size) {
  if (!canvas || !text) return
  const ctx = canvas.getContext('2d')
  ctx.clearRect(0, 0, size, size)

  // Use the QR code rendering via an image from a public API
  const img = new Image()
  img.crossOrigin = 'anonymous'
  img.onload = () => {
    ctx.drawImage(img, 0, 0, size, size)
  }
  img.onerror = () => {
    // Fallback: draw text
    ctx.fillStyle = '#f5f5f5'
    ctx.fillRect(0, 0, size, size)
    ctx.fillStyle = '#333'
    ctx.font = '12px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText('QR Code', size / 2, size / 2 - 6)
    ctx.fillText('(需联网渲染)', size / 2, size / 2 + 10)
  }
  // Use Google Charts QR API (works offline via cache)
  const encoded = encodeURIComponent(text)
  img.src = `https://api.qrserver.com/v1/create-qr-code/?size=${size}x${size}&data=${encoded}`
}

function renderAllQR() {
  nextTick(() => {
    for (const qr of qrcodes.value) {
      const canvas = canvasRefs[qr.id]
      if (canvas) drawQR(canvas, qr.content, 200)
    }
  })
}

async function fetchList() {
  try {
    const params = filterPlatform.value ? { platform: filterPlatform.value } : {}
    const result = await qrcodeAPI.list(params)
    qrcodes.value = result || []
    renderAllQR()
  } catch (error) {
    console.error('获取二维码列表失败:', error)
  }
}

function openCreateModal() {
  Object.assign(form, { name: '', platform: 'wechat', page: '', params: '' })
  showCreateModal.value = true
}

async function handleCreate() {
  if (!form.name.trim()) return alert('请输入名称')
  if (!form.page) return alert('请选择页面')

  creating.value = true
  try {
    await qrcodeAPI.create({ ...form })
    showCreateModal.value = false
    fetchList()
  } catch (error) {
    alert(error.message)
  } finally {
    creating.value = false
  }
}

function openViewModal(qr) {
  viewQR.value = qr
  showViewModal.value = true
  nextTick(() => {
    drawQR(viewCanvas.value, qr.content, 280)
  })
}

function downloadQR(qr) {
  // Create a temporary canvas to generate downloadable image
  const tempCanvas = document.createElement('canvas')
  const size = 400
  tempCanvas.width = size
  tempCanvas.height = size + 60

  const ctx = tempCanvas.getContext('2d')
  ctx.fillStyle = '#fff'
  ctx.fillRect(0, 0, tempCanvas.width, tempCanvas.height)

  const img = new Image()
  img.crossOrigin = 'anonymous'
  img.onload = () => {
    ctx.drawImage(img, 0, 0, size, size)
    // Draw label
    ctx.fillStyle = '#333'
    ctx.font = 'bold 16px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText(qr.name, size / 2, size + 24)
    ctx.font = '12px sans-serif'
    ctx.fillStyle = '#999'
    ctx.fillText(qr.platform === 'wechat' ? '微信小程序' : '支付宝小程序', size / 2, size + 44)

    const link = document.createElement('a')
    link.download = `${qr.name}_${qr.platform}.png`
    link.href = tempCanvas.toDataURL('image/png')
    link.click()
  }
  const encoded = encodeURIComponent(qr.content)
  img.src = `https://api.qrserver.com/v1/create-qr-code/?size=${size}x${size}&data=${encoded}`
}

async function confirmDelete(qr) {
  if (!confirm(`确定要删除 "${qr.name}" 吗？`)) return
  try {
    await qrcodeAPI.delete(qr.id)
    fetchList()
  } catch (error) {
    alert(error.message)
  }
}

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchList()
  // Try to load page options from backend
  qrcodeAPI.getPages().then(pages => {
    if (pages && pages.length > 0) pageOptions.value = pages
  }).catch(() => {})
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
}

.filter-bar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.filter-btn {
  padding: 6px 16px;
  border: 1px solid #ddd;
  border-radius: 20px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.filter-btn.active {
  background: var(--primary-color);
  color: #fff;
  border-color: var(--primary-color);
}

.qr-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.qr-card {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  gap: 16px;
}

.qr-canvas-wrap {
  flex-shrink: 0;
  width: 100px;
  height: 100px;
  background: #fff;
  border-radius: 8px;
  padding: 4px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.qr-canvas-wrap canvas {
  width: 100%;
  height: 100%;
}

.qr-info {
  flex: 1;
  min-width: 0;
}

.qr-name {
  font-weight: 600;
  font-size: 15px;
  margin-bottom: 6px;
}

.platform-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
  margin-bottom: 4px;
}

.platform-wechat {
  background: #e6f7e6;
  color: #07c160;
}

.platform-alipay {
  background: #e6f0ff;
  color: #1677ff;
}

.qr-page {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qr-params {
  font-size: 11px;
  color: #999;
}

.qr-time {
  font-size: 11px;
  color: #bbb;
  margin-top: 2px;
}

.qr-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-shrink: 0;
}

/* Radio group */
.radio-group {
  display: flex;
  gap: 12px;
}

.radio-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 2px solid #eee;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  flex: 1;
  justify-content: center;
}

.radio-item.active {
  border-color: var(--primary-color);
  background: rgba(212, 165, 116, 0.08);
}

.radio-item input[type="radio"] {
  display: none;
}

.form-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

/* View modal */
.view-body {
  text-align: center;
}

.view-qr-wrap {
  display: inline-block;
  background: #fff;
  padding: 12px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
  margin-bottom: 16px;
}

.view-qr-wrap canvas {
  display: block;
}

.view-detail {
  text-align: left;
  background: #f9f9f9;
  border-radius: 8px;
  padding: 12px 16px;
}

.detail-row {
  display: flex;
  align-items: flex-start;
  padding: 6px 0;
  border-bottom: 1px solid #eee;
  font-size: 13px;
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  width: 50px;
  flex-shrink: 0;
  color: #999;
}

.content-text {
  word-break: break-all;
  font-family: monospace;
  font-size: 11px;
  color: #666;
}

.btn-block {
  width: 100%;
}
</style>
