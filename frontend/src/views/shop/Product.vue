<template>
  <div class="product-detail" v-if="product">
    <!-- 商品图片 -->
    <div class="product-gallery">
      <img :src="product.main_image || '/placeholder.png'" class="main-image" />
    </div>

    <!-- 商品信息 -->
    <div class="product-info">
      <h1 class="product-name">{{ product.name }}</h1>
      <div class="product-price">
        <span class="current-price">¥{{ product.price.toFixed(2) }}</span>
        <span v-if="product.original_price" class="original-price">
          ¥{{ product.original_price.toFixed(2) }}
        </span>
      </div>
      <div class="product-meta">
        <span v-if="product.material">材质: {{ product.material }}</span>
        <span v-if="product.size">尺寸: {{ product.size }}</span>
        <span v-if="product.color">颜色: {{ product.color }}</span>
      </div>
      <div class="product-sales">
        已售 {{ product.sales }} 件 | 库存 {{ product.stock }} 件
      </div>
    </div>

    <!-- 商品描述 -->
    <div class="product-desc" v-if="product.description">
      <h3>商品详情</h3>
      <p>{{ product.description }}</p>
    </div>

    <!-- 评价 -->
    <div class="product-reviews">
      <h3>商品评价 ({{ reviews.length }})</h3>
      <div v-if="reviews.length > 0" class="reviews-list">
        <div v-for="review in reviews.slice(0, 3)" :key="review.id" class="review-item">
          <div class="review-header">
            <span class="review-user">{{ review.username }}</span>
            <span class="review-rating">
              <span v-for="i in 5" :key="i" class="star">{{ i <= review.rating ? '★' : '☆' }}</span>
            </span>
          </div>
          <p class="review-content">{{ review.content }}</p>
          <div class="review-date">{{ formatDate(review.created_at) }}</div>
        </div>
      </div>
      <div v-else class="no-reviews">
        暂无评价
      </div>
    </div>

    <!-- 底部操作栏 -->
    <div class="action-bar">
      <button class="btn-favorite" @click="toggleFavorite">
        <span>{{ isFavorite ? '❤️' : '🤍' }}</span>
        <span>{{ isFavorite ? '已收藏' : '收藏' }}</span>
      </button>
      <button class="btn-cart" @click="addToCart" :disabled="product.stock === 0">
        加入购物车
      </button>
      <button class="btn-buy" @click="buyNow" :disabled="product.stock === 0">
        {{ product.stock === 0 ? '已售罄' : '立即购买' }}
      </button>
    </div>

    <!-- 数量选择弹窗 -->
    <div v-if="showQuantityModal" class="modal-overlay" @click.self="showQuantityModal = false">
      <div class="quantity-modal">
        <div class="modal-header">
          <span>{{ isBuyNow ? '立即购买' : '加入购物车' }}</span>
          <button class="modal-close" @click="showQuantityModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="product-preview">
            <img :src="product.main_image || '/placeholder.png'" />
            <div>
              <div class="preview-price">¥{{ product.price.toFixed(2) }}</div>
              <div class="preview-stock">库存 {{ product.stock }} 件</div>
            </div>
          </div>
          <div class="quantity-selector">
            <span>数量</span>
            <div class="quantity-controls">
              <button @click="quantity > 1 && quantity--">-</button>
              <input v-model.number="quantity" type="number" min="1" :max="product.stock" />
              <button @click="quantity < product.stock && quantity++">+</button>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary btn-block" @click="confirmAction">
            确定
          </button>
        </div>
      </div>
    </div>
  </div>

  <div v-else class="loading">
    <div class="loading-spinner"></div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { publicAPI, cartAPI, favoriteAPI, reviewAPI } from '../../api'
import { useAuthStore } from '../../store/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const product = ref(null)
const reviews = ref([])
const isFavorite = ref(false)

const showQuantityModal = ref(false)
const quantity = ref(1)
const isBuyNow = ref(false)

const fetchProduct = async () => {
  try {
    const id = route.params.id
    product.value = await publicAPI.getProduct(id)

    // 获取评价
    const result = await reviewAPI.getProductReviews(id, { page: 1, page_size: 10 })
    reviews.value = result.data || []

    // 检查收藏状态
    if (authStore.isAuthenticated) {
      const favResult = await favoriteAPI.check(id)
      isFavorite.value = favResult.is_favorite
    }
  } catch (error) {
    console.error('获取商品详情失败:', error)
  }
}

const toggleFavorite = async () => {
  if (!authStore.isAuthenticated) {
    router.push('/login')
    return
  }

  try {
    if (isFavorite.value) {
      await favoriteAPI.remove(product.value.id)
    } else {
      await favoriteAPI.add(product.value.id)
    }
    isFavorite.value = !isFavorite.value
  } catch (error) {
    alert(error.message)
  }
}

const addToCart = () => {
  if (!authStore.isAuthenticated) {
    router.push('/login')
    return
  }

  isBuyNow.value = false
  quantity.value = 1
  showQuantityModal.value = true
}

const buyNow = () => {
  if (!authStore.isAuthenticated) {
    router.push('/login')
    return
  }

  isBuyNow.value = true
  quantity.value = 1
  showQuantityModal.value = true
}

const confirmAction = async () => {
  try {
    await cartAPI.add({
      product_id: product.value.id,
      quantity: quantity.value
    })

    showQuantityModal.value = false

    if (isBuyNow.value) {
      router.push('/shop/cart')
    } else {
      alert('已添加到购物车')
    }
  } catch (error) {
    alert(error.message)
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

onMounted(() => {
  fetchProduct()
})
</script>

<style scoped>
.product-detail {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 70px;
}

.product-gallery {
  background: #fff;
}

.main-image {
  width: 100%;
  height: 375px;
  object-fit: cover;
}

.product-info {
  background: #fff;
  padding: 16px;
  margin-bottom: 10px;
}

.product-name {
  font-size: 18px;
  font-weight: 500;
  margin-bottom: 8px;
}

.product-price {
  margin-bottom: 12px;
}

.current-price {
  font-size: 24px;
  color: var(--error-color);
  font-weight: 600;
}

.original-price {
  font-size: 14px;
  color: var(--text-secondary);
  text-decoration: line-through;
  margin-left: 8px;
}

.product-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.product-sales {
  font-size: 12px;
  color: var(--text-secondary);
}

.product-desc {
  background: #fff;
  padding: 16px;
  margin-bottom: 10px;
}

.product-desc h3 {
  font-size: 16px;
  margin-bottom: 12px;
}

.product-desc p {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.product-reviews {
  background: #fff;
  padding: 16px;
}

.product-reviews h3 {
  font-size: 16px;
  margin-bottom: 12px;
}

.review-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.review-item:last-child {
  border-bottom: none;
}

.review-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.review-user {
  font-weight: 500;
}

.review-rating .star {
  color: #faad14;
}

.review-content {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.review-date {
  font-size: 12px;
  color: #999;
}

.no-reviews {
  text-align: center;
  color: var(--text-secondary);
  padding: 20px 0;
}

.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: #fff;
  border-top: 1px solid var(--border-color);
  padding: 10px 16px;
  z-index: 100;
}

.btn-favorite {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 60px;
  background: none;
  border: none;
  font-size: 12px;
  cursor: pointer;
}

.btn-cart,
.btn-buy {
  flex: 1;
  padding: 12px;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  cursor: pointer;
  margin-left: 10px;
}

.btn-cart {
  background: #ffe4c4;
  color: var(--primary-color);
}

.btn-buy {
  background: var(--primary-color);
  color: #fff;
}

.btn-buy:disabled,
.btn-cart:disabled {
  background: #ccc;
  color: #999;
}

/* 数量选择弹窗 */
.quantity-modal {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: #fff;
  border-radius: 16px 16px 0 0;
  z-index: 1001;
}

.quantity-modal .modal-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.quantity-modal .modal-body {
  padding: 16px;
}

.product-preview {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.product-preview img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
}

.preview-price {
  font-size: 18px;
  color: var(--error-color);
  font-weight: 600;
}

.preview-stock {
  font-size: 12px;
  color: var(--text-secondary);
}

.quantity-selector {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.quantity-controls {
  display: flex;
  align-items: center;
}

.quantity-controls button {
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-color);
  background: #fff;
  cursor: pointer;
}

.quantity-controls input {
  width: 50px;
  height: 32px;
  text-align: center;
  border: 1px solid var(--border-color);
  border-left: none;
  border-right: none;
}

.quantity-modal .modal-footer {
  padding: 16px;
}

.btn-block {
  width: 100%;
}
</style>
