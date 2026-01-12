<template>
  <div v-loading="loading" class="task-detail">
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span class="title">任务信息</span>
          <div class="card-header-actions">
            <el-button type="primary" link @click="$router.back()">
              <el-icon><ArrowLeft /></el-icon>
              返回
            </el-button>
            <el-tag :type="statusTypes[task.status]">{{ statusLabels[task.status] }}</el-tag>
          </div>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="任务名称" :span="2">{{ task.task_name }}</el-descriptions-item>
        <el-descriptions-item label="所属项目">
          <el-link type="primary" @click="$router.push(`/projects/${task.project_id}`)">{{ task.project?.name }}</el-link>
        </el-descriptions-item>
        <el-descriptions-item label="所属阶段">{{ phaseLabels[task.phase?.phase_name] || '-' }}</el-descriptions-item>
        <el-descriptions-item label="负责人">{{ task.assignee?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(task.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="截止时间" :span="2">
          <span :class="{ 'overdue': isOverdue }">{{ formatDateTime(task.deadline) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="优先级">
          <el-tag :type="priorityTypes[task.priority]" size="small">{{ priorityLabels[task.priority] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatDateTime(task.completed_at) }}</el-descriptions-item>
        <el-descriptions-item label="任务描述" :span="2">{{ task.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="交付要求" :span="2">{{ task.deliverables || '-' }}</el-descriptions-item>
      </el-descriptions>

      <!-- 审核信息 -->
      <div v-if="task.review_status" class="review-info">
        <el-divider content-position="left">审核信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="审核状态">
            <el-tag :type="task.review_status === 'approved' ? 'success' : 'danger'">
              {{ task.review_status === 'approved' ? '已通过' : '被驳回' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="审核时间">{{ formatDateTime(task.reviewed_at) }}</el-descriptions-item>
          <el-descriptions-item label="审核意见" :span="2">{{ task.review_comment || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>

    <!-- 操作按钮 -->
    <div v-if="canOperate" class="action-bar">
      <el-button v-if="task.status === 'not_started'" type="primary" @click="handleStatusChange('in_progress')">开始任务</el-button>
      <el-button v-if="task.status === 'in_progress'" type="success" @click="handleStatusChange('completed')">完成任务</el-button>
      <el-button v-if="task.status === 'rejected'" type="warning" @click="handleStatusChange('in_progress')">重新开始</el-button>
      <el-button v-if="task.status === 'completed' && isProjectManager" type="warning" @click="handleStatusChange('in_progress')">重新开始</el-button>
    </div>

    <!-- 交付件 -->
    <el-card class="docs-card">
      <template #header>
        <div class="card-header">
          <span>交付件</span>
          <el-upload
            v-if="canUpload"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :data="{ task_id: taskId, project_id: task.project_id }"
            :show-file-list="false"
            :on-success="handleUploadSuccess"
          >
            <el-button type="primary" size="small"><el-icon><Upload /></el-icon> 上传文件</el-button>
          </el-upload>
        </div>
      </template>
      <el-table :data="documents" stripe>
        <el-table-column prop="doc_name" label="文件名" min-width="200" />
        <el-table-column prop="file_size" label="大小" width="100">
          <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column prop="uploader.name" label="上传人" width="100" />
        <el-table-column prop="created_at" label="上传时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="success" link @click="handlePreview(row)">预览</el-button>
            <el-button type="primary" link @click="handleDownload(row)">下载</el-button>
            <el-popconfirm v-if="canDeleteDoc" title="确定删除该交付件吗？" width="210" @confirm="handleDeleteDoc(row)">
              <template #reference>
                <el-button type="danger" link>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="documents.length === 0" description="暂无交付件" />
    </el-card>

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
            <el-button type="primary" @click="handleDownload(currentPreviewFile)">下载文件</el-button>
          </template>
        </el-result>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getTask, updateTaskStatus } from '@/api/task'
import { getDocuments, getDownloadUrl, deleteDocument } from '@/api/document'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const route = useRoute()
const userStore = useUserStore()
const taskId = computed(() => route.params.id)

const loading = ref(false)
const task = ref({})
const documents = ref([])
const previewVisible = ref(false)
const previewUrl = ref('')
const previewType = ref('')
const previewContent = ref('')
const currentPreviewFile = ref(null)

const uploadUrl = '/project_track/api/documents/upload'
const uploadHeaders = computed(() => ({ Authorization: `Bearer ${localStorage.getItem('token')}` }))

const statusLabels = { not_started: '未开始', in_progress: '进行中', completed: '已完成', rejected: '被驳回' }
const statusTypes = { not_started: 'info', in_progress: 'warning', completed: 'success', rejected: 'danger' }
const priorityLabels = { 1: '高', 2: '中', 3: '低' }
const priorityTypes = { 1: 'danger', 2: 'warning', 3: 'info' }
const phaseLabels = { initiation: '立项', bidding: '招标', contract: '合同签订', execution: '项目实施', testing: '测试', acceptance: '验收', closing: '结项' }

const formatDate = (dateStr) => dateStr ? dateStr.split('T')[0] : '-'
const formatDateTime = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'
const formatFileSize = (bytes) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) { bytes /= 1024; i++ }
  return `${bytes.toFixed(1)} ${units[i]}`
}

const isOverdue = computed(() => {
  if (!task.value.deadline || task.value.status === 'completed') return false
  return new Date(task.value.deadline) < new Date()
})

// 是否为项目经理
const isProjectManager = computed(() => {
  return task.value.project?.manager_id === userStore.user.id
})

// 是否为任务负责人
const isTaskAssignee = computed(() => {
  return task.value.assignee_id === userStore.user.id
})

// 能否操作任务状态
const canOperate = computed(() => {
  // 任务未完成：任务负责人或项目经理都可以操作
  if (task.value.status !== 'completed') {
    return isTaskAssignee.value || isProjectManager.value
  }
  // 任务已完成：只有项目经理可以重新开始
  return isProjectManager.value
})

// 能否上传文件
const canUpload = computed(() => {
  // 任务已完成：只有项目经理可以上传
  if (task.value.status === 'completed') {
    return isProjectManager.value
  }
  // 任务未完成：任务负责人可以上传
  return isTaskAssignee.value
})

// 能否删除交付件
const canDeleteDoc = computed(() => {
  // 任务已完成：只有项目经理可以删除
  if (task.value.status === 'completed') {
    return isProjectManager.value
  }
  // 任务未完成：任务负责人可以删除
  return isTaskAssignee.value
})

const fetchTask = async () => {
  const res = await getTask(taskId.value)
  task.value = res.data
}

const fetchDocuments = async () => {
  const res = await getDocuments({ task_id: taskId.value, page: 1, page_size: 100 })
  documents.value = res.data?.list || []
}

const handleStatusChange = async (status) => {
  try {
    await updateTaskStatus(taskId.value, status)
    ElMessage.success('状态更新成功')
    fetchTask()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

const handleUploadSuccess = () => { ElMessage.success('上传成功'); fetchDocuments() }
const handleDownload = (row) => { window.open(getDownloadUrl(row.id), '_blank') }

const handleDeleteDoc = async (row) => {
  try {
    await deleteDocument(row.id)
    ElMessage.success('删除成功')
    fetchDocuments()
  } catch (error) {
    console.error('删除失败:', error)
    ElMessage.error('删除失败')
  }
}

const handlePreview = async (row) => {
  currentPreviewFile.value = row
  const fileName = row.doc_name.toLowerCase()
  const ext = fileName.substring(fileName.lastIndexOf('.'))
  
  // 判断文件类型
  if (['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp'].includes(ext)) {
    previewType.value = 'image'
    previewUrl.value = getDownloadUrl(row.id)
    previewVisible.value = true
  } else if (ext === '.pdf') {
    previewType.value = 'pdf'
    previewUrl.value = getDownloadUrl(row.id)
    previewVisible.value = true
  } else if (['.txt', '.log', '.md', '.json', '.xml', '.csv'].includes(ext)) {
    previewType.value = 'text'
    try {
      // 获取文本内容
      const response = await fetch(getDownloadUrl(row.id))
      previewContent.value = await response.text()
      previewVisible.value = true
    } catch (error) {
      ElMessage.error('预览失败')
    }
  } else {
    previewType.value = 'unsupported'
    previewVisible.value = true
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([fetchTask(), fetchDocuments()])
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.task-detail { max-width: 1000px; }
.info-card, .docs-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.card-header .title { font-size: 16px; font-weight: 500; }
.card-header-actions { display: flex; align-items: center; gap: 15px; }
.action-bar { margin-bottom: 20px; }
.overdue { color: #F56C6C; font-weight: bold; }
.review-info { margin-top: 20px; }
.preview-container { text-align: center; }
</style>
