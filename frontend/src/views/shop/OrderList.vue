<template>
  <div class="orders-page">
    <header class="page-header">
      <button class="back-btn" @click="$router.push('/shop')">←</button>
      <h1>我的订单</h1>
      <span></span>
    </header>

    <!-- 状态筛选 -->
    <div class="status-tabs">
      <span
        v-for="tab in statusTabs"
        :key="tab.value"
        class="tab-item"
        :class="{ active: currentStatus === tab.value }"
        @click="currentStatus = tab.value; fetchOrders()"
      >
        {{ tab.label }}
      </span>
    </div>

    <div class="orders-list">
      <div v-for="order in orders" :key="order.id" class="order-card">
        <div class="order-header">
          <span class="order-no">订单号: {{ order.order_no }}</span>
          <span class="order-status" :class="'status-' + order.status">
            {{ getStatusText(order.status) }}
          </span>
        </div>

        <div class="order-items">
          <div v-for="item in order.items" :key="item.id" class="order-item">
            <img :src="item.product_image || '/placeholder.png'" class="item-image" />
            <div class="item-info">
              <div class="item-name">{{ item.product_name }}</div>
              <div class="item-spec" v-if="item.spec_name">{{ item.spec_name }}</div>
              <div class="item-price">¥{{ item.price.toFixed(2) }} × {{ item.quantity }}</div>
            </div>
          </div>
        </div>

        <div class="order-footer">
          <div class="order-total">
            共{{ getTotalCount(order) }}件商品，合计: <span class="total-price">¥{{ order.pay_price.toFixed(2) }}</span>
          </div>
          <div class="order-actions">
            <button
              v-if="order.status === 'pending'"
              class="btn btn-small btn-primary"
              @click="payOrder(order)"
            >去付款</button>
            <button
              v-if="order.status === 'pending'"
              class="btn btn-small btn-secondary"
              @click="cancelOrder(order)"
            >取消</button>
            <button
              v-if="order.status === 'shipped'"
              class="btn btn-small btn-primary"
              @click="confirmReceive(order)"
            >确认收货</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="orders.length === 0" class="empty-state">
      暂无订单
    </div>

    <!-- 底部导航 -->
    <nav class="bottom-nav">
      <router-link to="/shop" class="nav-item">
        <span class="nav-icon">🏠</span>
        <span>首页</span>
      </router-link>
      <router-link to="/shop/cart" class="nav-item">
        <span class="nav-icon">🛒</span>
        <span>购物车</span>
      </router-link>
      <router-link to="/shop/orders" class="nav-item active">
        <span class="nav-icon">📦</span>
        <span>订单</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { orderAPI } from '../../api'

const statusTabs = [
  { label: '全部', value: '' },
  { label: '待付款', value: 'pending' },
  { label: '已付款', value: 'paid' },
  { label: '已发货', value: 'shipped' },
  { label: '已完成', value: 'completed' }
]

const orders = ref([])
const currentStatus = ref('')

const fetchOrders = async () => {
  try {
    const result = await orderAPI.list({
      page: 1,
      page_size: 50,
      status: currentStatus.value
    })
    orders.value = result.data || []
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

const getTotalCount = (order) => {
  return order.items?.reduce((sum, item) => sum + item.quantity, 0) || 0
}

const payOrder = async (order) => {
  if (!confirm('确认支付？')) return

  try {
    const result = await orderAPI.pay(order.id, { pay_method: 'wechat' })
    if (!result.need_pay) {
      alert('支付成功')
    } else {
      alert('已创建支付订单，请在小程序中完成支付')
    }
    fetchOrders()
  } catch (error) {
    alert(error.message)
  }
}

const cancelOrder = async (order) => {
  if (!confirm('确定要取消订单吗？')) return

  try {
    await orderAPI.cancel(order.id)
    fetchOrders()
  } catch (error) {
    alert(error.message)
  }
}

const confirmReceive = async (order) => {
  if (!confirm('确认已收到商品？')) return

  try {
    await orderAPI.receive(order.id)
    fetchOrders()
  } catch (error) {
    alert(error.message)
  }
}

onMounted(() => {
  fetchOrders()
})
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 70px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #fff;
}

.back-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
}

.page-header h1 {
  font-size: 18px;
}

.status-tabs {
  display: flex;
  background: #fff;
  border-bottom: 1px solid var(--border-color);
}

.tab-item {
  flex: 1;
  text-align: center;
  padding: 12px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
}

.tab-item.active {
  color: var(--primary-color);
  border-bottom: 2px solid var(--primary-color);
}

.orders-list {
  padding: 10px;
}

.order-card {
  background: #fff;
  border-radius: 8px;
  margin-bottom: 10px;
  overflow: hidden;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
}

.order-no {
  font-size: 12px;
  color: var(--text-secondary);
}

.order-status {
  font-size: 13px;
  font-weight: 500;
}

.status-pending {
  color: #faad14;
}

.status-paid,
.status-shipped {
  color: #1890ff;
}

.status-completed {
  color: #52c41a;
}

.status-cancelled {
  color: #999;
}

.order-items {
  padding: 12px;
}

.order-item {
  display: flex;
  margin-bottom: 10px;
}

.order-item:last-child {
  margin-bottom: 0;
}

.item-image {
  width: 60px;
  height: 60px;
  object-fit: cover;
  border-radius: 4px;
}

.item-info {
  flex: 1;
  margin-left: 10px;
}

.item-name {
  font-size: 14px;
  margin-bottom: 4px;
}

.item-spec {
  font-size: 12px;
  color: var(--text-secondary);
}

.item-price {
  font-size: 13px;
  color: var(--text-secondary);
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-top: 1px solid var(--border-color);
}

.order-total {
  font-size: 13px;
}

.total-price {
  color: var(--error-color);
  font-weight: 600;
}

.order-actions {
  display: flex;
  gap: 8px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: #fff;
  border-top: 1px solid var(--border-color);
  padding: 8px 0;
  z-index: 100;
}

.bottom-nav .nav-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 12px;
  color: var(--text-secondary);
  text-decoration: none;
}

.bottom-nav .nav-item.active {
  color: var(--primary-color);
}

.bottom-nav .nav-icon {
  font-size: 20px;
  margin-bottom: 2px;
}
</style>
