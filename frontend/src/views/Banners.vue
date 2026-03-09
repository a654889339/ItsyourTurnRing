<template>
  <div class="banners-page">
    <div class="page-header">
      <h2 class="page-title">轮播图管理</h2>
      <button class="btn btn-primary" @click="openCreateModal">+ 添加轮播图</button>
    </div>

    <div class="card">
      <div class="banner-list">
        <div v-for="banner in banners" :key="banner.id" class="banner-item">
          <img :src="banner.image" class="banner-image" />
          <div class="banner-info">
            <div class="banner-title">{{ banner.title || '无标题' }}</div>
            <div class="banner-link" v-if="banner.link">链接: {{ banner.link }}</div>
            <div class="banner-meta">
              <span class="status-tag" :class="'status-' + banner.status">
                {{ banner.status === 'active' ? '启用' : '禁用' }}
              </span>
              <span>排序: {{ banner.sort_order }}</span>
            </div>
          </div>
          <div class="banner-actions">
            <button class="btn btn-small btn-secondary" @click="openEditModal(banner)">编辑</button>
            <button class="btn btn-small btn-secondary" @click="toggleStatus(banner)">
              {{ banner.status === 'active' ? '禁用' : '启用' }}
            </button>
            <button class="btn btn-small btn-danger" @click="confirmDelete(banner)">删除</button>
          </div>
        </div>
      </div>

      <div v-if="banners.length === 0" class="empty-state">
        暂无轮播图
      </div>
    </div>

    <!-- 表单弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal" style="max-width: 500px">
        <div class="modal-header">
          <h3 class="modal-title">{{ isEdit ? '编辑轮播图' : '添加轮播图' }}</h3>
          <button class="modal-close" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">标题</label>
            <input v-model="form.title" type="text" class="form-input" placeholder="可选" />
          </div>

          <div class="form-group">
            <label class="form-label">图片 *</label>
            <div class="image-upload">
              <input type="file" accept="image/*" @change="handleImageUpload" ref="fileInput" style="display: none" />
              <div class="image-preview" v-if="form.image">
                <img :src="form.image" class="preview-img" />
                <button class="remove-image" @click="form.image = ''">&times;</button>
              </div>
              <button class="btn btn-secondary" @click="$refs.fileInput.click()">
                {{ form.image ? '更换图片' : '上传图片' }}
              </button>
            </div>
            <p class="form-hint">建议尺寸: 750x300</p>
          </div>

          <div class="form-group">
            <label class="form-label">链接</label>
            <input v-model="form.link" type="text" class="form-input" placeholder="点击跳转链接,可选" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">排序</label>
              <input v-model.number="form.sort_order" type="number" class="form-input" />
            </div>
            <div class="form-group">
              <label class="form-label">状态</label>
              <select v-model="form.status" class="form-input form-select">
                <option value="active">启用</option>
                <option value="inactive">禁用</option>
              </select>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="saveBanner" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { bannerAPI, uploadAPI } from '../api'

const banners = ref([])

const showModal = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const editingId = ref(null)

const form = reactive({
  title: '',
  image: '',
  link: '',
  sort_order: 0,
  status: 'active'
})

const fetchBanners = async () => {
  try {
    banners.value = await bannerAPI.list()
  } catch (error) {
    console.error('获取轮播图失败:', error)
  }
}

const openCreateModal = () => {
  isEdit.value = false
  editingId.value = null
  Object.assign(form, {
    title: '',
    image: '',
    link: '',
    sort_order: 0,
    status: 'active'
  })
  showModal.value = true
}

const openEditModal = (banner) => {
  isEdit.value = true
  editingId.value = banner.id
  Object.assign(form, {
    title: banner.title,
    image: banner.image,
    link: banner.link,
    sort_order: banner.sort_order,
    status: banner.status
  })
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const saveBanner = async () => {
  if (!form.image) {
    alert('请上传图片')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await bannerAPI.update(editingId.value, form)
    } else {
      await bannerAPI.create(form)
    }
    closeModal()
    fetchBanners()
  } catch (error) {
    alert(error.message)
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (banner) => {
  const newStatus = banner.status === 'active' ? 'inactive' : 'active'
  try {
    await bannerAPI.update(banner.id, { ...banner, status: newStatus })
    fetchBanners()
  } catch (error) {
    alert(error.message)
  }
}

const confirmDelete = async (banner) => {
  if (!confirm('确定要删除这个轮播图吗？')) return

  try {
    await bannerAPI.delete(banner.id)
    fetchBanners()
  } catch (error) {
    alert(error.message)
  }
}

const handleImageUpload = async (e) => {
  const file = e.target.files[0]
  if (!file) return

  try {
    const result = await uploadAPI.uploadImage(file)
    form.image = result.url
  } catch (error) {
    alert('上传失败: ' + error.message)
  }
}

onMounted(() => {
  fetchBanners()
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

.banner-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.banner-item {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
}

.banner-image {
  width: 200px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
}

.banner-info {
  flex: 1;
  margin-left: 20px;
}

.banner-title {
  font-weight: 500;
  margin-bottom: 4px;
}

.banner-link {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 300px;
}

.banner-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.banner-actions {
  display: flex;
  gap: 8px;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-row .form-group {
  flex: 1;
}

.image-upload {
  display: flex;
  align-items: center;
  gap: 16px;
}

.image-preview {
  position: relative;
  width: 200px;
  height: 80px;
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

.form-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.status-inactive {
  background-color: #f5f5f5;
  color: var(--text-secondary);
}
</style>
