<template>
  <div class="knowledge-list">
    <el-row :gutter="20">
      <!-- 左侧分类 -->
      <el-col :span="5">
        <el-card class="category-card">
          <template #header>
            <span>资料分类</span>
          </template>
          <el-menu :default-active="currentCategory" @select="handleCategoryChange">
            <el-menu-item index="">
              <el-icon><Files /></el-icon>全部资料
            </el-menu-item>
            <el-menu-item v-for="cat in categories" :key="cat.id" :index="String(cat.id)">
              <el-icon><Folder /></el-icon>{{ cat.name }}
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <!-- 右侧内容 -->
      <el-col :span="19">
        <!-- 搜索栏 -->
        <el-card class="search-card">
          <el-form :inline="true" class="search-form">
            <el-form-item label="关键词">
              <el-input v-model="keyword" placeholder="请输入标题、关键词或描述" clearable @keyup.enter="handleSearch" style="width: 150px;">
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
            </el-form-item>
            <el-form-item label="上传人">
              <el-select v-model="uploadedBy" placeholder="请选择上传人" clearable @change="handleSearch" style="width: 150px;">
                <el-option v-for="user in users" :key="user.id" :label="user.name" :value="user.id" />
              </el-select>
            </el-form-item>
            <el-form-item class="search-button-item">
              <el-button type="primary" @click="handleSearch">搜索</el-button>
              <el-button @click="resetSearch">重置</el-button>
            </el-form-item>
            <el-form-item v-if="userStore.canManageKnowledge">
              <el-button type="primary" @click="showUploadDialog">
                <el-icon><Upload /></el-icon> 上传资料
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 资料列表 -->
        <el-card>
          <el-table v-loading="loading" :data="items" stripe>
            <el-table-column prop="title" label="资料名称" min-width="250" header-align="center" align="center">
              <template #default="{ row }">
                <el-link type="primary" @click="$router.push(`/knowledge/${row.id}`)">{{ row.title }}</el-link>
              </template>
            </el-table-column>
            <el-table-column prop="category" label="分类" width="120" header-align="center" align="center">
              <template #default="{ row }">{{ row.category?.name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="version" label="版本" width="80" header-align="center" align="center" />
            <el-table-column prop="view_count" label="查看" width="80" header-align="center" align="center" />
            <el-table-column prop="download_count" label="下载" width="80" header-align="center" align="center" />
            <el-table-column prop="uploader" label="上传人" width="100" header-align="center" align="center">
              <template #default="{ row }">{{ row.uploader?.name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="created_at" label="上传时间" width="160" header-align="center" align="center">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="150" header-align="center" align="center">
              <template #default="{ row }">
                <el-button type="primary" link @click="handleDownload(row)">下载</el-button>
                <el-button v-if="canDelete(row)" type="danger" link @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :total="pagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            class="pagination"
            @size-change="fetchItems"
            @current-change="fetchItems"
          />
        </el-card>
      </el-col>
    </el-row>

    <!-- 上传弹窗 -->
    <el-dialog v-model="uploadDialogVisible" title="上传资料" width="500px">
      <el-form :model="uploadForm" label-width="80px">
        <el-form-item label="资料标题">
          <el-input v-model="uploadForm.title" placeholder="请输入标题" />
        </el-form-item>
        <el-form-item label="所属分类">
          <el-select v-model="uploadForm.category_id" placeholder="请选择分类">
            <el-option
              v-for="cat in categories"
              :key="cat.id"
              :label="cat.name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="uploadForm.keywords" placeholder="多个关键词用逗号分隔" />
        </el-form-item>
        <el-form-item label="资料描述">
          <el-input v-model="uploadForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="选择文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getKnowledgeList, uploadKnowledge, downloadKnowledge, deleteKnowledge } from '@/api/knowledge'
import { getUsers } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'

const userStore = useUserStore()

const loading = ref(false)
const items = ref([])
// 前端固定资料分类，ID 与后端数据一一对应
const categories = ref([
  { id: 1, name: '项目资料' },
  { id: 3, name: '技术规范' },
  { id: 4, name: '案例资料' },
  { id: 5, name: '培训材料' },
  { id: 2, name: '政策文件' }
])
const users = ref([]) // 用户列表
const currentCategory = ref('')
const keyword = ref('')
const uploadedBy = ref('') // 上传人筛选
const uploadDialogVisible = ref(false)
const uploading = ref(false)
const uploadRef = ref(null)
const selectedFile = ref(null)

const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const uploadForm = reactive({ title: '', category_id: null, keywords: '', description: '' })

const formatDateTime = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'

// 检查是否可以删除：管理员/部门经理可以删除所有，其他用户只能删除自己上传的
const canDelete = (item) => {
  if (userStore.isAdmin || userStore.isDeptManager) return true
  return item.uploaded_by === userStore.userId
}

const fetchUsers = async () => {
  const res = await getUsers({ page: 1, page_size: 100 })
  users.value = res.data?.list || []
}

const fetchItems = async () => {
  loading.value = true
  try {
    const res = await getKnowledgeList({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: keyword.value,
      category_id: currentCategory.value,
      uploaded_by: uploadedBy.value
    })
    items.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleCategoryChange = (index) => {
  currentCategory.value = index
  pagination.page = 1
  fetchItems()
}

const handleSearch = () => {
  pagination.page = 1
  fetchItems()
}

const resetSearch = () => {
  keyword.value = ''
  uploadedBy.value = ''
  currentCategory.value = ''
  pagination.page = 1
  fetchItems()
}

const showUploadDialog = () => {
  Object.assign(uploadForm, { title: '', category_id: null, keywords: '', description: '' })
  selectedFile.value = null
  uploadDialogVisible.value = true
}

const handleFileChange = (file) => {
  selectedFile.value = file.raw
  if (!uploadForm.title) {
    uploadForm.title = file.name
  }
}

const handleUpload = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请选择文件')
    return
  }
  
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    formData.append('title', uploadForm.title)
    if (uploadForm.category_id) formData.append('category_id', uploadForm.category_id)
    if (uploadForm.keywords) formData.append('keywords', uploadForm.keywords)
    if (uploadForm.description) formData.append('description', uploadForm.description)
    
    await uploadKnowledge(formData)
    ElMessage.success('上传成功')
    uploadDialogVisible.value = false
    fetchItems()
  } catch (error) {
    console.error('上传失败:', error)
  } finally {
    uploading.value = false
  }
}

const handleDownload = async (row) => {
  try {
    const response = await downloadKnowledge(row.id)
    
    // 从响应头获取文件名，如果没有则使用标题
    let filename = row.title || '资料下载'
    const contentDisposition = response.headers['content-disposition']
    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename=(.+)/)
      if (filenameMatch && filenameMatch[1]) {
        filename = decodeURIComponent(filenameMatch[1])
      }
    }
    
    const blob = new Blob([response.data])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (error) {
    console.error('下载失败详情:', error)
    console.error('错误响应:', error.response)
    
    // 尝试解析错误消息
    let errorMsg = '下载失败，请稍后重试'
    if (error.response) {
      if (error.response.data) {
        // 如果响应是Blob，尝试转换为JSON
        if (error.response.data instanceof Blob) {
          try {
            const text = await error.response.data.text()
            const jsonData = JSON.parse(text)
            errorMsg = jsonData.message || errorMsg
          } catch (e) {
            // Blob转换失败，使用默认消息
          }
        } else if (error.response.data.message) {
          errorMsg = error.response.data.message
        }
      }
      errorMsg = `${errorMsg} (${error.response.status})`
    }
    
    ElMessage.error(errorMsg)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除资料「${row.title}」吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteKnowledge(row.id)
    ElMessage.success('删除成功')
    fetchItems()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

onMounted(() => {
  fetchUsers()
  fetchItems()
})
</script>

<style scoped>
.category-card { position: sticky; top: 20px; }
.search-card { margin-bottom: 20px; }
.search-form {
  display: flex;
  flex-wrap: nowrap;
  align-items: flex-end;
}
.search-form .el-form-item {
  margin-bottom: 0;
  margin-right: 12px;
}
.search-button-item {
  margin-left: auto;
  margin-bottom: 0 !important;
}
.pagination { margin-top: 20px; justify-content: flex-end; }
</style>
