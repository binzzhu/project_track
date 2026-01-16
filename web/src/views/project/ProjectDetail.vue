<template>
  <div v-loading="loading" class="project-detail">
    <!-- 基本信息 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span class="title">项目信息</span>
          <div class="card-header-actions">
            <el-button type="primary" link @click="handleBack">
              <el-icon><ArrowLeft /></el-icon>
              返回
            </el-button>
            <el-button v-if="canEditProject" type="primary" link @click="showEditDialog">
              <el-icon><Edit /></el-icon> 编辑
            </el-button>
          </div>
        </div>
      </template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="项目编号">{{ project.project_no }}</el-descriptions-item>
        <el-descriptions-item label="项目名称">{{ project.name }}</el-descriptions-item>
        <el-descriptions-item label="项目类型">{{ project.project_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="项目负责人">{{ project.manager?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="当前阶段">
          <el-tag>{{ phaseLabels[project.current_phase] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="立项日期">{{ formatDate(project.initiation_date) }}</el-descriptions-item>
        <el-descriptions-item label="结项日期">{{ formatDate(project.closing_date) }}</el-descriptions-item>
        <el-descriptions-item label="项目状态">
          <el-tag :type="statusTypes[project.status]">{{ statusLabels[project.status] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="合同编号">
          <template #default>
            <div v-if="project.contract_no">
              <div v-for="(no, index) in splitContractNos(project.contract_no)" :key="index">{{ no }}</div>
            </div>
            <span v-else>-</span>
          </template>
        </el-descriptions-item>
        <el-descriptions-item label="预算编码">{{ project.budget_code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创新项目编码">{{ project.innovation_code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="人工费用">{{ project.labor_cost ? `¥${project.labor_cost.toLocaleString()}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="直接投入费用">{{ project.direct_cost ? `¥${project.direct_cost.toLocaleString()}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="委托研发费用">{{ project.outsourcing_cost ? `¥${project.outsourcing_cost.toLocaleString()}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="其他费用">{{ project.other_cost ? `¥${project.other_cost.toLocaleString()}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="总费用" :span="2">
          <strong>{{ totalCost ? `¥${totalCost.toLocaleString()}` : '-' }}</strong>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 项目成员 -->
    <el-card class="members-card">
      <template #header>
        <div class="card-header">
          <span>项目成员</span>
          <el-button v-if="canEditProject" type="primary" size="small" @click="showMemberDialog">
            <el-icon><Plus /></el-icon> 添加成员
          </el-button>
        </div>
      </template>
      <el-table :data="members" stripe size="small">
        <el-table-column prop="user.name" label="姓名" width="100" />
        <el-table-column prop="user.username" label="用户名" width="120" />
        <el-table-column prop="role_type" label="项目角色" width="100">
          <template #default="{ row }">{{ roleLabels[row.role_type] || row.role_type }}</template>
        </el-table-column>
        <el-table-column prop="join_date" label="加入时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.join_date) }}</template>
        </el-table-column>
        <el-table-column v-if="canEditProject" label="操作" width="80">
          <template #default="{ row }">
            <el-popconfirm title="确定移除该成员吗？" width="200" @confirm="handleRemoveMember(row.id)">
              <template #reference>
                <el-button type="danger" link size="small">移除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 阶段管理 -->
    <el-card class="phase-card">
      <template #header>
        <div class="card-header">
          <span>项目阶段</span>
          <el-button v-if="canEditProject" type="primary" size="small" @click="showPhaseDialog">
            <el-icon><Plus /></el-icon> 添加研发阶段
          </el-button>
        </div>
      </template>
      <div class="phase-tips">
        <el-text type="info" size="small">提示：立项、招标、合同签订、验收、结项为固定阶段，中间的研发阶段可由项目总工自定义。每个阶段包含对应的任务和文档。</el-text>
      </div>
      
      <!-- 阶段卡片列表 -->
      <el-collapse v-model="activePhases" class="phase-collapse">
        <el-collapse-item v-for="phase in sortedPhases" :key="phase.id" :name="phase.id">
          <template #title>
            <div class="phase-header">
              <span class="phase-order">{{ phase.phase_order }}</span>
              <span class="phase-name">{{ phaseLabels[phase.phase_name] || phase.phase_name }}</span>
              <el-tag v-if="phase.is_fixed" size="small" type="info">固定</el-tag>
              <el-tag :type="statusTypes[phase.status]" size="small">{{ statusLabels[phase.status] }}</el-tag>
              <span class="phase-stats">
                <el-text size="small" type="info">任务: {{ getPhaseTaskCount(phase.id) }} | 文档: {{ getPhaseDocCount(phase.id) }}</el-text>
              </span>
            </div>
          </template>
          
          <!-- 阶段操作栏 -->
          <div class="phase-actions">
            <el-button v-if="canEditProject && canUpdatePhase(phase)" type="primary" size="small" @click.stop="handlePhaseAction(phase)">
              {{ phase.status === 'in_progress' ? '完成阶段' : '开始阶段' }}
            </el-button>
            <el-button v-if="canEditProject" type="success" size="small" @click.stop="showTaskDialog(phase.id)">
              <el-icon><Plus /></el-icon> 添加任务
            </el-button>
            <el-upload
              v-if="canUploadDocument"
              :action="uploadUrl"
              :headers="uploadHeaders"
              :data="getUploadData(phase.id)"
              :show-file-list="false"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
              class="upload-inline"
            >
              <el-button type="warning" size="small"><el-icon><Upload /></el-icon> 上传文档</el-button>
            </el-upload>
            <el-popconfirm v-if="canEditProject && !phase.is_fixed" title="确定删除该阶段吗？" width="200" @confirm="handleDeletePhase(phase.id)">
              <template #reference>
                <el-button type="danger" size="small" @click.stop>删除阶段</el-button>
              </template>
            </el-popconfirm>
          </div>
          
          <!-- 阶段内容：任务和文档 -->
          <el-tabs type="border-card" class="phase-content-tabs">
            <el-tab-pane label="任务列表">
              <el-table :data="getPhaseTasks(phase.id)" stripe size="small" empty-text="暂无任务">
                <el-table-column prop="task_name" label="任务名称" min-width="180" header-align="center" align="center">
                  <template #default="{ row }">
                    <el-link type="primary" @click="showTaskDetail(row)">{{ row.task_name }}</el-link>
                  </template>
                </el-table-column>
                <el-table-column prop="assignee" label="负责人" width="100" header-align="center" align="center">
                  <template #default="{ row }">{{ row.assignee?.name || '-' }}</template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="100" header-align="center" align="center">
                  <template #default="{ row }">
                    <el-tag :type="taskStatusTypes[row.status]" size="small">{{ taskStatusLabels[row.status] }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="created_at" label="创建时间" width="160" header-align="center" align="center">
                  <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
                </el-table-column>
                <el-table-column prop="deadline" label="截止时间" width="160" header-align="center" align="center">
                  <template #default="{ row }">{{ formatDateTime(row.deadline) }}</template>
                </el-table-column>
                <el-table-column v-if="canEditProject" label="操作" width="180" header-align="center" align="center">
                  <template #default="{ row }">
                    <div style="display: flex; justify-content: center; align-items: center; gap: 4px;">
                      <el-dropdown v-if="canUpdateTaskStatus(row)" @command="(cmd) => handleTaskStatusUpdate(row, cmd)">
                        <el-button type="primary" link size="small">更新状态</el-button>
                        <template #dropdown>
                          <el-dropdown-menu>
                            <el-dropdown-item v-if="row.status === 'not_started'" command="in_progress">开始任务</el-dropdown-item>
                            <el-dropdown-item v-if="row.status === 'in_progress'" command="completed">完成任务</el-dropdown-item>
                            <el-dropdown-item v-if="row.status === 'completed'" command="in_progress">重新开始</el-dropdown-item>
                          </el-dropdown-menu>
                        </template>
                      </el-dropdown>
                      <el-popconfirm title="确定删除该任务吗？" width="200" @confirm="handleDeleteTask(row.id)">
                        <template #reference>
                          <el-button type="danger" link size="small">删除</el-button>
                        </template>
                      </el-popconfirm>
                    </div>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            <el-tab-pane label="文档资料">
              <el-table :data="getPhaseDocs(phase.id)" stripe size="small" empty-text="暂无文档">
                <el-table-column prop="doc_name" label="文档名称" min-width="180" header-align="center" align="center" />
                <el-table-column prop="doc_type" label="类型" width="80" header-align="center" align="center" />
                <el-table-column prop="file_size" label="大小" width="80" header-align="center" align="center">
                  <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
                </el-table-column>
                <el-table-column prop="uploader.name" label="上传人" width="80" header-align="center" align="center" />
                <el-table-column prop="created_at" label="上传时间" width="140" header-align="center" align="center">
                  <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
                </el-table-column>
                <el-table-column label="操作" width="180" header-align="center" align="center">
                  <template #default="{ row }">
                    <el-button type="success" link size="small" @click="handlePreview(row)">预览</el-button>
                    <el-button type="primary" link size="small" @click="handleDownload(row)">下载</el-button>
                    <el-popconfirm v-if="canDeleteDocument(row)" title="确定删除该文档吗？" width="200" @confirm="handleDeleteDocument(row.id)">
                      <template #reference>
                        <el-button type="danger" link size="small">删除</el-button>
                      </template>
                    </el-popconfirm>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <!-- 添加阶段弹窗 -->
    <el-dialog v-model="phaseDialogVisible" title="添加研发阶段" width="400px">
      <el-form :model="phaseForm" label-width="80px">
        <el-form-item label="阶段名称">
          <el-input v-model="phaseForm.phase_name" placeholder="请输入研发阶段名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="phaseDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddPhase">确定</el-button>
      </template>
    </el-dialog>

    <!-- 添加成员弹窗 -->
    <el-dialog v-model="memberDialogVisible" title="添加成员" width="400px">
      <el-form :model="memberForm" label-width="80px">
        <el-form-item label="选择用户">
          <el-select v-model="memberForm.user_id" placeholder="请选择" filterable>
            <el-option v-for="user in availableUsers" :key="user.id" :label="`${user.name} (${user.username})`" :value="user.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="项目角色">
          <el-select v-model="memberForm.role_type" placeholder="请选择">
            <el-option label="子负责人" value="sub_manager" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="memberDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddMember">确定</el-button>
      </template>
    </el-dialog>

    <!-- 创建任务弹窗 -->
    <el-dialog v-model="taskDialogVisible" title="创建任务" width="500px">
      <el-form ref="taskFormRef" :model="taskForm" :rules="taskRules" label-width="100px">
        <el-form-item label="任务名称" prop="task_name">
          <el-input v-model="taskForm.task_name" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-select v-model="taskForm.assignee_id" placeholder="请选择" filterable style="width: 100%;">
            <el-option v-for="user in allUsers" :key="user.id" :label="user.name" :value="user.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="截止时间">
          <el-date-picker v-model="taskForm.deadline" type="datetime" value-format="YYYY-MM-DD HH:mm" placeholder="请选择截止时间" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="taskForm.priority" placeholder="请选择" style="width: 100%;">
            <el-option :value="1" label="高" />
            <el-option :value="2" label="中" />
            <el-option :value="3" label="低" />
          </el-select>
        </el-form-item>
        <el-form-item label="任务描述">
          <el-input v-model="taskForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="交付要求">
          <el-input v-model="taskForm.deliverables" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="taskDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateTask">确定</el-button>
      </template>
    </el-dialog>

    <!-- 任务详情抽屉 -->
    <el-drawer v-model="taskDrawerVisible" title="任务详情" size="450px" direction="rtl">
      <div v-if="currentTask" class="task-drawer-content">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="任务名称">{{ currentTask.task_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="taskStatusTypes[currentTask.status]" size="small">{{ taskStatusLabels[currentTask.status] }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="负责人">{{ currentTask.assignee?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="截止时间">{{ formatDateTime(currentTask.deadline) }}</el-descriptions-item>
          <el-descriptions-item label="优先级">
            <el-tag :type="priorityTypes[currentTask.priority]" size="small">{{ priorityLabels[currentTask.priority] }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="任务描述">{{ currentTask.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="交付要求">{{ currentTask.deliverables || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDateTime(currentTask.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <!-- 任务操作按钮 -->
        <div v-if="canOperateTask" class="task-actions">
          <el-button v-if="currentTask.status === 'not_started'" type="primary" size="small" @click="handleTaskStatusChange('in_progress')">开始任务</el-button>
          <el-button v-if="currentTask.status === 'in_progress'" type="success" size="small" @click="handleTaskStatusChange('completed')">完成任务</el-button>
          <el-button v-if="currentTask.status === 'rejected'" type="warning" size="small" @click="handleTaskStatusChange('in_progress')">重新开始</el-button>
        </div>

        <!-- 任务交付件 -->
        <div class="task-docs">
          <div class="docs-header">
            <span class="docs-title">交付件</span>
            <el-upload
              v-if="canUploadTaskDoc"
              :action="uploadUrl"
              :headers="uploadHeaders"
              :data="{ task_id: currentTask.id, project_id: route.params.id, phase_id: currentTask.phase_id }"
              :show-file-list="false"
              :on-success="handleTaskDocUploadSuccess"
              :on-error="handleUploadError"
            >
              <el-button type="primary" size="small"><el-icon><Upload /></el-icon> 上传</el-button>
            </el-upload>
          </div>
          <el-table :data="taskDocuments" stripe size="small" empty-text="暂无交付件">
            <el-table-column prop="doc_name" label="文件名" min-width="150" />
            <el-table-column prop="file_size" label="大小" width="80">
              <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button type="success" link size="small" @click="handlePreview(row)">预览</el-button>
                <el-button type="primary" link size="small" @click="handleDownload(row)">下载</el-button>
                <el-popconfirm v-if="canDeleteDocument(row)" title="确定删除该文件吗？" width="200" @confirm="handleDeleteDocument(row.id)">
                  <template #reference>
                    <el-button type="danger" link size="small">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-drawer>

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
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProject, getProjectPhases, updateProjectPhase, addProjectPhase, deleteProjectPhase, getProjectMembers, addProjectMember, removeProjectMember } from '@/api/project'
import { getTasks, createTask, getTask, updateTaskStatus, deleteTask } from '@/api/task'
import { getDocuments, getDownloadUrl, deleteDocument } from '@/api/document'
import { getUsers } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const projectId = computed(() => route.params.id)

const loading = ref(false)
const project = ref({})
const phases = ref([])
const tasks = ref([])
const members = ref([])
const documents = ref([])
const allUsers = ref([])

const memberDialogVisible = ref(false)
const taskDialogVisible = ref(false)
const phaseDialogVisible = ref(false)
const taskFormRef = ref(null)
const activePhases = ref([])  // 当前展开的阶段
const taskDrawerVisible = ref(false)  // 任务详情抽屉
const currentTask = ref(null)  // 当前查看的任务
const taskDocuments = ref([])  // 任务交付件
const previewVisible = ref(false)
const previewUrl = ref('')
const previewType = ref('')
const previewContent = ref('')
const currentPreviewFile = ref(null)

const memberForm = reactive({ user_id: null, role_type: 'sub_manager' })
const taskForm = reactive({ task_name: '', phase_id: null, assignee_id: null, deadline: '', priority: 2, description: '', deliverables: '' })
const phaseForm = reactive({ phase_name: '' })

const taskRules = { task_name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }] }

const uploadUrl = '/project_track/api/documents/upload'
const uploadHeaders = computed(() => ({ Authorization: `Bearer ${localStorage.getItem('token')}` }))

// 权限判断：是否可以编辑项目（管理员或项目负责人）
const canEditProject = computed(() => {
  if (userStore.isAdmin) return true
  return project.value.manager_id === userStore.userId
})

// 权限判断：是否可以上传文档（管理员、项目负责人或子负责人）
const canUploadDocument = computed(() => {
  if (userStore.isAdmin) return true
  if (project.value.manager_id === userStore.userId) return true
  // 检查是否是子负责人（在成员列表中的用户即为子负责人）
  return members.value.some(m => m.user_id === userStore.userId)
})

const phaseLabels = { initiation: '立项', bidding: '招标', contract: '合同签订', acceptance: '验收', closing: '结项' }
const statusLabels = { not_started: '未开始', in_progress: '进行中', completed: '已完成' }
const statusTypes = { not_started: 'info', in_progress: 'primary', completed: 'success' }
const taskStatusLabels = { not_started: '未开始', in_progress: '进行中', completed: '已完成', rejected: '被驳回' }
const taskStatusTypes = { not_started: 'info', in_progress: 'warning', completed: 'success', rejected: 'danger' }
const roleLabels = { manager: '项目负责人', sub_manager: '子负责人' }
const priorityLabels = { 1: '高', 2: '中', 3: '低' }
const priorityTypes = { 1: 'danger', 2: 'warning', 3: 'info' }

// 按顺序排列阶段
const sortedPhases = computed(() => {
  return [...phases.value].sort((a, b) => a.phase_order - b.phase_order)
})

const currentPhaseIndex = computed(() => {
  const idx = phases.value.findIndex(p => p.phase_name === project.value.current_phase)
  return idx >= 0 ? idx : 0
})

const availableUsers = computed(() => {
  const memberIds = members.value.map(m => m.user_id)
  return allUsers.value.filter(u => !memberIds.includes(u.id))
})

const handleBack = () => {
  router.back()
}

// 计算总费用
const totalCost = computed(() => {
  const { labor_cost = 0, direct_cost = 0, outsourcing_cost = 0, other_cost = 0 } = project.value
  return labor_cost + direct_cost + outsourcing_cost + other_cost
})

const formatDate = (dateStr) => dateStr ? dateStr.split('T')[0] : '-'
const formatDateTime = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'
const formatFileSize = (bytes) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) { bytes /= 1024; i++ }
  return `${bytes.toFixed(1)} ${units[i]}`
}

const splitContractNos = (val) => {
  if (!val) return []
  return val.split(',').map(item => item.trim()).filter(item => item)
}

const getPhaseStatus = (phase) => {
  if (phase.status === 'completed') return 'success'
  if (phase.status === 'in_progress') return 'process'
  return 'wait'
}

const canUpdatePhase = (phase) => {
  const idx = phases.value.findIndex(p => p.id === phase.id)
  if (phase.status === 'completed') return false
  if (phase.status === 'in_progress') return true
  // 只有前一阶段完成才能开始当前阶段
  if (idx === 0) return phase.status === 'not_started'
  return phases.value[idx - 1]?.status === 'completed'
}

const handlePhaseAction = async (phase) => {
  const newStatus = phase.status === 'in_progress' ? 'completed' : 'in_progress'
  try {
    await updateProjectPhase(projectId.value, phase.id, { status: newStatus })
    ElMessage.success('更新成功')
    fetchData()
  } catch (error) {
    console.error('更新阶段失败:', error)
  }
}

const showMemberDialog = () => { memberForm.user_id = null; memberForm.role_type = 'sub_manager'; memberDialogVisible.value = true }
const showTaskDialog = (phaseId = null) => { Object.assign(taskForm, { task_name: '', phase_id: phaseId, assignee_id: null, deadline: '', priority: 2, description: '', deliverables: '' }); taskDialogVisible.value = true }
const showPhaseDialog = () => { phaseForm.phase_name = ''; phaseDialogVisible.value = true }

// 获取阶段下的任务列表（类型转换确保比较正确）
const getPhaseTasks = (phaseId) => tasks.value.filter(t => Number(t.phase_id) === Number(phaseId))
const getPhaseTaskCount = (phaseId) => tasks.value.filter(t => Number(t.phase_id) === Number(phaseId)).length

// 获取阶段下的文档列表（类型转换确保比较正确）
const getPhaseDocs = (phaseId) => documents.value.filter(d => Number(d.phase_id) === Number(phaseId))
const getPhaseDocCount = (phaseId) => documents.value.filter(d => Number(d.phase_id) === Number(phaseId)).length
const showEditDialog = () => { /* 编辑逻辑 */ }

const handleAddPhase = async () => {
  if (!phaseForm.phase_name) { ElMessage.warning('请输入阶段名称'); return }
  try {
    await addProjectPhase(projectId.value, phaseForm)
    ElMessage.success('添加成功')
    phaseDialogVisible.value = false
    fetchProject()
  } catch (error) {
    console.error('添加阶段失败:', error)
  }
}

const handleDeletePhase = async (phaseId) => {
  try {
    await deleteProjectPhase(projectId.value, phaseId)
    ElMessage.success('删除成功')
    fetchProject()
  } catch (error) {
    console.error('删除阶段失败:', error)
  }
}

const handleAddMember = async () => {
  if (!memberForm.user_id) { ElMessage.warning('请选择用户'); return }
  try {
    await addProjectMember(projectId.value, memberForm)
    ElMessage.success('添加成功')
    memberDialogVisible.value = false
    fetchMembers()
  } catch (error) {
    console.error('添加成员失败:', error)
  }
}

const handleRemoveMember = async (memberId) => {
  try {
    await removeProjectMember(projectId.value, memberId)
    ElMessage.success('移除成功')
    fetchMembers()
  } catch (error) {
    console.error('移除成员失败:', error)
  }
}

const handleCreateTask = async () => {
  if (!taskFormRef.value) return
  await taskFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      await createTask({ ...taskForm, project_id: parseInt(projectId.value) })
      ElMessage.success('创建成功')
      taskDialogVisible.value = false
      fetchTasks()
    } catch (error) {
      console.error('创建任务失败:', error)
    }
  })
}

const handleDeleteTask = async (taskId) => {
  try {
    await deleteTask(taskId)
    ElMessage.success('删除成功')
    fetchTasks()
  } catch (error) {
    console.error('删除任务失败:', error)
  }
}

const handleUploadSuccess = (response) => {
  if (response.code === 200) {
    ElMessage.success('上传成功')
    fetchDocuments()
  } else {
    ElMessage.error(response.message || '上传失败')
  }
}
const handleUploadError = (error) => {
  console.error('上传失败:', error)
  ElMessage.error('上传失败，请重试')
}
const getUploadData = (phaseId) => ({ project_id: route.params.id, phase_id: phaseId })
const handleDownload = (row) => { window.open(getDownloadUrl(row.id), '_blank') }

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

const handleDeleteDocument = async (docId) => {
  try {
    await deleteDocument(docId)
    ElMessage.success('删除成功')
    // 刷新文档列表
    fetchDocuments()
    // 如果当前在任务抽屉中，也需要刷新任务交付件
    if (currentTask.value) {
      const res = await getDocuments({ task_id: currentTask.value.id, page: 1, page_size: 100 })
      taskDocuments.value = res.data?.list || []
    }
  } catch (error) {
    console.error('删除文档失败:', error)
  }
}

// 任务详情抽屉相关
const showTaskDetail = async (task) => {
  currentTask.value = task
  taskDrawerVisible.value = true
  // 获取任务交付件
  const res = await getDocuments({ task_id: task.id, page: 1, page_size: 100 })
  taskDocuments.value = res.data?.list || []
}

// 是否为项目经理
const isProjectManager = computed(() => {
  return project.value.manager_id === userStore.userId
})

// 能否操作任务状态
const canOperateTask = computed(() => {
  if (!currentTask.value) return false
  const isTaskAssignee = currentTask.value.assignee_id === userStore.userId
  // 任务未完成：任务负责人或项目经理都可以操作
  if (currentTask.value.status !== 'completed') {
    return isTaskAssignee || isProjectManager.value
  }
  // 任务已完成：只有项目经理可以重新开始
  return isProjectManager.value
})

// 能否上传文件
const canUploadTaskDoc = computed(() => {
  if (!currentTask.value) return false
  const isTaskAssignee = currentTask.value.assignee_id === userStore.userId
  // 任务已完成：只有项目经理可以上传
  if (currentTask.value.status === 'completed') {
    return isProjectManager.value
  }
  // 任务未完成：任务负责人可以上传
  return isTaskAssignee
})

// 判断是否可以删除文档
const canDeleteDocument = (doc) => {
  if (!currentTask.value) return false
  const isTaskAssignee = currentTask.value.assignee_id === userStore.userId
  // 任务已完成：只有项目经理可以删除
  if (currentTask.value.status === 'completed') {
    return isProjectManager.value
  }
  // 任务未完成：任务负责人可以删除
  return isTaskAssignee
}

const handleTaskStatusChange = async (status) => {
  try {
    await updateTaskStatus(currentTask.value.id, status)
    ElMessage.success('状态更新成功')
    currentTask.value.status = status
    fetchTasks()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

// 判断是否可以更新任务状态
const canUpdateTaskStatus = (task) => {
  const isTaskAssignee = task.assignee_id === userStore.userId
  const isProjectManager = project.value.manager_id === userStore.userId
  
  // 任务未完成：任务负责人或项目经理都可以更改状态
  if (task.status !== 'completed') {
    return isTaskAssignee || isProjectManager
  }
  
  // 任务已完成：只有项目经理可以重新开始
  return isProjectManager
}

// 处理任务状态更新（项目阶段任务列表中）
const handleTaskStatusUpdate = async (task, status) => {
  try {
    await updateTaskStatus(task.id, status)
    ElMessage.success('状态更新成功')
    fetchTasks()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

const handleTaskDocUploadSuccess = (response) => {
  if (response.code === 200) {
    ElMessage.success('上传成功')
    // 刷新任务交付件列表
    getDocuments({ task_id: currentTask.value.id, page: 1, page_size: 100 }).then(res => {
      taskDocuments.value = res.data?.list || []
    })
    fetchDocuments()
  } else {
    ElMessage.error(response.message || '上传失败')
  }
}

const fetchProject = async () => {
  const res = await getProject(projectId.value)
  project.value = res.data
  phases.value = res.data.phases || []
}

const fetchTasks = async () => {
  const res = await getTasks({ project_id: projectId.value, page: 1, page_size: 100 })
  tasks.value = res.data?.list || []
}

const fetchMembers = async () => {
  const res = await getProjectMembers(projectId.value)
  members.value = res.data || []
}

const fetchDocuments = async () => {
  const res = await getDocuments({ project_id: projectId.value, page: 1, page_size: 100 })
  documents.value = res.data?.list || []
}

const fetchUsers = async () => {
  const res = await getUsers({ page: 1, page_size: 100 })
  allUsers.value = res.data?.list || []
}

const fetchData = async () => {
  loading.value = true
  try {
    await Promise.all([fetchProject(), fetchTasks(), fetchMembers(), fetchDocuments(), fetchUsers()])
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchData())
</script>

<style scoped>
.project-detail { max-width: 1200px; }
.info-card, .members-card, .phase-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.title { font-weight: 500; }
.card-header-actions { display: flex; gap: 8px; }
.phase-tips { margin-bottom: 16px; padding: 8px 12px; background: #f5f7fa; border-radius: 4px; }

/* 阶段折叠卡片样式 */
.phase-collapse { margin-top: 12px; }
.phase-header { display: flex; align-items: center; gap: 12px; flex: 1; }
.phase-order { width: 24px; height: 24px; background: #409eff; color: #fff; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 12px; }
.phase-name { font-weight: 500; min-width: 100px; }
.phase-stats { margin-left: auto; }
.phase-actions { display: flex; gap: 8px; margin-bottom: 16px; padding: 12px; background: #fafafa; border-radius: 4px; }
.upload-inline { display: inline-block; }
.phase-content-tabs { margin-top: 8px; }

/* 任务抽屉样式 */
.task-drawer-content { padding: 0 8px; }
.task-actions { margin: 16px 0; display: flex; gap: 8px; }
.task-docs { margin-top: 20px; }
.docs-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.docs-title { font-weight: 500; font-size: 14px; }
.preview-container { text-align: center; }
</style>
