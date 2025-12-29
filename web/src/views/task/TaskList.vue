<template>
  <div class="task-list">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="任务名称" clearable @keyup.enter="handleSearch" />
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
        <el-table-column prop="assignee" label="负责人" width="100" header-align="center" align="center">
          <template #default="{ row }">{{ row.assignee?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTypes[row.status]">{{ statusLabels[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止日期" width="120" header-align="center" align="center">
          <template #default="{ row }">
            <span :class="{ 'overdue': isOverdue(row.deadline, row.status) }">{{ formatDate(row.deadline) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="priorityTypes[row.priority]" size="small">{{ priorityLabels[row.priority] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" header-align="center" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="$router.push(`/tasks/${row.id}`)">查看</el-button>
            <el-dropdown v-if="canUpdateStatus(row)" @command="(cmd) => handleStatusChange(row, cmd)">
              <el-button type="primary" link>更新状态<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="row.status === 'not_started'" command="in_progress">开始任务</el-dropdown-item>
                  <el-dropdown-item v-if="row.status === 'in_progress'" command="completed">完成任务</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
import { getTasks, updateTaskStatus } from '@/api/task'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const userStore = useUserStore()

const loading = ref(false)
const tasks = ref([])

const searchForm = reactive({ keyword: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const statusLabels = { not_started: '未开始', in_progress: '进行中', completed: '已完成', rejected: '被驳回' }
const statusTypes = { not_started: 'info', in_progress: 'warning', completed: 'success', rejected: 'danger' }
const priorityLabels = { 1: '高', 2: '中', 3: '低' }
const priorityTypes = { 1: 'danger', 2: 'warning', 3: 'info' }

const formatDate = (dateStr) => dateStr ? dateStr.split('T')[0] : '-'
const isOverdue = (deadline, status) => {
  if (!deadline || status === 'completed') return false
  return new Date(deadline) < new Date()
}

const canUpdateStatus = (row) => {
  // 只有任务负责人有权限更改任务状态
  return row.assignee_id === userStore.user.id
}

const fetchTasks = async () => {
  loading.value = true
  try {
    const res = await getTasks({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword,
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

const handleSearch = () => { pagination.page = 1; fetchTasks() }
const resetSearch = () => { searchForm.keyword = ''; searchForm.status = ''; handleSearch() }

const handleStatusChange = async (row, status) => {
  try {
    await updateTaskStatus(row.id, status)
    ElMessage.success('状态更新成功')
    fetchTasks()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

onMounted(() => fetchTasks())
</script>

<style scoped>
.search-card { margin-bottom: 20px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
.overdue { color: #F56C6C; }
</style>
