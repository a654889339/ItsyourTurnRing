<template>
  <div class="products-page">
    <div class="page-header">
      <h2 class="page-title">商品管理</h2>
      <button class="btn btn-primary" @click="openCreateModal">+ 添加商品</button>
    </div>

    <div class="card">
      <div class="filter-bar">
        <input
          v-model="filter.keyword"
          type="text"
          class="form-input"
          placeholder="搜索商品名称..."
          style="width: 200px"
          @input="debouncedSearch"
        />
        <select v-model="filter.category" class="form-input form-select" style="width: 150px" @change="fetchProducts">
          <option value="">全部分类</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.code">{{ cat.name }}</option>
        </select>
        <select v-model="filter.status" class="form-input form-select" style="width: 120px" @change="fetchProducts">
          <option value="">全部状态</option>
          <option value="available">在售</option>
          <option value="sold_out">售罄</option>
          <option value="disabled">下架</option>
        </select>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>商品</th>
            <th>分类</th>
            <th>价格</th>
            <th>库存</th>
            <th>销量</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="product in products" :key="product.id">
            <td>
              <div class="product-cell">
                <img :src="product.main_image || '/placeholder.png'" class="product-image" />
                <div class="product-info">
                  <div class="product-name">{{ product.name }}</div>
                  <div class="product-material" v-if="product.material">{{ product.material }}</div>
                </div>
              </div>
            </td>
            <td>{{ product.category_name }}</td>
            <td>
              <span class="price">¥{{ product.price.toFixed(2) }}</span>
              <span v-if="product.original_price" class="price-original">¥{{ product.original_price.toFixed(2) }}</span>
            </td>
            <td :class="{ 'low-stock': product.stock < 10 }">{{ product.stock }}</td>
            <td>{{ product.sales }}</td>
            <td>
              <span class="status-tag" :class="'status-' + product.status">
                {{ getStatusText(product.status) }}
              </span>
            </td>
            <td>
              <div class="action-btns">
                <button class="btn btn-small btn-secondary" @click="openEditModal(product)">编辑</button>
                <button class="btn btn-small btn-secondary" @click="toggleStatus(product)">
                  {{ product.status === 'available' ? '下架' : '上架' }}
                </button>
                <button class="btn btn-small btn-danger" @click="confirmDelete(product)">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="products.length === 0" class="empty-state">
        暂无商品
      </div>

      <div class="pagination" v-if="total > pageSize">
        <button class="pagination-btn" :disabled="page === 1" @click="page--; fetchProducts()">上一页</button>
        <span class="pagination-info">第 {{ page }} / {{ totalPages }} 页</span>
        <button class="pagination-btn" :disabled="page >= totalPages" @click="page++; fetchProducts()">下一页</button>
      </div>
    </div>

    <!-- 商品表单弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal" style="max-width: 700px">
        <div class="modal-header">
          <h3 class="modal-title">{{ isEdit ? '编辑商品' : '添加商品' }}</h3>
          <button class="modal-close" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group" style="flex: 1">
              <label class="form-label">商品名称 *</label>
              <input v-model="form.name" type="text" class="form-input" required />
            </div>
            <div class="form-group" style="width: 150px">
              <label class="form-label">分类 *</label>
              <select v-model="form.category_id" class="form-input form-select" required>
                <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">商品描述</label>
            <textarea v-model="form.description" class="form-input form-textarea" rows="3"></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">售价 *</label>
              <input v-model.number="form.price" type="number" step="0.01" class="form-input" required />
            </div>
            <div class="form-group">
              <label class="form-label">原价</label>
              <input v-model.number="form.original_price" type="number" step="0.01" class="form-input" />
            </div>
            <div class="form-group">
              <label class="form-label">库存 *</label>
              <input v-model.number="form.stock" type="number" class="form-input" required />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">材质</label>
              <input v-model="form.material" type="text" class="form-input" placeholder="如: 925银、18K金" />
            </div>
            <div class="form-group">
              <label class="form-label">尺寸</label>
              <input v-model="form.size" type="text" class="form-input" placeholder="如: 16cm-20cm" />
            </div>
            <div class="form-group">
              <label class="form-label">颜色</label>
              <input v-model="form.color" type="text" class="form-input" placeholder="如: 银色、金色" />
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">商品图片</label>
            <div class="image-upload">
              <input type="file" accept="image/*" @change="handleImageUpload" ref="fileInput" style="display: none" />
              <div class="image-preview" v-if="form.main_image">
                <img :src="form.main_image" class="preview-img" />
                <button class="remove-image" @click="form.main_image = ''">&times;</button>
              </div>
              <button class="btn btn-secondary" @click="$refs.fileInput.click()">
                {{ form.main_image ? '更换图片' : '上传图片' }}
              </button>
            </div>
          </div>

          <div class="form-row">
            <label class="checkbox-label">
              <input type="checkbox" v-model="form.is_featured" />
              <span>推荐商品</span>
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="form.is_new" />
              <span>新品</span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="saveProduct" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { productAPI, categoryAPI, uploadAPI } from '../api'

const products = ref([])
const categories = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

const filter = reactive({
  keyword: '',
  category: '',
  status: ''
})

const showModal = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const editingId = ref(null)

const form = reactive({
  name: '',
  category_id: null,
  description: '',
  price: 0,
  original_price: 0,
  stock: 0,
  material: '',
  size: '',
  color: '',
  main_image: '',
  images: '',
  is_featured: false,
  is_new: false
})

const totalPages = computed(() => Math.ceil(total.value / pageSize))

const fetchProducts = async () => {
  try {
    const result = await productAPI.list({
      page: page.value,
      page_size: pageSize,
      keyword: filter.keyword,
      category: filter.category,
      status: filter.status
    })
    products.value = result.data || []
    total.value = result.total || 0
  } catch (error) {
    console.error('获取商品列表失败:', error)
  }
}

const fetchCategories = async () => {
  try {
    categories.value = await categoryAPI.list()
  } catch (error) {
    console.error('获取分类失败:', error)
  }
}

const getStatusText = (status) => {
  const map = {
    available: '在售',
    sold_out: '售罄',
    disabled: '下架'
  }
  return map[status] || status
}

const openCreateModal = () => {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, {
    name: '',
    category_id: categories.value[0]?.id || null,
    description: '',
    price: 0,
    original_price: 0,
    stock: 0,
    material: '',
    size: '',
    color: '',
    main_image: '',
    images: '',
    is_featured: false,
    is_new: false
  })
  showModal.value = true
}

const openEditModal = (product) => {
  isEdit.value = true
  editingId.value = product.id
  Object.assign(form, {
    name: product.name,
    category_id: product.category_id,
    description: product.description,
    price: product.price,
    original_price: product.original_price,
    stock: product.stock,
    material: product.material,
    size: product.size,
    color: product.color,
    main_image: product.main_image,
    images: product.images,
    is_featured: product.is_featured,
    is_new: product.is_new
  })
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const saveProduct = async () => {
  if (!form.name || !form.category_id || form.price <= 0) {
    alert('请填写必填项')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await productAPI.update(editingId.value, form)
    } else {
      await productAPI.create(form)
    }
    closeModal()
    fetchProducts()
  } catch (error) {
    alert(error.message)
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (product) => {
  const newStatus = product.status === 'available' ? 'disabled' : 'available'
  try {
    await productAPI.update(product.id, { ...product, status: newStatus })
    fetchProducts()
  } catch (error) {
    alert(error.message)
  }
}

const confirmDelete = async (product) => {
  if (!confirm(`确定要删除商品「${product.name}」吗？`)) return

  try {
    await productAPI.delete(product.id)
    fetchProducts()
  } catch (error) {
    alert(error.message)
  }
}

const handleImageUpload = async (e) => {
  const file = e.target.files[0]
  if (!file) return

  try {
    const result = await uploadAPI.uploadImage(file)
    form.main_image = result.url
  } catch (error) {
    alert('上传失败: ' + error.message)
  }
}

let searchTimer = null
const debouncedSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchProducts()
  }, 300)
}

onMounted(() => {
  fetchCategories()
  fetchProducts()
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

.product-cell {
  display: flex;
  align-items: center;
}

.product-info {
  margin-left: 12px;
}

.product-name {
  font-weight: 500;
}

.product-material {
  font-size: 12px;
  color: var(--text-secondary);
}

.low-stock {
  color: var(--error-color);
  font-weight: 500;
}

.action-btns {
  display: flex;
  gap: 8px;
}

.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.form-row .form-group {
  flex: 1;
  margin-bottom: 0;
}

.image-upload {
  display: flex;
  align-items: center;
  gap: 16px;
}

.image-preview {
  position: relative;
  width: 100px;
  height: 100px;
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 4px;
}

.remove-image {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--error-color);
  color: #fff;
  border: none;
  cursor: pointer;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.pagination-info {
  padding: 0 16px;
  color: var(--text-secondary);
}
</style>
