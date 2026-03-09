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
            <td>
              <span class="price">¥{{ order.pay_price.toFixed(2) }}</span>
            </td>
            <td>
              <div>{{ order.address_name }}</div>
              <div class="contact">{{ order.address_phone }}</div>
            </td>
            <td>
              <span class="status-tag" :class="'status-' + order.status">
                {{ getStatusText(order.status) }}
              </span>
            </td>
            <td>{{ formatDate(order.created_at) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn btn-small btn-secondary" @click="viewOrder(order)">详情</button>
                <button
                  v-if="order.status === 'paid'"
                  class="btn btn-small btn-primary"
                  @click="openShipModal(order)"
                >发货</button>
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
const currentOrder = ref(null)
const shipping = ref(false)

const shipForm = reactive({
  express_company: '顺丰速运',
  express_no: ''
})

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

const getStatusText = (status) => {
  const map = {
    pending: '待付款',
    paid: '已付款',
    shipped: '已发货',
    received: '已收货',
    completed: '已完成',
    cancelled: '已取消'
  }
  return map[status] || status
}

const getSourceText = (source) => {
  const map = {
    web: '网页',
    wechat_mp: '微信小程序',
    alipay_mp: '支付宝小程序'
  }
  return map[source] || source
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

const viewOrder = (order) => {
  currentOrder.value = order
  showDetailModal.value = true
}

const openShipModal = (order) => {
  currentOrder.value = order
  shipForm.express_company = '顺丰速运'
  shipForm.express_no = ''
  showShipModal.value = true
}

const confirmShip = async () => {
  if (!shipForm.express_no) {
    alert('请输入快递单号')
    return
  }

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

.page-title {
  font-size: 24px;
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.order-no {
  font-weight: 500;
  font-family: monospace;
}

.order-source {
  font-size: 12px;
  color: var(--text-secondary);
}

.order-items {
  max-width: 200px;
}

.order-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.item-image {
  width: 32px;
  height: 32px;
  object-fit: cover;
  border-radius: 4px;
}

.item-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.item-qty {
  font-size: 12px;
  color: var(--text-secondary);
}

.more-items {
  font-size: 12px;
  color: var(--text-secondary);
}

.contact {
  font-size: 12px;
  color: var(--text-secondary);
}

.action-btns {
  display: flex;
  gap: 8px;
}

.pagination-info {
  padding: 0 16px;
  color: var(--text-secondary);
}

/* 详情弹窗样式 */
.detail-section {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.detail-section:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.detail-section h4 {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.detail-row {
  display: flex;
  margin-bottom: 8px;
}

.detail-label {
  width: 80px;
  color: var(--text-secondary);
}

.total-row {
  font-size: 16px;
  font-weight: 500;
}

.order-detail-item {
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color);
}

.order-detail-item:last-child {
  border-bottom: none;
}

.detail-item-image {
  width: 50px;
  height: 50px;
  object-fit: cover;
  border-radius: 4px;
}

.detail-item-info {
  flex: 1;
  margin-left: 12px;
}

.detail-item-name {
  font-weight: 500;
}

.detail-item-spec {
  font-size: 12px;
  color: var(--text-secondary);
}

.detail-item-price {
  font-weight: 500;
}
</style>
