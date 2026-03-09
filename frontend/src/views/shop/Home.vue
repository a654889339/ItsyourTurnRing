<template>
  <div class="shop-home">
    <!-- 头部 -->
    <header class="shop-header">
      <h1>RingShop</h1>
      <div class="search-bar">
        <input v-model="searchKeyword" type="text" placeholder="搜索商品..." @keyup.enter="handleSearch" />
      </div>
    </header>

    <!-- 轮播图 -->
    <div class="banner-carousel" v-if="banners.length > 0">
      <div class="banner-slider" :style="{ transform: `translateX(-${bannerIndex * 100}%)` }">
        <div v-for="banner in banners" :key="banner.id" class="banner-slide">
          <img :src="banner.image" :alt="banner.title" @click="handleBannerClick(banner)" />
        </div>
      </div>
      <div class="banner-dots">
        <span
          v-for="(_, index) in banners"
          :key="index"
          class="dot"
          :class="{ active: index === bannerIndex }"
          @click="bannerIndex = index"
        ></span>
      </div>
    </div>

    <!-- 分类导航 -->
    <div class="category-nav">
      <div
        v-for="cat in categories"
        :key="cat.id"
        class="category-item"
        :class="{ active: selectedCategory === cat.code }"
        @click="selectCategory(cat.code)"
      >
        <div class="category-icon">{{ getCategoryIcon(cat.code) }}</div>
        <div class="category-name">{{ cat.name }}</div>
      </div>
    </div>

    <!-- 推荐商品 -->
    <section class="product-section" v-if="featuredProducts.length > 0">
      <h2 class="section-title">精选推荐</h2>
      <div class="product-grid">
        <div
          v-for="product in featuredProducts"
          :key="product.id"
          class="product-card"
          @click="goToProduct(product.id)"
        >
          <img :src="product.main_image || '/placeholder.png'" class="product-image" />
          <div class="product-info">
            <div class="product-name">{{ product.name }}</div>
            <div class="product-price">
              <span class="current-price">¥{{ product.price.toFixed(2) }}</span>
              <span v-if="product.original_price" class="original-price">
                ¥{{ product.original_price.toFixed(2) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 新品上市 -->
    <section class="product-section" v-if="newProducts.length > 0">
      <h2 class="section-title">新品上市</h2>
      <div class="product-grid">
        <div
          v-for="product in newProducts"
          :key="product.id"
          class="product-card"
          @click="goToProduct(product.id)"
        >
          <img :src="product.main_image || '/placeholder.png'" class="product-image" />
          <div class="product-badge new">新品</div>
          <div class="product-info">
            <div class="product-name">{{ product.name }}</div>
            <div class="product-price">
              <span class="current-price">¥{{ product.price.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 全部商品 -->
    <section class="product-section">
      <h2 class="section-title">全部商品</h2>
      <div class="product-grid">
        <div
          v-for="product in products"
          :key="product.id"
          class="product-card"
          @click="goToProduct(product.id)"
        >
          <img :src="product.main_image || '/placeholder.png'" class="product-image" />
          <div class="product-info">
            <div class="product-name">{{ product.name }}</div>
            <div class="product-material" v-if="product.material">{{ product.material }}</div>
            <div class="product-price">
              <span class="current-price">¥{{ product.price.toFixed(2) }}</span>
              <span v-if="product.original_price" class="original-price">
                ¥{{ product.original_price.toFixed(2) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="hasMore" class="load-more">
        <button class="btn btn-secondary" @click="loadMore" :disabled="loading">
          {{ loading ? '加载中...' : '加载更多' }}
        </button>
      </div>
    </section>

    <!-- 底部导航 -->
    <nav class="bottom-nav">
      <router-link to="/shop" class="nav-item active">
        <span class="nav-icon">🏠</span>
        <span>首页</span>
      </router-link>
      <router-link to="/shop/cart" class="nav-item">
        <span class="nav-icon">🛒</span>
        <span>购物车</span>
      </router-link>
      <router-link to="/shop/orders" class="nav-item">
        <span class="nav-icon">📦</span>
        <span>订单</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { publicAPI } from '../../api'

const router = useRouter()

const banners = ref([])
const categories = ref([])
const featuredProducts = ref([])
const newProducts = ref([])
const products = ref([])

const bannerIndex = ref(0)
const selectedCategory = ref('')
const searchKeyword = ref('')

const page = ref(1)
const hasMore = ref(true)
const loading = ref(false)

let bannerTimer = null

const getCategoryIcon = (code) => {
  const icons = {
    bracelet: '📿',
    necklace: '📿',
    accessory: '✨',
    ring: '💍',
    earring: '👂'
  }
  return icons[code] || '💎'
}

const fetchHomeData = async () => {
  try {
    const data = await publicAPI.getHome()
    banners.value = data.banners || []
    categories.value = data.categories || []
    featuredProducts.value = data.featured || []
    newProducts.value = data.new || []

    startBannerAutoPlay()
  } catch (error) {
    console.error('获取首页数据失败:', error)
  }
}

const fetchProducts = async () => {
  if (loading.value) return

  loading.value = true
  try {
    const result = await publicAPI.getProducts({
      page: page.value,
      page_size: 10,
      category: selectedCategory.value,
      keyword: searchKeyword.value
    })
    const newData = result.data || []

    if (page.value === 1) {
      products.value = newData
    } else {
      products.value = [...products.value, ...newData]
    }

    hasMore.value = products.value.length < result.total
  } catch (error) {
    console.error('获取商品列表失败:', error)
  } finally {
    loading.value = false
  }
}

const selectCategory = (code) => {
  if (selectedCategory.value === code) {
    selectedCategory.value = ''
  } else {
    selectedCategory.value = code
  }
  page.value = 1
  fetchProducts()
}

const handleSearch = () => {
  page.value = 1
  fetchProducts()
}

const loadMore = () => {
  page.value++
  fetchProducts()
}

const goToProduct = (id) => {
  router.push(`/shop/product/${id}`)
}

const handleBannerClick = (banner) => {
  if (banner.link) {
    window.location.href = banner.link
  }
}

const startBannerAutoPlay = () => {
  if (banners.value.length <= 1) return

  bannerTimer = setInterval(() => {
    bannerIndex.value = (bannerIndex.value + 1) % banners.value.length
  }, 4000)
}

onMounted(() => {
  fetchHomeData()
  fetchProducts()
})

onUnmounted(() => {
  if (bannerTimer) {
    clearInterval(bannerTimer)
  }
})
</script>

<style scoped>
.shop-home {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 70px;
}

.shop-header {
  background: linear-gradient(135deg, #d4a574 0%, #8b7355 100%);
  padding: 16px;
  color: #fff;
}

.shop-header h1 {
  font-size: 20px;
  margin-bottom: 12px;
}

.search-bar input {
  width: 100%;
  padding: 10px 16px;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  outline: none;
}

.banner-carousel {
  position: relative;
  overflow: hidden;
}

.banner-slider {
  display: flex;
  transition: transform 0.3s ease;
}

.banner-slide {
  min-width: 100%;
}

.banner-slide img {
  width: 100%;
  height: 150px;
  object-fit: cover;
}

.banner-dots {
  position: absolute;
  bottom: 10px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 6px;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
}

.dot.active {
  background: #fff;
}

.category-nav {
  display: flex;
  overflow-x: auto;
  background: #fff;
  padding: 12px;
  gap: 16px;
}

.category-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 60px;
  cursor: pointer;
}

.category-item.active .category-name {
  color: var(--primary-color);
}

.category-icon {
  font-size: 28px;
  margin-bottom: 4px;
}

.category-name {
  font-size: 12px;
  color: var(--text-secondary);
}

.product-section {
  padding: 16px;
}

.section-title {
  font-size: 16px;
  margin-bottom: 12px;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.product-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.product-image {
  width: 100%;
  height: 150px;
  object-fit: cover;
}

.product-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  color: #fff;
}

.product-badge.new {
  background: #52c41a;
}

.product-info {
  padding: 10px;
}

.product-name {
  font-size: 14px;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-material {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.product-price {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.current-price {
  color: var(--error-color);
  font-weight: 600;
}

.original-price {
  font-size: 12px;
  color: var(--text-secondary);
  text-decoration: line-through;
}

.load-more {
  text-align: center;
  margin-top: 16px;
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
