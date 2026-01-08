<template>
  <div class="task-list">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="任务名称" clearable @keyup.enter="handleSearch" style="width: 220px;" />
        </el-form-item>
        <el-form-item label="任务负责人">
          <el-select v-model="searchForm.assignee_id" placeholder="请选择负责人" clearable filterable style="width: 150px;">
            <el-option v-for="user in users" :key="user.id" :label="user.name" :value="user.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px;">
            <el-option label="未开始" value="not_started" />
            <el-option label="进行中" value="in_progress" />
            <el-option label="已完成" value="completed" />
            <el-option label="被驳回" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <el-table v-loading="loading" :data="tasks" stripe>
        <el-table-column prop="task_name" label="任务名称" min-width="200" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="$router.push(`/tasks/${row.id}`)">{{ row.task_name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="project" label="所属项目" min-width="200" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="$router.push(`/projects/${row.project_id}`)">{{ row.project?.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="project.manager" label="项目负责人" width="120" header-align="center" align="center">
          <template #default="{ row }">
            {{ row.project?.manager?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="assignee" label="任务负责人" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ row.assignee?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTypes[row.status]">{{ statusLabels[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" header-align="center" align="center">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止时间" width="160" header-align="center" align="center">
          <template #default="{ row }">
            <span :class="{ 'overdue': isOverdue(row.deadline, row.status) }">{{ formatDateTime(row.deadline) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="priorityTypes[row.priority]" size="small">{{ priorityLabels[row.priority] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" header-align="center" fixed="right">
          <template #default="{ row }">
            <div style="display: flex; justify-content: center; align-items: center; gap: 4px;">
              <el-button type="primary" link @click="$router.push(`/tasks/${row.id}`)">查看</el-button>
              <el-dropdown v-if="canUpdateStatus(row)" @command="(cmd) => handleStatusChange(row, cmd)">
                <el-button type="primary" link>更新状态<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="row.status === 'not_started'" command="in_progress">开始任务</el-dropdown-item>
                    <el-dropdown-item v-if="row.status === 'in_progress'" command="completed">完成任务</el-dropdown-item>
                    <el-dropdown-item v-if="row.status === 'completed'" command="in_progress">重新开始</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-popconfirm v-if="canDeleteTask(row)" title="确定删除该任务吗？" width="200" @confirm="handleDeleteTask(row.id)">
                <template #reference>
                  <el-button type="danger" link>删除</el-button>
                </template>
              </el-popconfirm>
            </div>
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
        @size-change="fetchTasks"
        @current-change="fetchTasks"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getTasks, updateTaskStatus, deleteTask } from '@/api/task'
import { getUsers } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const userStore = useUserStore()

const loading = ref(false)
const tasks = ref([])
const users = ref([]) // 用户列表

const searchForm = reactive({ keyword: '', assignee_id: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const statusLabels = { not_started: '未开始', in_progress: '进行中', completed: '已完成', rejected: '被驳回' }
const statusTypes = { not_started: 'info', in_progress: 'warning', completed: 'success', rejected: 'danger' }
const priorityLabels = { 1: '高', 2: '中', 3: '低' }
const priorityTypes = { 1: 'danger', 2: 'warning', 3: 'info' }

const formatDate = (dateStr) => dateStr ? dateStr.split('T')[0] : '-'
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit', 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}
const isOverdue = (deadline, status) => {
  if (!deadline || status === 'completed') return false
  return new Date(deadline) < new Date()
}

const canUpdateStatus = (row) => {
  const isTaskAssignee = row.assignee_id === userStore.user.id
  const isProjectManager = row.project?.manager_id === userStore.user.id
  
  // 任务未完成：任务负责人或项目经理都可以更改状态
  if (row.status !== 'completed') {
    return isTaskAssignee || isProjectManager
  }
  
  // 任务已完成：只有项目经理可以重新开始
  return isProjectManager
}

const fetchTasks = async () => {
  loading.value = true
  try {
    const res = await getTasks({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword,
      assignee_id: searchForm.assignee_id,
      status: searchForm.status
    })
    tasks.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取任务列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchUsers = async () => {
  const res = await getUsers({ page: 1, page_size: 100 })
  users.value = res.data?.list || []
}

const handleSearch = () => { pagination.page = 1; fetchTasks() }
const resetSearch = () => { 
  searchForm.keyword = ''
  searchForm.assignee_id = ''
  searchForm.status = ''
  handleSearch()
}

const handleStatusChange = async (row, status) => {
  try {
    await updateTaskStatus(row.id, status)
    ElMessage.success('状态更新成功')
    fetchTasks()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

const canDeleteTask = (row) => {
  // 只有项目经理可以删除任务
  return row.project?.manager_id === userStore.user.id
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

onMounted(() => {
  fetchUsers()
  fetchTasks()
})
</script>

<style scoped>
.search-card { margin-bottom: 20px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
.overdue { color: #F56C6C; }
</style>
