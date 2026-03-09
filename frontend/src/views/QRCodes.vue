<template>
  <div class="qrcodes-page">
    <div class="page-header">
      <h2 class="page-title">小程序二维码管理</h2>
      <button class="btn btn-primary" @click="openCreateModal">+ 生成二维码</button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <button class="filter-btn" :class="{ active: filterPlatform === '' }" @click="filterPlatform = ''; fetchList()">全部</button>
      <button class="filter-btn" :class="{ active: filterPlatform === 'wechat' }" @click="filterPlatform = 'wechat'; fetchList()">微信小程序</button>
      <button class="filter-btn" :class="{ active: filterPlatform === 'alipay' }" @click="filterPlatform = 'alipay'; fetchList()">支付宝小程序</button>
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
            <div class="qr-tags">
              <span class="platform-tag" :class="'platform-' + qr.platform">
                {{ qr.platform === 'wechat' ? '微信' : '支付宝' }}
              </span>
              <span class="scene-tag" :class="'scene-' + qr.scene">{{ sceneLabel(qr.scene) }}</span>
            </div>
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
      <div class="modal" style="max-width: 520px">
        <div class="modal-header">
          <h3 class="modal-title">生成小程序二维码</h3>
          <button class="modal-close" @click="showCreateModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <!-- 场景选择 -->
          <div class="form-group">
            <label class="form-label">使用场景 *</label>
            <div class="scene-grid">
              <div
                v-for="s in sceneOptions"
                :key="s.value"
                class="scene-card"
                :class="{ active: form.scene === s.value }"
                @click="form.scene = s.value; onSceneChange()"
              >
                <span class="scene-icon">{{ s.icon }}</span>
                <span class="scene-name">{{ s.label }}</span>
              </div>
            </div>
          </div>

          <!-- 名称 -->
          <div class="form-group">
            <label class="form-label">名称 *</label>
            <input v-model="form.name" type="text" class="form-input" :placeholder="namePlaceholder" />
          </div>

          <!-- 平台 -->
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

          <!-- 商品选择 (商品查看/下单) -->
          <div class="form-group" v-if="form.scene === 'product_view' || form.scene === 'product_buy'">
            <label class="form-label">选择商品 *</label>
            <div class="product-selector">
              <input
                v-model="productSearch"
                type="text"
                class="form-input"
                placeholder="搜索商品名称..."
                @input="filterProducts"
              />
              <div class="product-list" v-if="filteredProducts.length > 0">
                <div
                  v-for="p in filteredProducts"
                  :key="p.id"
                  class="product-option"
                  :class="{ selected: form.product_id === p.id }"
                  @click="selectProduct(p)"
                >
                  <img v-if="p.main_image" :src="p.main_image" class="product-thumb" />
                  <div class="product-thumb placeholder" v-else>无图</div>
                  <div class="product-detail">
                    <div class="product-title">{{ p.name }}</div>
                    <div class="product-price">¥{{ p.price }}</div>
                  </div>
                  <span class="check-mark" v-if="form.product_id === p.id">✓</span>
                </div>
              </div>
              <div class="selected-product" v-if="selectedProduct">
                已选: <strong>{{ selectedProduct.name }}</strong> (ID: {{ selectedProduct.id }})
              </div>
            </div>
          </div>

          <!-- 订单号 (订单状态) -->
          <div class="form-group" v-if="form.scene === 'order_status'">
            <label class="form-label">订单号 *</label>
            <input v-model="form.order_no" type="text" class="form-input" placeholder="请输入订单号" />
          </div>

          <!-- 自定义页面 -->
          <div v-if="form.scene === 'custom'">
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
            </div>
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
              <span class="detail-label">场景</span>
              <span class="scene-tag" :class="'scene-' + viewQR.scene">{{ sceneLabel(viewQR.scene) }}</span>
            </div>
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
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { qrcodeAPI, productAPI } from '../api'

const qrcodes = ref([])
const filterPlatform = ref('')
const showCreateModal = ref(false)
const showViewModal = ref(false)
const creating = ref(false)
const viewQR = ref({})
const viewCanvas = ref(null)
const canvasRefs = {}

const products = ref([])
const productSearch = ref('')
const filteredProducts = ref([])
const selectedProduct = ref(null)

const sceneOptions = [
  { value: 'product_view', label: '商品查看', icon: '🔍' },
  { value: 'product_buy', label: '下单购买', icon: '🛒' },
  { value: 'order_status', label: '订单状态', icon: '📋' },
  { value: 'home', label: '首页', icon: '🏠' },
  { value: 'custom', label: '自定义', icon: '⚙️' }
]

const pageOptions = [
  { value: 'pages/index/index', label: '首页' },
  { value: 'pages/product/index', label: '商品详情' },
  { value: 'pages/cart/index', label: '购物车' },
  { value: 'pages/order/index', label: '订单列表' },
  { value: 'pages/user/index', label: '个人中心' }
]

const form = reactive({
  name: '',
  scene: 'product_view',
  platform: 'wechat',
  product_id: 0,
  order_no: '',
  page: '',
  params: ''
})

const namePlaceholder = computed(() => {
  const map = {
    product_view: '例如: 新款银戒指二维码',
    product_buy: '例如: 限时特价手链',
    order_status: '例如: 订单 #20240101 查询',
    home: '例如: 首页入口',
    custom: '例如: 活动页面'
  }
  return map[form.scene] || '请输入名称'
})

function sceneLabel(scene) {
  const item = sceneOptions.find(s => s.value === scene)
  return item ? item.label : '自定义'
}

function setCanvasRef(el, id) {
  if (el) canvasRefs[id] = el
}

function onSceneChange() {
  form.product_id = 0
  form.order_no = ''
  form.page = ''
  form.params = ''
  selectedProduct.value = null
  productSearch.value = ''
  filteredProducts.value = []

  if (form.scene === 'product_view' || form.scene === 'product_buy') {
    loadProducts()
  }
}

async function loadProducts() {
  if (products.value.length > 0) {
    filteredProducts.value = products.value.slice(0, 10)
    return
  }
  try {
    const result = await productAPI.list({ page: 1, page_size: 200 })
    products.value = result?.data || result || []
    filteredProducts.value = products.value.slice(0, 10)
  } catch (e) {
    console.error('加载商品失败', e)
  }
}

function filterProducts() {
  const kw = productSearch.value.toLowerCase().trim()
  if (!kw) {
    filteredProducts.value = products.value.slice(0, 10)
    return
  }
  filteredProducts.value = products.value
    .filter(p => p.name.toLowerCase().includes(kw))
    .slice(0, 10)
}

function selectProduct(p) {
  form.product_id = p.id
  selectedProduct.value = p
  if (!form.name) {
    form.name = `${p.name} - ${form.scene === 'product_buy' ? '下单购买' : '商品查看'}`
  }
}

function drawQR(canvas, text, size) {
  if (!canvas || !text) return
  const ctx = canvas.getContext('2d')
  ctx.clearRect(0, 0, size, size)

  const img = new Image()
  img.crossOrigin = 'anonymous'
  img.onload = () => ctx.drawImage(img, 0, 0, size, size)
  img.onerror = () => {
    ctx.fillStyle = '#f5f5f5'
    ctx.fillRect(0, 0, size, size)
    ctx.fillStyle = '#333'
    ctx.font = '12px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText('QR Code', size / 2, size / 2)
  }
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
  Object.assign(form, {
    name: '', scene: 'product_view', platform: 'wechat',
    product_id: 0, order_no: '', page: '', params: ''
  })
  selectedProduct.value = null
  productSearch.value = ''
  filteredProducts.value = []
  showCreateModal.value = true
  loadProducts()
}

async function handleCreate() {
  if (!form.name.trim()) return alert('请输入名称')

  if ((form.scene === 'product_view' || form.scene === 'product_buy') && !form.product_id) {
    return alert('请选择一个商品')
  }
  if (form.scene === 'order_status' && !form.order_no.trim()) {
    return alert('请输入订单号')
  }
  if (form.scene === 'custom' && !form.page) {
    return alert('请选择页面路径')
  }

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
  nextTick(() => drawQR(viewCanvas.value, qr.content, 280))
}

function downloadQR(qr) {
  const tempCanvas = document.createElement('canvas')
  const size = 400
  tempCanvas.width = size
  tempCanvas.height = size + 80

  const ctx = tempCanvas.getContext('2d')
  ctx.fillStyle = '#fff'
  ctx.fillRect(0, 0, tempCanvas.width, tempCanvas.height)

  const img = new Image()
  img.crossOrigin = 'anonymous'
  img.onload = () => {
    ctx.drawImage(img, 0, 0, size, size)
    ctx.fillStyle = '#333'
    ctx.font = 'bold 16px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText(qr.name, size / 2, size + 24)
    ctx.font = '13px sans-serif'
    ctx.fillStyle = '#888'
    const platformText = qr.platform === 'wechat' ? '微信小程序' : '支付宝小程序'
    ctx.fillText(`${platformText} · ${sceneLabel(qr.scene)}`, size / 2, size + 48)
    ctx.font = '11px sans-serif'
    ctx.fillStyle = '#bbb'
    ctx.fillText('扫码体验', size / 2, size + 68)

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

onMounted(() => fetchList())
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-title { font-size: 24px; }

.filter-bar { display: flex; gap: 8px; margin-bottom: 16px; }
.filter-btn {
  padding: 6px 16px; border: 1px solid #ddd; border-radius: 20px;
  background: #fff; cursor: pointer; font-size: 13px; transition: all 0.2s;
}
.filter-btn.active { background: var(--primary-color); color: #fff; border-color: var(--primary-color); }

.qr-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(360px, 1fr)); gap: 16px; }
.qr-card {
  display: flex; align-items: center; padding: 16px;
  background: #fafafa; border-radius: 8px; gap: 16px;
}

.qr-canvas-wrap {
  flex-shrink: 0; width: 100px; height: 100px; background: #fff;
  border-radius: 8px; padding: 4px; box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}
.qr-canvas-wrap canvas { width: 100%; height: 100%; }

.qr-info { flex: 1; min-width: 0; }
.qr-name { font-weight: 600; font-size: 15px; margin-bottom: 4px; }
.qr-tags { display: flex; gap: 6px; margin-bottom: 4px; flex-wrap: wrap; }

.platform-tag {
  display: inline-block; padding: 2px 8px; border-radius: 10px;
  font-size: 11px; font-weight: 500;
}
.platform-wechat { background: #e6f7e6; color: #07c160; }
.platform-alipay { background: #e6f0ff; color: #1677ff; }

.scene-tag {
  display: inline-block; padding: 2px 8px; border-radius: 10px;
  font-size: 11px; font-weight: 500; background: #f0f0f0; color: #666;
}
.scene-product_view { background: #fff7e6; color: #d48806; }
.scene-product_buy { background: #f6ffed; color: #389e0d; }
.scene-order_status { background: #e6f7ff; color: #096dd9; }
.scene-home { background: #f9f0ff; color: #722ed1; }

.qr-page { font-size: 12px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.qr-params { font-size: 11px; color: #999; }
.qr-time { font-size: 11px; color: #bbb; margin-top: 2px; }
.qr-actions { display: flex; flex-direction: column; gap: 6px; flex-shrink: 0; }

/* Scene grid */
.scene-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; }
.scene-card {
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  padding: 12px 4px; border: 2px solid #eee; border-radius: 8px;
  cursor: pointer; transition: all 0.2s; text-align: center;
}
.scene-card:hover { border-color: #ccc; }
.scene-card.active { border-color: var(--primary-color); background: rgba(212, 165, 116, 0.08); }
.scene-icon { font-size: 24px; }
.scene-name { font-size: 12px; color: #666; }

/* Radio group */
.radio-group { display: flex; gap: 12px; }
.radio-item {
  display: flex; align-items: center; gap: 6px; padding: 8px 16px;
  border: 2px solid #eee; border-radius: 8px; cursor: pointer;
  transition: all 0.2s; flex: 1; justify-content: center;
}
.radio-item.active { border-color: var(--primary-color); background: rgba(212, 165, 116, 0.08); }
.radio-item input[type="radio"] { display: none; }

/* Product selector */
.product-selector { position: relative; }
.product-list {
  max-height: 200px; overflow-y: auto; border: 1px solid #eee;
  border-radius: 8px; margin-top: 8px; background: #fff;
}
.product-option {
  display: flex; align-items: center; padding: 8px 12px; gap: 10px;
  cursor: pointer; transition: background 0.15s; border-bottom: 1px solid #f5f5f5;
}
.product-option:last-child { border-bottom: none; }
.product-option:hover { background: #f9f9f9; }
.product-option.selected { background: rgba(212, 165, 116, 0.1); }
.product-thumb { width: 40px; height: 40px; object-fit: cover; border-radius: 4px; flex-shrink: 0; }
.product-thumb.placeholder {
  display: flex; align-items: center; justify-content: center;
  background: #f0f0f0; color: #bbb; font-size: 10px;
}
.product-detail { flex: 1; min-width: 0; }
.product-title { font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.product-price { font-size: 12px; color: var(--primary-color); font-weight: 500; }
.check-mark { color: var(--primary-color); font-weight: bold; font-size: 18px; }
.selected-product {
  margin-top: 8px; padding: 8px 12px; background: #f6f6f6;
  border-radius: 6px; font-size: 13px; color: #555;
}

.form-hint { font-size: 12px; color: var(--text-secondary); margin-top: 4px; }

/* View modal */
.view-body { text-align: center; }
.view-qr-wrap {
  display: inline-block; background: #fff; padding: 12px;
  border-radius: 12px; box-shadow: 0 2px 12px rgba(0,0,0,0.08); margin-bottom: 16px;
}
.view-qr-wrap canvas { display: block; }
.view-detail { text-align: left; background: #f9f9f9; border-radius: 8px; padding: 12px 16px; }
.detail-row {
  display: flex; align-items: flex-start; padding: 6px 0;
  border-bottom: 1px solid #eee; font-size: 13px;
}
.detail-row:last-child { border-bottom: none; }
.detail-label { width: 50px; flex-shrink: 0; color: #999; }
.content-text { word-break: break-all; font-family: monospace; font-size: 11px; color: #666; }
.btn-block { width: 100%; }
</style>
