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
                  <div class="product-media-count">
                    <span v-if="getImageCount(product)">{{ getImageCount(product) }}张图</span>
                    <span v-if="product.video" class="has-video">有视频</span>
                  </div>
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

      <div v-if="products.length === 0" class="empty-state">暂无商品</div>

      <div class="pagination" v-if="total > pageSize">
        <button class="pagination-btn" :disabled="page === 1" @click="page--; fetchProducts()">上一页</button>
        <span class="pagination-info">第 {{ page }} / {{ totalPages }} 页</span>
        <button class="pagination-btn" :disabled="page >= totalPages" @click="page++; fetchProducts()">下一页</button>
      </div>
    </div>

    <!-- 商品表单弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal" style="max-width: 720px">
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

          <!-- 多图上传 -->
          <div class="form-group">
            <label class="form-label">商品图片 <span class="form-hint-inline">(最多9张，第一张为主图)</span></label>
            <div class="multi-upload">
              <div class="upload-grid">
                <div
                  v-for="(img, idx) in imageList"
                  :key="idx"
                  class="upload-item"
                  :class="{ 'is-main': idx === 0 }"
                  draggable="true"
                  @dragstart="dragStart(idx)"
                  @dragover.prevent
                  @drop="dragDrop(idx)"
                >
                  <img :src="img" class="upload-thumb" />
                  <button class="upload-remove" @click="removeImage(idx)">&times;</button>
                  <span v-if="idx === 0" class="main-badge">主图</span>
                </div>
                <div
                  v-if="imageList.length < 9"
                  class="upload-add"
                  @click="$refs.imageInput.click()"
                >
                  <span class="upload-add-icon">+</span>
                  <span class="upload-add-text">添加图片</span>
                </div>
              </div>
              <input
                ref="imageInput"
                type="file"
                accept="image/*"
                multiple
                style="display: none"
                @change="handleMultiImageUpload"
              />
              <p class="form-hint" v-if="uploading.image">正在上传图片...</p>
            </div>
          </div>

          <!-- 视频上传 -->
          <div class="form-group">
            <label class="form-label">商品视频 <span class="form-hint-inline">(最大50MB，可选)</span></label>
            <div class="video-upload">
              <div v-if="form.video" class="video-preview">
                <video :src="form.video" controls class="video-player"></video>
                <button class="upload-remove" @click="form.video = ''">&times;</button>
              </div>
              <button
                v-if="!form.video"
                class="btn btn-secondary"
                @click="$refs.videoInput.click()"
                :disabled="uploading.video"
              >
                {{ uploading.video ? '上传中...' : '上传视频' }}
              </button>
              <button
                v-if="form.video"
                class="btn btn-secondary btn-small"
                @click="$refs.videoInput.click()"
              >
                更换视频
              </button>
              <input
                ref="videoInput"
                type="file"
                accept="video/*"
                style="display: none"
                @change="handleVideoUpload"
              />
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
          <button class="btn btn-primary" @click="saveProduct" :disabled="saving || uploading.image || uploading.video">
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

const filter = reactive({ keyword: '', category: '', status: '' })

const showModal = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const editingId = ref(null)
const imageList = ref([])
const uploading = reactive({ image: false, video: false })
let dragIdx = -1

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
  video: '',
  is_featured: false,
  is_new: false
})

const totalPages = computed(() => Math.ceil(total.value / pageSize))

function getImageCount(product) {
  if (!product.images) return 0
  try {
    return JSON.parse(product.images).length
  } catch { return product.main_image ? 1 : 0 }
}

const fetchProducts = async () => {
  try {
    const result = await productAPI.list({
      page: page.value, page_size: pageSize,
      keyword: filter.keyword, category: filter.category, status: filter.status
    })
    products.value = result.data || []
    total.value = result.total || 0
  } catch (error) {
    console.error('获取商品列表失败:', error)
  }
}

const fetchCategories = async () => {
  try { categories.value = await categoryAPI.list() }
  catch (error) { console.error('获取分类失败:', error) }
}

const getStatusText = (status) => {
  return { available: '在售', sold_out: '售罄', disabled: '下架' }[status] || status
}

function parseImages(imagesStr) {
  if (!imagesStr) return []
  try { return JSON.parse(imagesStr) }
  catch { return imagesStr ? [imagesStr] : [] }
}

const openCreateModal = () => {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, {
    name: '', category_id: categories.value[0]?.id || null,
    description: '', price: 0, original_price: 0, stock: 0,
    material: '', size: '', color: '', main_image: '', images: '', video: '',
    is_featured: false, is_new: false
  })
  imageList.value = []
  showModal.value = true
}

const openEditModal = (product) => {
  isEdit.value = true
  editingId.value = product.id
  Object.assign(form, {
    name: product.name, category_id: product.category_id,
    description: product.description, price: product.price,
    original_price: product.original_price, stock: product.stock,
    material: product.material, size: product.size, color: product.color,
    main_image: product.main_image, images: product.images,
    video: product.video || '', is_featured: product.is_featured, is_new: product.is_new
  })
  imageList.value = parseImages(product.images)
  if (imageList.value.length === 0 && product.main_image) {
    imageList.value = [product.main_image]
  }
  showModal.value = true
}

const closeModal = () => { showModal.value = false }

async function handleMultiImageUpload(e) {
  const files = Array.from(e.target.files)
  if (!files.length) return
  const remaining = 9 - imageList.value.length
  const toUpload = files.slice(0, remaining)

  uploading.image = true
  for (const file of toUpload) {
    if (!file.type.startsWith('image/')) continue
    if (file.size > 10 * 1024 * 1024) { alert(`${file.name} 超过10MB`); continue }
    try {
      const result = await uploadAPI.uploadImage(file)
      imageList.value.push(result.url)
    } catch (err) {
      alert(`上传 ${file.name} 失败: ${err.message}`)
    }
  }
  uploading.image = false
  e.target.value = ''
}

function removeImage(idx) {
  imageList.value.splice(idx, 1)
}

function dragStart(idx) { dragIdx = idx }
function dragDrop(idx) {
  if (dragIdx === idx) return
  const moved = imageList.value.splice(dragIdx, 1)[0]
  imageList.value.splice(idx, 0, moved)
  dragIdx = -1
}

async function handleVideoUpload(e) {
  const file = e.target.files[0]
  if (!file) return
  if (file.size > 50 * 1024 * 1024) { alert('视频大小不能超过50MB'); return }

  uploading.video = true
  try {
    const result = await uploadAPI.uploadFile(file)
    form.video = result.url
  } catch (err) {
    alert('视频上传失败: ' + err.message)
  } finally {
    uploading.video = false
    e.target.value = ''
  }
}

function syncImageFields() {
  form.images = imageList.value.length > 0 ? JSON.stringify(imageList.value) : ''
  form.main_image = imageList.value[0] || ''
}

const saveProduct = async () => {
  if (!form.name || !form.category_id || form.price <= 0) {
    alert('请填写必填项'); return
  }
  syncImageFields()

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
  } catch (error) { alert(error.message) }
}

const confirmDelete = async (product) => {
  if (!confirm(`确定要删除商品「${product.name}」吗？`)) return
  try { await productAPI.delete(product.id); fetchProducts() }
  catch (error) { alert(error.message) }
}

let searchTimer = null
const debouncedSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchProducts() }, 300)
}

onMounted(() => { fetchCategories(); fetchProducts() })
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-title { font-size: 24px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.product-cell { display: flex; align-items: center; }
.product-info { margin-left: 12px; }
.product-name { font-weight: 500; }
.product-material { font-size: 12px; color: var(--text-secondary); }
.product-media-count { font-size: 11px; color: #999; margin-top: 2px; display: flex; gap: 6px; }
.has-video { color: var(--primary-color); font-weight: 500; }
.low-stock { color: var(--error-color); font-weight: 500; }
.action-btns { display: flex; gap: 8px; }
.form-row { display: flex; gap: 16px; margin-bottom: 16px; }
.form-row .form-group { flex: 1; margin-bottom: 0; }
.form-hint-inline { font-size: 12px; color: #999; font-weight: normal; }
.form-hint { font-size: 12px; color: var(--primary-color); margin-top: 4px; }
.checkbox-label { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.pagination-info { padding: 0 16px; color: var(--text-secondary); }

/* Multi-image upload */
.upload-grid {
  display: grid; grid-template-columns: repeat(auto-fill, 100px); gap: 10px;
}
.upload-item {
  position: relative; width: 100px; height: 100px;
  border-radius: 6px; overflow: hidden; cursor: grab;
  border: 2px solid transparent; transition: border-color 0.2s;
}
.upload-item.is-main { border-color: var(--primary-color); }
.upload-item:hover { border-color: #ccc; }
.upload-thumb { width: 100%; height: 100%; object-fit: cover; }
.upload-remove {
  position: absolute; top: -1px; right: -1px; width: 22px; height: 22px;
  border-radius: 50%; background: var(--error-color); color: #fff;
  border: none; cursor: pointer; font-size: 14px; line-height: 20px;
  display: flex; align-items: center; justify-content: center;
}
.main-badge {
  position: absolute; bottom: 0; left: 0; right: 0;
  background: rgba(212, 165, 116, 0.9); color: #fff; font-size: 10px;
  text-align: center; padding: 1px 0;
}
.upload-add {
  width: 100px; height: 100px; border: 2px dashed #ddd;
  border-radius: 6px; display: flex; flex-direction: column;
  align-items: center; justify-content: center; cursor: pointer;
  transition: border-color 0.2s; gap: 4px;
}
.upload-add:hover { border-color: var(--primary-color); }
.upload-add-icon { font-size: 28px; color: #bbb; line-height: 1; }
.upload-add-text { font-size: 11px; color: #999; }

/* Video upload */
.video-upload { display: flex; align-items: flex-start; gap: 12px; flex-wrap: wrap; }
.video-preview {
  position: relative; width: 240px; border-radius: 8px; overflow: hidden;
  background: #000; box-shadow: 0 2px 8px rgba(0,0,0,0.12);
}
.video-player { width: 100%; max-height: 160px; display: block; }
.video-preview .upload-remove { top: 4px; right: 4px; }
</style>
