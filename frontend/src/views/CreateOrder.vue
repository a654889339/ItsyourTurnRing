<template>
  <div class="create-order-page">
    <div class="page-header">
      <h2 class="page-title">快速下单</h2>
    </div>

    <div class="order-layout">
      <!-- 左：选择商品 -->
      <div class="card order-products">
        <h3 class="section-title">选择商品</h3>
        <div class="product-search">
          <input v-model="searchKw" type="text" class="form-input" placeholder="搜索商品名称..." @input="filterProductList" />
        </div>
        <div class="product-pick-list">
          <div
            v-for="p in filteredProducts"
            :key="p.id"
            class="pick-item"
            :class="{ disabled: p.stock <= 0 }"
            @click="addToCart(p)"
          >
            <img :src="p.main_image || '/placeholder.png'" class="pick-img" />
            <div class="pick-info">
              <div class="pick-name">{{ p.name }}</div>
              <div class="pick-meta">
                <span class="pick-price">¥{{ p.price.toFixed(2) }}</span>
                <span class="pick-stock" :class="{ low: p.stock < 5 }">库存 {{ p.stock }}</span>
              </div>
            </div>
            <span class="pick-add">+</span>
          </div>
          <div v-if="filteredProducts.length === 0" class="empty-state" style="padding: 20px;">暂无商品</div>
        </div>
      </div>

      <!-- 右：订单详情 -->
      <div class="order-detail-col">
        <!-- 已选商品 -->
        <div class="card">
          <h3 class="section-title">已选商品 <span class="item-count" v-if="cart.length">({{ cart.length }})</span></h3>
          <div v-if="cart.length === 0" class="empty-state" style="padding: 20px;">请从左侧选择商品</div>
          <div v-else class="cart-list">
            <div v-for="(item, idx) in cart" :key="idx" class="cart-item">
              <img :src="item.main_image || '/placeholder.png'" class="cart-img" />
              <div class="cart-info">
                <div class="cart-name">{{ item.name }}</div>
                <div class="cart-price">¥{{ item.price.toFixed(2) }}</div>
              </div>
              <div class="cart-qty">
                <button class="qty-btn" @click="changeQty(idx, -1)">-</button>
                <span class="qty-val">{{ item.quantity }}</span>
                <button class="qty-btn" @click="changeQty(idx, 1)">+</button>
              </div>
              <div class="cart-subtotal">¥{{ (item.price * item.quantity).toFixed(2) }}</div>
              <button class="cart-remove" @click="cart.splice(idx, 1)">&times;</button>
            </div>
            <div class="cart-total">
              合计: <strong>¥{{ totalPrice.toFixed(2) }}</strong>
            </div>
          </div>
        </div>

        <!-- 收货信息 -->
        <div class="card">
          <h3 class="section-title">收货信息</h3>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">收货人 *</label>
              <input v-model="address.name" type="text" class="form-input" placeholder="姓名" />
            </div>
            <div class="form-group">
              <label class="form-label">手机号 *</label>
              <input v-model="address.phone" type="text" class="form-input" placeholder="手机号" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">省</label>
              <input v-model="address.province" type="text" class="form-input" placeholder="省份" />
            </div>
            <div class="form-group">
              <label class="form-label">市</label>
              <input v-model="address.city" type="text" class="form-input" placeholder="城市" />
            </div>
            <div class="form-group">
              <label class="form-label">区</label>
              <input v-model="address.district" type="text" class="form-input" placeholder="区县" />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">详细地址 *</label>
            <input v-model="address.detail" type="text" class="form-input" placeholder="街道、门牌号等" />
          </div>
        </div>

        <!-- 备注 -->
        <div class="card">
          <h3 class="section-title">订单备注</h3>
          <textarea v-model="remark" class="form-input form-textarea" rows="2" placeholder="填写备注信息（可选）"></textarea>
        </div>

        <!-- 提交 -->
        <button
          class="btn btn-primary btn-submit"
          @click="submitOrder"
          :disabled="submitting || cart.length === 0"
        >
          {{ submitting ? '提交中...' : '确认下单' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { productAPI, orderAPI } from '../api'

const router = useRouter()
const allProducts = ref([])
const filteredProducts = ref([])
const searchKw = ref('')
const cart = ref([])
const submitting = ref(false)
const remark = ref('')

const address = reactive({
  name: '', phone: '', province: '', city: '', district: '', detail: ''
})

const totalPrice = computed(() => cart.value.reduce((sum, i) => sum + i.price * i.quantity, 0))

async function loadProducts() {
  try {
    const res = await productAPI.list({ page: 1, page_size: 200, status: 'available' })
    allProducts.value = res.data || []
    filteredProducts.value = allProducts.value
  } catch (e) { console.error(e) }
}

function filterProductList() {
  const kw = searchKw.value.toLowerCase().trim()
  if (!kw) { filteredProducts.value = allProducts.value; return }
  filteredProducts.value = allProducts.value.filter(p => p.name.toLowerCase().includes(kw))
}

function addToCart(product) {
  if (product.stock <= 0) return
  const existing = cart.value.find(i => i.product_id === product.id)
  if (existing) {
    if (existing.quantity < product.stock) existing.quantity++
    return
  }
  cart.value.push({
    product_id: product.id,
    name: product.name,
    price: product.price,
    main_image: product.main_image,
    stock: product.stock,
    quantity: 1
  })
}

function changeQty(idx, delta) {
  const item = cart.value[idx]
  const newQty = item.quantity + delta
  if (newQty <= 0) { cart.value.splice(idx, 1); return }
  if (newQty > item.stock) { alert('超出库存'); return }
  item.quantity = newQty
}

async function submitOrder() {
  if (cart.value.length === 0) return alert('请选择商品')
  if (!address.name.trim()) return alert('请填写收货人')
  if (!address.phone.trim()) return alert('请填写手机号')
  if (!address.detail.trim()) return alert('请填写详细地址')

  submitting.value = true
  try {
    const data = {
      items: cart.value.map(i => ({ product_id: i.product_id, quantity: i.quantity })),
      address_name: address.name,
      address_phone: address.phone,
      address_province: address.province,
      address_city: address.city,
      address_district: address.district,
      address_detail: address.detail,
      remark: remark.value
    }
    await orderAPI.createAdmin(data)
    alert('下单成功！')
    router.push('/orders')
  } catch (err) {
    alert('下单失败: ' + err.message)
  } finally {
    submitting.value = false
  }
}

onMounted(loadProducts)
</script>

<style scoped>
.page-header { margin-bottom: 24px; }
.page-title { font-size: 24px; }

.order-layout { display: flex; gap: 20px; align-items: flex-start; }
.order-products { width: 380px; flex-shrink: 0; }
.order-detail-col { flex: 1; display: flex; flex-direction: column; gap: 16px; }

.section-title { font-size: 16px; margin-bottom: 12px; }
.item-count { color: #999; font-weight: normal; font-size: 13px; }

/* Product pick list */
.product-search { margin-bottom: 10px; }
.product-pick-list { max-height: 520px; overflow-y: auto; }
.pick-item {
  display: flex; align-items: center; gap: 10px; padding: 10px;
  border-bottom: 1px solid #f0f0f0; cursor: pointer; transition: background 0.15s;
}
.pick-item:hover { background: #f9f9f9; }
.pick-item.disabled { opacity: 0.4; pointer-events: none; }
.pick-img { width: 48px; height: 48px; object-fit: cover; border-radius: 6px; flex-shrink: 0; }
.pick-info { flex: 1; min-width: 0; }
.pick-name { font-size: 14px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pick-meta { display: flex; gap: 8px; font-size: 12px; margin-top: 2px; }
.pick-price { color: var(--primary-color); font-weight: 600; }
.pick-stock { color: #999; }
.pick-stock.low { color: var(--error-color); }
.pick-add {
  width: 28px; height: 28px; border-radius: 50%; background: var(--primary-color);
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-size: 18px; font-weight: bold; flex-shrink: 0;
}

/* Cart */
.cart-list { }
.cart-item {
  display: flex; align-items: center; gap: 10px; padding: 10px 0;
  border-bottom: 1px solid #f5f5f5;
}
.cart-item:last-child { border-bottom: none; }
.cart-img { width: 40px; height: 40px; object-fit: cover; border-radius: 4px; }
.cart-info { flex: 1; min-width: 0; }
.cart-name { font-size: 13px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cart-price { font-size: 12px; color: #999; }
.cart-qty { display: flex; align-items: center; gap: 4px; }
.qty-btn {
  width: 24px; height: 24px; border: 1px solid #ddd; border-radius: 4px;
  background: #fff; cursor: pointer; font-size: 14px; display: flex;
  align-items: center; justify-content: center;
}
.qty-btn:hover { border-color: var(--primary-color); }
.qty-val { width: 28px; text-align: center; font-size: 14px; font-weight: 500; }
.cart-subtotal { font-size: 14px; font-weight: 600; color: var(--primary-color); width: 70px; text-align: right; }
.cart-remove {
  width: 20px; height: 20px; border-radius: 50%; background: none;
  border: none; color: #ccc; cursor: pointer; font-size: 16px;
}
.cart-remove:hover { color: var(--error-color); }
.cart-total {
  text-align: right; padding: 12px 0 4px; font-size: 15px;
  border-top: 1px solid #eee; margin-top: 8px;
}
.cart-total strong { color: var(--primary-color); font-size: 18px; }

/* Form */
.form-row { display: flex; gap: 12px; margin-bottom: 12px; }
.form-row .form-group { flex: 1; margin-bottom: 0; }

/* Submit */
.btn-submit { width: 100%; padding: 14px; font-size: 16px; border-radius: 8px; }
</style>
