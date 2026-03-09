<template>
  <div class="orders-page">
    <div class="page-header">
      <h2 class="page-title">订单管理</h2>
    </div>

    <div class="card">
      <div class="filter-bar">
        <input
          v-model="filter.keyword"
          type="text"
          class="form-input"
          placeholder="搜索订单号/收货人..."
          style="width: 200px"
          @input="debouncedSearch"
        />
        <select v-model="filter.status" class="form-input form-select" style="width: 150px" @change="fetchOrders">
          <option value="">全部状态</option>
          <option value="pending">待付款</option>
          <option value="paid">已付款</option>
          <option value="shipped">已发货</option>
          <option value="received">已收货</option>
          <option value="completed">已完成</option>
          <option value="cancelled">已取消</option>
        </select>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>订单号</th>
            <th>商品</th>
            <th>金额</th>
            <th>收货人</th>
            <th>状态</th>
            <th>下单时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="order in orders" :key="order.id">
            <td>
              <div class="order-no">{{ order.order_no }}</div>
              <div class="order-source">{{ getSourceText(order.order_source) }}</div>
            </td>
            <td>
              <div class="order-items">
                <div v-for="item in order.items?.slice(0, 2)" :key="item.id" class="order-item">
                  <img :src="item.product_image || '/placeholder.png'" class="item-image" />
                  <span class="item-name">{{ item.product_name }}</span>
                  <span class="item-qty">x{{ item.quantity }}</span>
                </div>
                <div v-if="order.items?.length > 2" class="more-items">
                  等{{ order.items.length }}件商品
                </div>
              </div>
            </td>
            <!-- 可编辑金额 -->
            <td>
              <div class="editable-price" @click="openEditPrice(order)">
                <span class="price">¥{{ order.pay_price.toFixed(2) }}</span>
                <span class="edit-icon">✏️</span>
              </div>
            </td>
            <td>
              <div>{{ order.address_name }}</div>
              <div class="contact">{{ order.address_phone }}</div>
            </td>
            <!-- 可选择状态 -->
            <td>
              <select
                class="status-select"
                :class="'status-' + order.status"
                :value="order.status"
                @change="onStatusChange(order, $event)"
              >
                <option value="pending">待付款</option>
                <option value="paid">已付款</option>
                <option value="shipped">已发货</option>
                <option value="received">已收货</option>
                <option value="completed">已完成</option>
                <option value="cancelled">已取消</option>
              </select>
            </td>
            <td>{{ formatDate(order.created_at) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn btn-small btn-secondary" @click="viewOrder(order)">详情</button>
                <button class="btn btn-small btn-outline" @click="openRemarkModal(order)">备注</button>
                <button class="btn btn-small btn-outline" @click="openLogsModal(order)">历史</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="orders.length === 0" class="empty-state">
        暂无订单
      </div>

      <div class="pagination" v-if="total > pageSize">
        <button class="pagination-btn" :disabled="page === 1" @click="page--; fetchOrders()">上一页</button>
        <span class="pagination-info">第 {{ page }} / {{ totalPages }} 页</span>
        <button class="pagination-btn" :disabled="page >= totalPages" @click="page++; fetchOrders()">下一页</button>
      </div>
    </div>

    <!-- 订单详情弹窗 -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="showDetailModal = false">
      <div class="modal" style="max-width: 600px">
        <div class="modal-header">
          <h3 class="modal-title">订单详情</h3>
          <button class="modal-close" @click="showDetailModal = false">&times;</button>
        </div>
        <div class="modal-body" v-if="currentOrder">
          <div class="detail-section">
            <h4>订单信息</h4>
            <div class="detail-row">
              <span class="detail-label">订单号:</span>
              <span>{{ currentOrder.order_no }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">状态:</span>
              <span class="status-tag" :class="'status-' + currentOrder.status">
                {{ getStatusText(currentOrder.status) }}
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">下单时间:</span>
              <span>{{ formatDate(currentOrder.created_at) }}</span>
            </div>
            <div class="detail-row" v-if="currentOrder.pay_time">
              <span class="detail-label">付款时间:</span>
              <span>{{ formatDate(currentOrder.pay_time) }}</span>
            </div>
            <div class="detail-row" v-if="currentOrder.ship_time">
              <span class="detail-label">发货时间:</span>
              <span>{{ formatDate(currentOrder.ship_time) }}</span>
            </div>
            <div class="detail-row" v-if="currentOrder.remark">
              <span class="detail-label">备注:</span>
              <span>{{ currentOrder.remark }}</span>
            </div>
          </div>

          <div class="detail-section">
            <h4>收货信息</h4>
            <div class="detail-row">
              <span class="detail-label">收货人:</span>
              <span>{{ currentOrder.address_name }} {{ currentOrder.address_phone }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">地址:</span>
              <span>{{ currentOrder.address_province }}{{ currentOrder.address_city }}{{ currentOrder.address_district }}{{ currentOrder.address_detail }}</span>
            </div>
          </div>

          <div class="detail-section" v-if="currentOrder.express_no">
            <h4>物流信息</h4>
            <div class="detail-row">
              <span class="detail-label">快递公司:</span>
              <span>{{ currentOrder.express_company }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">快递单号:</span>
              <span>{{ currentOrder.express_no }}</span>
            </div>
          </div>

          <div class="detail-section">
            <h4>商品清单</h4>
            <div v-for="item in currentOrder.items" :key="item.id" class="order-detail-item">
              <img :src="item.product_image || '/placeholder.png'" class="detail-item-image" />
              <div class="detail-item-info">
                <div class="detail-item-name">{{ item.product_name }}</div>
                <div class="detail-item-spec" v-if="item.spec_name">{{ item.spec_name }}</div>
              </div>
              <div class="detail-item-price">¥{{ item.price.toFixed(2) }} x {{ item.quantity }}</div>
            </div>
          </div>

          <div class="detail-section">
            <div class="detail-row total-row">
              <span class="detail-label">订单总额:</span>
              <span class="price">¥{{ currentOrder.pay_price.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 发货弹窗 -->
    <div v-if="showShipModal" class="modal-overlay" @click.self="showShipModal = false">
      <div class="modal" style="max-width: 400px">
        <div class="modal-header">
          <h3 class="modal-title">填写物流信息</h3>
          <button class="modal-close" @click="showShipModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">快递公司</label>
            <select v-model="shipForm.express_company" class="form-input form-select">
              <option value="顺丰速运">顺丰速运</option>
              <option value="圆通速递">圆通速递</option>
              <option value="中通快递">中通快递</option>
              <option value="韵达快递">韵达快递</option>
              <option value="申通快递">申通快递</option>
              <option value="EMS">EMS</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">快递单号</label>
            <input v-model="shipForm.express_no" type="text" class="form-input" placeholder="请输入快递单号" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showShipModal = false">取消</button>
          <button class="btn btn-primary" @click="confirmShip" :disabled="shipping">
            {{ shipping ? '处理中...' : '确认发货' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 修改价格弹窗 -->
    <div v-if="showPriceModal" class="modal-overlay" @click.self="showPriceModal = false">
      <div class="modal" style="max-width: 380px">
        <div class="modal-header">
          <h3 class="modal-title">修改订单金额</h3>
          <button class="modal-close" @click="showPriceModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">当前金额</label>
            <div class="current-val">¥{{ priceForm.oldPrice.toFixed(2) }}</div>
          </div>
          <div class="form-group">
            <label class="form-label">新金额</label>
            <input v-model.number="priceForm.newPrice" type="number" step="0.01" min="0" class="form-input" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showPriceModal = false">取消</button>
          <button class="btn btn-primary" @click="confirmUpdatePrice" :disabled="savingPrice">
            {{ savingPrice ? '保存中...' : '确认修改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 备注弹窗 -->
    <div v-if="showRemarkModal" class="modal-overlay" @click.self="showRemarkModal = false">
      <div class="modal" style="max-width: 450px">
        <div class="modal-header">
          <h3 class="modal-title">订单备注</h3>
          <button class="modal-close" @click="showRemarkModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group" v-if="remarkForm.oldRemark">
            <label class="form-label">当前备注</label>
            <div class="current-val remark-text">{{ remarkForm.oldRemark }}</div>
          </div>
          <div class="form-group">
            <label class="form-label">{{ remarkForm.oldRemark ? '修改备注' : '添加备注' }}</label>
            <textarea v-model="remarkForm.newRemark" class="form-input form-textarea" rows="3" placeholder="输入备注内容..."></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showRemarkModal = false">取消</button>
          <button class="btn btn-primary" @click="confirmUpdateRemark" :disabled="savingRemark">
            {{ savingRemark ? '保存中...' : '确认保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 历史记录弹窗 -->
    <div v-if="showLogsModal" class="modal-overlay" @click.self="showLogsModal = false">
      <div class="modal" style="max-width: 620px">
        <div class="modal-header">
          <h3 class="modal-title">订单变更历史 - {{ logsOrderNo }}</h3>
          <button class="modal-close" @click="showLogsModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div v-if="loadingLogs" class="loading-text">加载中...</div>
          <div v-else-if="changeLogs.length === 0" class="empty-logs">暂无变更记录</div>
          <div v-else class="logs-timeline">
            <div v-for="log in changeLogs" :key="log.id" class="log-item">
              <div class="log-icon" :class="'log-type-' + log.change_type">
                {{ getLogIcon(log.change_type) }}
              </div>
              <div class="log-content">
                <div class="log-title">
                  <span class="log-type-label">{{ getLogTypeLabel(log.change_type) }}</span>
                  <span class="log-time">{{ formatDate(log.created_at) }}</span>
                </div>
                <div class="log-detail">
                  <template v-if="log.change_type === 'status'">
                    <span class="status-tag mini" :class="'status-' + log.old_value">{{ getStatusText(log.old_value) }}</span>
                    <span class="log-arrow">→</span>
                    <span class="status-tag mini" :class="'status-' + log.new_value">{{ getStatusText(log.new_value) }}</span>
                  </template>
                  <template v-else-if="log.change_type === 'price'">
                    <span class="old-price">¥{{ log.old_value }}</span>
                    <span class="log-arrow">→</span>
                    <span class="new-price">¥{{ log.new_value }}</span>
                  </template>
                  <template v-else-if="log.change_type === 'remark'">
                    <div v-if="log.old_value" class="remark-change">
                      <div class="remark-old">旧: {{ log.old_value }}</div>
                    </div>
                    <div class="remark-new">新: {{ log.new_value }}</div>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { orderAPI } from '../api'

const orders = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

const filter = reactive({
  keyword: '',
  status: ''
})

const showDetailModal = ref(false)
const showShipModal = ref(false)
const showPriceModal = ref(false)
const showRemarkModal = ref(false)
const showLogsModal = ref(false)
const currentOrder = ref(null)
const shipping = ref(false)
const savingPrice = ref(false)
const savingRemark = ref(false)
const loadingLogs = ref(false)

const shipForm = reactive({ express_company: '顺丰速运', express_no: '' })
const priceForm = reactive({ orderId: 0, oldPrice: 0, newPrice: 0 })
const remarkForm = reactive({ orderId: 0, oldRemark: '', newRemark: '' })
const changeLogs = ref([])
const logsOrderNo = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize))

const fetchOrders = async () => {
  try {
    const result = await orderAPI.list({
      page: page.value,
      page_size: pageSize,
      keyword: filter.keyword,
      status: filter.status
    })
    orders.value = result.data || []
    total.value = result.total || 0
  } catch (error) {
    console.error('获取订单列表失败:', error)
  }
}

const statusOptions = [
  { value: 'pending', label: '待付款' },
  { value: 'paid', label: '已付款' },
  { value: 'shipped', label: '已发货' },
  { value: 'received', label: '已收货' },
  { value: 'completed', label: '已完成' },
  { value: 'cancelled', label: '已取消' }
]

const getStatusText = (status) => {
  const found = statusOptions.find(s => s.value === status)
  return found ? found.label : status
}

const getSourceText = (source) => {
  const map = { web: '网页', wechat_mp: '微信小程序', alipay_mp: '支付宝小程序' }
  return map[source] || source
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const viewOrder = (order) => {
  currentOrder.value = order
  showDetailModal.value = true
}

// ---- 状态修改 ----
const onStatusChange = async (order, event) => {
  const newStatus = event.target.value
  const oldStatus = order.status
  if (newStatus === oldStatus) return

  if (newStatus === 'shipped') {
    currentOrder.value = order
    shipForm.express_company = '顺丰速运'
    shipForm.express_no = ''
    showShipModal.value = true
    event.target.value = oldStatus
    return
  }

  if (!confirm(`确认将状态从「${getStatusText(oldStatus)}」修改为「${getStatusText(newStatus)}」？`)) {
    event.target.value = oldStatus
    return
  }

  try {
    await orderAPI.updateStatus(order.id, { status: newStatus })
    order.status = newStatus
  } catch (err) {
    alert('修改失败: ' + err.message)
    event.target.value = oldStatus
  }
}

const openShipModal = (order) => {
  currentOrder.value = order
  shipForm.express_company = '顺丰速运'
  shipForm.express_no = ''
  showShipModal.value = true
}

const confirmShip = async () => {
  if (!shipForm.express_no) return alert('请输入快递单号')
  shipping.value = true
  try {
    await orderAPI.updateStatus(currentOrder.value.id, {
      status: 'shipped',
      express_company: shipForm.express_company,
      express_no: shipForm.express_no
    })
    showShipModal.value = false
    fetchOrders()
  } catch (error) {
    alert(error.message)
  } finally {
    shipping.value = false
  }
}

// ---- 价格修改 ----
const openEditPrice = (order) => {
  priceForm.orderId = order.id
  priceForm.oldPrice = order.pay_price
  priceForm.newPrice = order.pay_price
  currentOrder.value = order
  showPriceModal.value = true
}

const confirmUpdatePrice = async () => {
  if (priceForm.newPrice < 0) return alert('金额不能为负数')
  if (priceForm.newPrice === priceForm.oldPrice) { showPriceModal.value = false; return }
  savingPrice.value = true
  try {
    await orderAPI.updatePrice(priceForm.orderId, priceForm.newPrice)
    currentOrder.value.pay_price = priceForm.newPrice
    currentOrder.value.total_price = priceForm.newPrice
    showPriceModal.value = false
  } catch (err) {
    alert('修改失败: ' + err.message)
  } finally {
    savingPrice.value = false
  }
}

// ---- 备注修改 ----
const openRemarkModal = (order) => {
  remarkForm.orderId = order.id
  remarkForm.oldRemark = order.remark || ''
  remarkForm.newRemark = order.remark || ''
  currentOrder.value = order
  showRemarkModal.value = true
}

const confirmUpdateRemark = async () => {
  savingRemark.value = true
  try {
    await orderAPI.updateRemark(remarkForm.orderId, remarkForm.newRemark)
    currentOrder.value.remark = remarkForm.newRemark
    showRemarkModal.value = false
  } catch (err) {
    alert('保存失败: ' + err.message)
  } finally {
    savingRemark.value = false
  }
}

// ---- 历史记录 ----
const getLogIcon = (type) => {
  const icons = { status: '🔄', price: '💰', remark: '📝' }
  return icons[type] || '📋'
}

const getLogTypeLabel = (type) => {
  const labels = { status: '状态变更', price: '金额修改', remark: '备注修改' }
  return labels[type] || type
}

const openLogsModal = async (order) => {
  logsOrderNo.value = order.order_no
  changeLogs.value = []
  showLogsModal.value = true
  loadingLogs.value = true
  try {
    const res = await orderAPI.getChangeLogs(order.id)
    changeLogs.value = res || []
  } catch (err) {
    console.error(err)
  } finally {
    loadingLogs.value = false
  }
}

let searchTimer = null
const debouncedSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchOrders()
  }, 300)
}

onMounted(() => {
  fetchOrders()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-title { font-size: 24px; }

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.order-no { font-weight: 500; font-family: monospace; }
.order-source { font-size: 12px; color: var(--text-secondary); }

.order-items { max-width: 200px; }
.order-item { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.item-image { width: 32px; height: 32px; object-fit: cover; border-radius: 4px; }
.item-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; }
.item-qty { font-size: 12px; color: var(--text-secondary); }
.more-items { font-size: 12px; color: var(--text-secondary); }
.contact { font-size: 12px; color: var(--text-secondary); }

/* 可编辑价格 */
.editable-price {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.15s;
}
.editable-price:hover { background: #f5f5f5; }
.editable-price .edit-icon { font-size: 12px; opacity: 0; transition: opacity 0.15s; }
.editable-price:hover .edit-icon { opacity: 1; }

/* 状态下拉 */
.status-select {
  padding: 4px 24px 4px 10px;
  border: 1px solid #e0e0e0;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 10 10'%3E%3Cpath fill='%23999' d='M5 7L1 3h8z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  transition: all 0.15s;
}
.status-select:focus { outline: none; border-color: var(--primary-color); }
.status-select.status-pending { background-color: #fff8e1; color: #f57f17; border-color: #ffe082; }
.status-select.status-paid { background-color: #e3f2fd; color: #1565c0; border-color: #90caf9; }
.status-select.status-shipped { background-color: #e8f5e9; color: #2e7d32; border-color: #a5d6a7; }
.status-select.status-received { background-color: #f3e5f5; color: #7b1fa2; border-color: #ce93d8; }
.status-select.status-completed { background-color: #e8f5e9; color: #1b5e20; border-color: #81c784; }
.status-select.status-cancelled { background-color: #fafafa; color: #9e9e9e; border-color: #e0e0e0; }

.action-btns { display: flex; gap: 6px; flex-wrap: wrap; }
.btn-outline {
  background: #fff;
  border: 1px solid #d0d0d0;
  color: #555;
  border-radius: 4px;
  padding: 4px 10px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-outline:hover { border-color: var(--primary-color); color: var(--primary-color); }

.pagination-info { padding: 0 16px; color: var(--text-secondary); }

/* 详情弹窗 */
.detail-section { margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px solid var(--border-color); }
.detail-section:last-child { border-bottom: none; margin-bottom: 0; }
.detail-section h4 { font-size: 14px; color: var(--text-secondary); margin-bottom: 12px; }
.detail-row { display: flex; margin-bottom: 8px; }
.detail-label { width: 80px; color: var(--text-secondary); flex-shrink: 0; }
.total-row { font-size: 16px; font-weight: 500; }

.order-detail-item { display: flex; align-items: center; padding: 10px 0; border-bottom: 1px solid var(--border-color); }
.order-detail-item:last-child { border-bottom: none; }
.detail-item-image { width: 50px; height: 50px; object-fit: cover; border-radius: 4px; }
.detail-item-info { flex: 1; margin-left: 12px; }
.detail-item-name { font-weight: 500; }
.detail-item-spec { font-size: 12px; color: var(--text-secondary); }
.detail-item-price { font-weight: 500; }

/* 价格/备注弹窗 */
.current-val {
  padding: 8px 12px;
  background: #f8f8f8;
  border-radius: 6px;
  font-size: 14px;
  color: #666;
}
.remark-text { white-space: pre-wrap; word-break: break-all; }
.form-textarea { resize: vertical; min-height: 60px; }

/* 历史记录 */
.loading-text { text-align: center; padding: 30px; color: #999; }
.empty-logs { text-align: center; padding: 30px; color: #999; }

.logs-timeline { padding: 0; }
.log-item {
  display: flex;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #f0f0f0;
}
.log-item:last-child { border-bottom: none; }

.log-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  flex-shrink: 0;
  background: #f5f5f5;
}
.log-type-status { background: #e3f2fd; }
.log-type-price { background: #fff8e1; }
.log-type-remark { background: #f3e5f5; }

.log-content { flex: 1; min-width: 0; }
.log-title { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.log-type-label { font-weight: 600; font-size: 13px; }
.log-time { font-size: 12px; color: #999; }

.log-detail { font-size: 13px; }
.log-arrow { margin: 0 6px; color: #999; }
.old-price { color: #999; text-decoration: line-through; }
.new-price { color: var(--primary-color); font-weight: 600; }

.status-tag.mini { font-size: 11px; padding: 2px 8px; border-radius: 10px; }

.remark-change { margin-bottom: 4px; }
.remark-old { color: #999; font-size: 12px; }
.remark-new { color: #333; }
</style>
