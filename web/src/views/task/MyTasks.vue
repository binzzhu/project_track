<template>
  <div class="my-tasks">
    <el-tabs v-model="activeTab" @tab-change="fetchTasks">
      <el-tab-pane label="待处理" name="pending" />
      <el-tab-pane label="进行中" name="in_progress" />
      <el-tab-pane label="已完成" name="completed" />
      <el-tab-pane label="全部" name="all" />
    </el-tabs>

    <el-row :gutter="20">
      <el-col v-for="task in tasks" :key="task.id" :span="8">
        <el-card class="task-card" shadow="hover" @click="$router.push(`/tasks/${task.id}`)">
          <template #header>
            <div class="task-header">
              <span class="task-name">{{ task.task_name }}</span>
              <el-tag :type="statusTypes[task.status]" size="small">{{ statusLabels[task.status] }}</el-tag>
            </div>
          </template>
          <div class="task-body">
            <div class="task-info">
              <el-icon><Folder /></el-icon>
              <span>{{ task.project?.name }}</span>
            </div>
            <div class="task-info">
              <el-icon><Calendar /></el-icon>
              <span :class="{ 'overdue': isOverdue(task) }">{{ formatDate(task.deadline) }}</span>
            </div>
            <div class="task-info">
              <el-icon><Flag /></el-icon>
              <el-tag :type="priorityTypes[task.priority]" size="small">{{ priorityLabels[task.priority] }}优先级</el-tag>
            </div>
          </div>
          <div class="task-footer">
            <el-button 
              v-if="task.status === 'not_started'" 
              type="primary" 
              size="small" 
              @click.stop="handleStatusChange(task, 'in_progress')"
            >开始</el-button>
            <el-button 
              v-if="task.status === 'in_progress'" 
              type="success" 
              size="small"
              @click.stop="handleStatusChange(task, 'completed')"
            >完成</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="tasks.length === 0" description="暂无任务" />

    <el-pagination
      v-if="pagination.total > pagination.pageSize"
      v-model:current-page="pagination.page"
      :total="pagination.total"
      :page-size="pagination.pageSize"
      layout="prev, pager, next"
      class="pagination"
      @current-change="fetchTasks"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getMyTasks, updateTaskStatus } from '@/api/task'
import { ElMessage } from 'element-plus'

const activeTab = ref('pending')
const tasks = ref([])
const pagination = reactive({ page: 1, pageSize: 12, total: 0 })

const statusLabels = { not_started: '未开始', in_progress: '进行中', completed: '已完成', rejected: '被驳回' }
const statusTypes = { not_started: 'info', in_progress: 'warning', completed: 'success', rejected: 'danger' }
const priorityLabels = { 1: '高', 2: '中', 3: '低' }
const priorityTypes = { 1: 'danger', 2: 'warning', 3: 'info' }

const formatDate = (dateStr) => dateStr ? dateStr.split('T')[0] : '无截止日期'
const isOverdue = (task) => {
  if (!task.deadline || task.status === 'completed') return false
  return new Date(task.deadline) < new Date()
}

const fetchTasks = async () => {
  let status = ''
  if (activeTab.value === 'pending') status = 'not_started'
  else if (activeTab.value === 'in_progress') status = 'in_progress'
  else if (activeTab.value === 'completed') status = 'completed'
  
  try {
    const res = await getMyTasks({ page: pagination.page, page_size: pagination.pageSize, status })
    tasks.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取任务失败:', error)
  }
}

const handleStatusChange = async (task, status) => {
  try {
    await updateTaskStatus(task.id, status)
    ElMessage.success('状态更新成功')
    fetchTasks()
  } catch (error) {
    console.error('更新状态失败:', error)
  }
}

onMounted(() => fetchTasks())
</script>

<style scoped>
.my-tasks { max-width: 1200px; }
.task-card { margin-bottom: 20px; cursor: pointer; }
.task-header { display: flex; justify-content: space-between; align-items: center; }
.task-name { font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 180px; }
.task-body { min-height: 80px; }
.task-info { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; font-size: 13px; color: #666; }
.task-footer { padding-top: 10px; border-top: 1px solid #eee; }
.overdue { color: #F56C6C; }
.pagination { margin-top: 20px; justify-content: center; }
</style>
