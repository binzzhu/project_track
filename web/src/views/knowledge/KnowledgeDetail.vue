<template>
  <div v-loading="loading" class="knowledge-detail">
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ item.title }}</span>
          <div>
            <el-button type="success" @click="handlePreview">
              <el-icon><View /></el-icon> 预览
            </el-button>
            <el-button type="primary" @click="handleDownload">
              <el-icon><Download /></el-icon> 下载
            </el-button>
            <el-button v-if="canDelete" type="danger" @click="handleDelete">
              <el-icon><Delete /></el-icon> 删除
            </el-button>
          </div>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="所属分类">{{ item.category?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="当前版本">{{ item.version }}</el-descriptions-item>
        <el-descriptions-item label="关键词">{{ item.keywords || '-' }}</el-descriptions-item>
        <el-descriptions-item label="文件大小">{{ formatFileSize(item.file_size) }}</el-descriptions-item>
        <el-descriptions-item label="查看次数">{{ item.view_count }}</el-descriptions-item>
        <el-descriptions-item label="下载次数">{{ item.download_count }}</el-descriptions-item>
        <el-descriptions-item label="上传人">{{ item.uploader?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="上传时间">{{ formatDateTime(item.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="资料描述" :span="2">{{ item.description || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 版本历史 -->
    <el-card class="version-card">
      <template #header>
        <div class="card-header">
          <span>版本历史</span>
          <el-button v-if="userStore.canManageKnowledge" type="primary" size="small" @click="showVersionDialog">
            <el-icon><Upload /></el-icon> 上传新版本
          </el-button>
        </div>
      </template>
      <el-timeline>
        <el-timeline-item v-for="ver in versions" :key="ver.id" :timestamp="formatDateTime(ver.created_at)" placement="top">
          <div class="version-item">
            <span class="version-tag">v{{ ver.version }}</span>
            <span class="version-note">{{ ver.change_note || '无更新说明' }}</span>
          </div>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-if="versions.length === 0" description="暂无版本记录" :image-size="60" />
    </el-card>

    <!-- 上传新版本弹窗 -->
    <el-dialog v-model="versionDialogVisible" title="上传新版本" width="400px">
      <el-form :model="versionForm" label-width="80px">
        <el-form-item label="更新说明">
          <el-input v-model="versionForm.change_note" type="textarea" :rows="3" placeholder="请输入更新说明" />
        </el-form-item>
        <el-form-item label="选择文件">
          <el-upload :auto-upload="false" :limit="1" :on-change="handleVersionFileChange">
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="versionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUploadVersion">上传</el-button>
      </template>
    </el-dialog>

    <!-- 文件预览弹窗 -->
    <el-dialog v-model="previewVisible" title="文件预览" width="80%" top="5vh">
      <div v-if="previewType === 'image'" class="preview-container">
        <img :src="previewUrl" style="max-width: 100%; max-height: 70vh; display: block; margin: 0 auto;" />
      </div>
      <div v-else-if="previewType === 'pdf'" class="preview-container">
        <iframe :src="previewUrl" style="width: 100%; height: 70vh; border: none;"></iframe>
      </div>
      <div v-else-if="previewType === 'text'" class="preview-container">
        <pre style="max-height: 70vh; overflow: auto; padding: 20px; background: #f5f5f5; border-radius: 4px;">{{ previewContent }}</pre>
      </div>
      <div v-else class="preview-container">
        <el-result icon="warning" title="无法预览" sub-title="该文件类型不支持在线预览，请下载后查看">
          <template #extra>
            <el-button type="primary" @click="handleDownload">下载文件</el-button>
          </template>
        </el-result>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getKnowledge, getVersions, uploadNewVersion, downloadKnowledge, deleteKnowledge } from '@/api/knowledge'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const itemId = computed(() => route.params.id)

const loading = ref(false)
const item = ref({})
const versions = ref([])
const versionDialogVisible = ref(false)
const uploading = ref(false)
const versionFile = ref(null)
const versionForm = reactive({ change_note: '' })
const previewVisible = ref(false)
const previewUrl = ref('')
const previewType = ref('')
const previewContent = ref('')

const formatDateTime = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'
const formatFileSize = (bytes) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) { bytes /= 1024; i++ }
  return `${bytes.toFixed(1)} ${units[i]}`
}

// 检查是否可以删除：管理员/部门经理可以删除所有，其他用户只能删除自己上传的
const canDelete = computed(() => {
  if (userStore.isAdmin || userStore.isDeptManager) return true
  return item.value.uploaded_by === userStore.userId
})

const fetchItem = async () => {
  const res = await getKnowledge(itemId.value)
  item.value = res.data
}

const fetchVersions = async () => {
  const res = await getVersions(itemId.value)
  versions.value = res.data || []
}

const handleDownload = async () => {
  try {
    const response = await downloadKnowledge(itemId.value)
    
    // 从响应头获取文件名，如果没有则使用标题
    let filename = item.value.title || '资料下载'
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

const handlePreview = async () => {
  const fileName = item.value.file_path ? item.value.file_path.toLowerCase() : ''
  const ext = fileName.substring(fileName.lastIndexOf('.'))
  
  // 判断文件类型
  if (['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp'].includes(ext)) {
    previewType.value = 'image'
    try {
      const response = await downloadKnowledge(itemId.value)
      const blob = new Blob([response.data])
      previewUrl.value = window.URL.createObjectURL(blob)
      previewVisible.value = true
    } catch (error) {
      ElMessage.error('预览失败')
    }
  } else if (ext === '.pdf') {
    previewType.value = 'pdf'
    try {
      const response = await downloadKnowledge(itemId.value)
      const blob = new Blob([response.data], { type: 'application/pdf' })
      previewUrl.value = window.URL.createObjectURL(blob)
      previewVisible.value = true
    } catch (error) {
      ElMessage.error('预览失败')
    }
  } else if (['.txt', '.log', '.md', '.json', '.xml', '.csv'].includes(ext)) {
    previewType.value = 'text'
    try {
      const response = await downloadKnowledge(itemId.value)
      const blob = new Blob([response.data])
      previewContent.value = await blob.text()
      previewVisible.value = true
    } catch (error) {
      ElMessage.error('预览失败')
    }
  } else {
    previewType.value = 'unsupported'
    previewVisible.value = true
  }
}

const showVersionDialog = () => {
  versionForm.change_note = ''
  versionFile.value = null
  versionDialogVisible.value = true
}

const handleVersionFileChange = (file) => {
  versionFile.value = file.raw
}

const handleUploadVersion = async () => {
  if (!versionFile.value) {
    ElMessage.warning('请选择文件')
    return
  }
  
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', versionFile.value)
    formData.append('change_note', versionForm.change_note)
    
    await uploadNewVersion(itemId.value, formData)
    ElMessage.success('新版本上传成功')
    versionDialogVisible.value = false
    fetchItem()
    fetchVersions()
  } catch (error) {
    console.error('上传失败:', error)
  } finally {
    uploading.value = false
  }
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除资料「${item.value.title}」吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteKnowledge(itemId.value)
    ElMessage.success('删除成功')
    router.push('/knowledge')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([fetchItem(), fetchVersions()])
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.knowledge-detail { max-width: 900px; }
.info-card, .version-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.version-item { display: flex; align-items: center; gap: 10px; }
.version-tag { background: #409EFF; color: #fff; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.version-note { color: #666; }
</style>
