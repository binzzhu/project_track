<template>
  <div class="log-list">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="操作类型">
          <el-select v-model="searchForm.action" placeholder="全部" clearable>
            <el-option v-for="action in actions" :key="action.value" :label="action.label" :value="action.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="模块">
          <el-select v-model="searchForm.module" placeholder="全部" clearable>
            <el-option v-for="mod in modules" :key="mod.value" :label="mod.label" :value="mod.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="描述/目标名称" clearable />
        </el-form-item>
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <el-table v-loading="loading" :data="logs" stripe>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="user" label="操作人" width="100">
          <template #default="{ row }">{{ row.user?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="action" label="操作类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ actionLabels[row.action] || row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="module" label="模块" width="100">
          <template #default="{ row }">{{ moduleLabels[row.module] || row.module }}</template>
        </el-table-column>
        <el-table-column prop="target_name" label="操作对象" width="150" />
        <el-table-column prop="description" label="操作描述" min-width="250" />
        <el-table-column prop="result" label="结果" width="80">
          <template #default="{ row }">
            <el-tag :type="row.result === 'success' ? 'success' : 'danger'" size="small">
              {{ row.result === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="130" />
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        class="pagination"
        @size-change="fetchLogs"
        @current-change="fetchLogs"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getLogs, getLogActions, getLogModules } from '@/api/log'

const loading = ref(false)
const logs = ref([])
const actions = ref([])
const modules = ref([])
const dateRange = ref([])

const searchForm = reactive({ action: '', module: '', keyword: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const actionLabels = { login: '登录', logout: '退出', create: '创建', update: '更新', delete: '删除', upload: '上传', download: '下载', review: '审核', archive: '归档', add_member: '添加成员', remove_member: '移除成员', update_phase: '更新阶段', update_status: '更新状态', change_password: '修改密码', reset_password: '重置密码', new_version: '上传新版本' }
const moduleLabels = { auth: '认证', user: '用户管理', project: '项目管理', task: '任务管理', document: '文档管理', knowledge: '知识库', contract: '合同管理' }

const formatDateTime = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'

const fetchLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      action: searchForm.action,
      module: searchForm.module,
      keyword: searchForm.keyword
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await getLogs(params)
    logs.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取日志失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchOptions = async () => {
  const [actionsRes, modulesRes] = await Promise.all([getLogActions(), getLogModules()])
  actions.value = actionsRes.data || []
  modules.value = modulesRes.data || []
}

const handleSearch = () => { pagination.page = 1; fetchLogs() }
const resetSearch = () => {
  searchForm.action = ''
  searchForm.module = ''
  searchForm.keyword = ''
  dateRange.value = []
  handleSearch()
}

onMounted(() => {
  fetchOptions()
  fetchLogs()
})
</script>

<style scoped>
.search-card { margin-bottom: 20px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
</style>
