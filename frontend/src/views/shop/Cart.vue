<template>
  <div class="cart-page">
    <header class="page-header">
      <button class="back-btn" @click="$router.back()">←</button>
      <h1>购物车</h1>
      <span></span>
    </header>

    <div class="cart-content" v-if="cartItems.length > 0">
      <div v-for="item in cartItems" :key="item.id" class="cart-item">
        <input
          type="checkbox"
          v-model="selectedItems"
          :value="item.id"
          class="item-checkbox"
        />
        <img :src="item.product?.main_image || '/placeholder.png'" class="item-image" />
        <div class="item-info">
          <div class="item-name">{{ item.product?.name }}</div>
          <div class="item-spec" v-if="item.spec">{{ item.spec.name }}: {{ item.spec.value }}</div>
          <div class="item-price">¥{{ getItemPrice(item).toFixed(2) }}</div>
        </div>
        <div class="item-quantity">
          <button @click="updateQuantity(item, -1)" :disabled="item.quantity <= 1">-</button>
          <span>{{ item.quantity }}</span>
          <button @click="updateQuantity(item, 1)" :disabled="item.quantity >= item.product?.stock">+</button>
        </div>
        <button class="delete-btn" @click="removeItem(item)">×</button>
      </div>
    </div>

    <div v-else class="empty-cart">
      <div class="empty-icon">🛒</div>
      <p>购物车是空的</p>
      <button class="btn btn-primary" @click="$router.push('/shop')">去逛逛</button>
    </div>

    <!-- 底部结算栏 -->
    <div class="checkout-bar" v-if="cartItems.length > 0">
      <label class="select-all">
        <input type="checkbox" v-model="selectAll" @change="toggleSelectAll" />
        <span>全选</span>
      </label>
      <div class="total-info">
        <span>合计: </span>
        <span class="total-price">¥{{ totalPrice.toFixed(2) }}</span>
      </div>
      <button class="checkout-btn" @click="checkout" :disabled="selectedItems.length === 0">
        结算({{ selectedItems.length }})
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { cartAPI } from '../../api'

const router = useRouter()

const cartItems = ref([])
const selectedItems = ref([])

const selectAll = computed({
  get() {
    return cartItems.value.length > 0 && selectedItems.value.length === cartItems.value.length
  },
  set(val) {
    // handled by toggleSelectAll
  }
})

const totalPrice = computed(() => {
  return cartItems.value
    .filter(item => selectedItems.value.includes(item.id))
    .reduce((sum, item) => sum + getItemPrice(item) * item.quantity, 0)
})

const fetchCart = async () => {
  try {
    cartItems.value = await cartAPI.list() || []
    // 默认全选
    selectedItems.value = cartItems.value.map(item => item.id)
  } catch (error) {
    console.error('获取购物车失败:', error)
  }
}

const getItemPrice = (item) => {
  let price = item.product?.price || 0
  if (item.spec) {
    price += item.spec.price_adjustment || 0
  }
  return price
}

const toggleSelectAll = () => {
  if (selectedItems.value.length === cartItems.value.length) {
    selectedItems.value = []
  } else {
    selectedItems.value = cartItems.value.map(item => item.id)
  }
}

const updateQuantity = async (item, delta) => {
  const newQuantity = item.quantity + delta
  if (newQuantity < 1 || newQuantity > item.product?.stock) return

  try {
    await cartAPI.update(item.id, { quantity: newQuantity })
    item.quantity = newQuantity
  } catch (error) {
    alert(error.message)
  }
}

const removeItem = async (item) => {
  if (!confirm('确定要删除这个商品吗？')) return

  try {
    await cartAPI.remove(item.id)
    cartItems.value = cartItems.value.filter(i => i.id !== item.id)
    selectedItems.value = selectedItems.value.filter(id => id !== item.id)
  } catch (error) {
    alert(error.message)
  }
}

const checkout = () => {
  if (selectedItems.value.length === 0) {
    alert('请选择商品')
    return
  }

  // 将选中的购物车ID存入sessionStorage
  sessionStorage.setItem('checkoutCartIds', JSON.stringify(selectedItems.value))
  router.push('/shop/orders?action=create')
}

onMounted(() => {
  fetchCart()
})
</script>

<style scoped>
.cart-page {
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
  border-bottom: 1px solid var(--border-color);
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

.cart-content {
  padding: 10px;
}

.cart-item {
  display: flex;
  align-items: center;
  background: #fff;
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 10px;
}

.item-checkbox {
  width: 18px;
  height: 18px;
  margin-right: 10px;
}

.item-image {
  width: 80px;
  height: 80px;
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
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-spec {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.item-price {
  color: var(--error-color);
  font-weight: 500;
}

.item-quantity {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-quantity button {
  width: 24px;
  height: 24px;
  border: 1px solid var(--border-color);
  background: #fff;
  border-radius: 4px;
  cursor: pointer;
}

.item-quantity button:disabled {
  opacity: 0.5;
}

.delete-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: none;
  font-size: 18px;
  color: #999;
  cursor: pointer;
  margin-left: 10px;
}

.empty-cart {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  text-align: center;
}

.empty-icon {
  font-size: 60px;
  margin-bottom: 16px;
}

.empty-cart p {
  color: var(--text-secondary);
  margin-bottom: 20px;
}

.checkout-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  background: #fff;
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  z-index: 100;
}

.select-all {
  display: flex;
  align-items: center;
  gap: 6px;
}

.total-info {
  flex: 1;
  text-align: right;
  margin-right: 16px;
}

.total-price {
  font-size: 18px;
  color: var(--error-color);
  font-weight: 600;
}

.checkout-btn {
  padding: 12px 24px;
  background: var(--primary-color);
  color: #fff;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  cursor: pointer;
}

.checkout-btn:disabled {
  background: #ccc;
}
</style>
