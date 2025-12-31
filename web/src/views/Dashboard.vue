<template>
  <div class="dashboard">
    <!-- 项目统计卡片 -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="12">
        <el-card shadow="hover" class="stat-card" style="height: 200px; display: flex; flex-direction: column; justify-content: center;">
          <div class="stat-header">
            <span class="stat-title">项目统计</span>
          </div>
          <div class="stat-content-wrapper">
            <el-row :gutter="20" style="margin-top: 10px;">
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #409EFF;">
                    <el-icon size="28"><Folder /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ projectStats.total || 0 }}</div>
                    <div class="stat-label">项目总数</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #E6A23C;">
                    <el-icon size="28"><Clock /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ projectStats.not_started || 0 }}</div>
                    <div class="stat-label">未开始</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #67C23A;">
                    <el-icon size="28"><Loading /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ projectStats.in_progress || 0 }}</div>
                    <div class="stat-label">进行中</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #909399;">
                    <el-icon size="28"><CircleCheck /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ projectStats.completed || 0 }}</div>
                    <div class="stat-label">已完成</div>
                  </div>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover" class="stat-card" style="height: 200px; display: flex; flex-direction: column; justify-content: center;">
          <div class="stat-header">
            <span class="stat-title">任务统计</span>
          </div>
          <div class="stat-content-wrapper">
            <el-row :gutter="20" style="margin-top: 10px;">
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #409EFF;">
                    <el-icon size="28"><List /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ taskStats.total || 0 }}</div>
                    <div class="stat-label">任务总数</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #E6A23C;">
                    <el-icon size="28"><Clock /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ taskStats.not_started || 0 }}</div>
                    <div class="stat-label">未开始</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #67C23A;">
                    <el-icon size="28"><Loading /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ taskStats.in_progress || 0 }}</div>
                    <div class="stat-label">进行中</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-content">
                  <div class="stat-icon" style="background: #909399;">
                    <el-icon size="28"><CircleCheck /></el-icon>
                  </div>
                  <div class="stat-info">
                    <div class="stat-value">{{ taskStats.completed || 0 }}</div>
                    <div class="stat-label">已完成</div>
                  </div>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="content-row">
      <!-- 最近项目 -->
      <el-col :span="14">
        <el-card class="content-card fixed-height-card">
          <template #header>
            <div class="card-header">
              <span>最近项目</span>
              <el-button type="primary" link @click="$router.push('/projects')">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="table-container">
            <el-table :data="recentProjects" stripe style="width: 100%">
              <el-table-column prop="project_no" label="项目编号" width="150" header-align="center" align="center" />
              <el-table-column prop="name" label="项目名称" min-width="250" header-align="center" align="center">
                <template #default="{ row }">
                  <el-link type="primary" @click="$router.push(`/projects/${row.id}`)">
                    {{ row.name }}
                  </el-link>
                </template>
              </el-table-column>
              <el-table-column prop="current_phase" label="当前阶段" width="100" header-align="center" align="center">
                <template #default="{ row }">
                  <el-tag>{{ phaseLabels[row.current_phase] || row.current_phase }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="90" header-align="center" align="center">
                <template #default="{ row }">
                  <el-tag :type="statusTypes[row.status]">
                    {{ statusLabels[row.status] || row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="manager" label="负责人" width="90" header-align="center" align="center">
                <template #default="{ row }">
                  {{ row.manager?.name || '-' }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-col>

      <!-- 我的待办 -->
      <el-col :span="10">
        <el-card class="content-card fixed-height-card">
          <template #header>
            <div class="card-header">
              <span>我的待办任务</span>
              <el-button type="primary" link @click="$router.push('/my-tasks')">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="table-container">
            <div v-if="myTasks.length === 0" class="empty-tip">
              <el-empty description="暂无待办任务" :image-size="80" />
            </div>
            <el-table v-else :data="myTasks" stripe size="small" style="width: 100%">
              <el-table-column prop="task_name" label="任务名称" min-width="120" header-align="center" align="center">
                <template #default="{ row }">
                  <el-link type="primary" @click="$router.push(`/tasks/${row.id}`)">{{ row.task_name }}</el-link>
                </template>
              </el-table-column>
              <el-table-column prop="project" label="所属项目" min-width="100" header-align="center" align="center">
                <template #default="{ row }">{{ row.project?.name || '-' }}</template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="80" header-align="center" align="center">
                <template #default="{ row }">
                  <el-tag :type="statusTypes[row.status]" size="small">{{ statusLabels[row.status] }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="120" header-align="center" align="center">
                <template #default="{ row }">
                  {{ formatDateTime(row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column prop="deadline" label="截止时间" width="120" header-align="center" align="center">
                <template #default="{ row }">
                  <span :style="{ color: isOverdue(row.deadline) ? '#F56C6C' : '' }">{{ formatDateTime(row.deadline) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="deliverables" label="交付要求" min-width="120" header-align="center" align="center" show-overflow-tooltip>
                <template #default="{ row }">{{ row.deliverables || '-' }}</template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 费用统计图表 -->
    <el-row :gutter="20" class="content-row">
      <!-- 研发项目费用执行情况统计 (100%) -->
      <el-col :span="24">
        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>研发项目费用执行情况统计</span>
              <div style="display: flex; align-items: center; gap: 10px;">
                <el-select 
                  v-model="selectedProjectId" 
                  placeholder="请选择项目" 
                  filterable 
                  style="width: 280px;"
                  @change="updateExpenseChart"
                >
                  <el-option
                    v-for="project in expenseData"
                    :key="project.project_id"
                    :label="project.project_name"
                    :value="project.project_id"
                  />
                </el-select>
                <el-button type="primary" link @click="$router.push('/expenses')">
                  查看详情 <el-icon><ArrowRight /></el-icon>
                </el-button>
              </div>
            </div>
          </template>
          <div ref="expenseChartRef" style="width: 100%; height: 350px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="content-row">
      <!-- 非研发项目费用统计 (100%) -->
      <el-col :span="24">
        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>非研发项目费用统计</span>
              <el-button type="primary" link @click="$router.push('/expenses')">
                查看详情 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div ref="nonProjectChartRef" style="width: 100%; height: 400px;"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { getProjectStatistics, getProjects } from '@/api/project'
import { getMyTasks, getTaskStatistics } from '@/api/task'
import { getProjectComparison, getNonProjectExpenseStats } from '@/api/expense'
import * as echarts from 'echarts'

const projectStats = ref({})
const taskStats = ref({})
const recentProjects = ref([])
const myTasks = ref([])
const expenseChartRef = ref(null)
const nonProjectChartRef = ref(null)
const expenseData = ref([])
const nonProjectStatsData = ref([])
const selectedProjectId = ref(null)
let expenseChart = null
let nonProjectChart = null

const phaseLabels = {
  initiation: '立项',
  bidding: '招标',
  contract: '合同签订',
  acceptance: '验收',
  closing: '结项'
}

const statusLabels = {
  not_started: '未开始',
  in_progress: '进行中',
  completed: '已完成',
  rejected: '被驳回'
}

const statusTypes = {
  not_started: 'info',
  in_progress: 'primary',
  completed: 'success',
  rejected: 'danger'
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', { 
    month: '2-digit', 
    day: '2-digit',
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const isOverdue = (deadline) => {
  if (!deadline) return false
  return new Date(deadline) < new Date()
}

const fetchData = async () => {
  try {
    const [projectStatsRes, taskStatsRes, projectsRes, tasksRes, expenseRes, nonProjectRes] = await Promise.all([
      getProjectStatistics(),
      getTaskStatistics(),
      getProjects({ page: 1, page_size: 5 }),
      getMyTasks({ page: 1, page_size: 5, status: 'not_started,in_progress' }),
      getProjectComparison(),
      getNonProjectExpenseStats()
    ])
    
    projectStats.value = projectStatsRes.data || {}
    taskStats.value = taskStatsRes.data || {}
    recentProjects.value = projectsRes.data?.list || []
    myTasks.value = tasksRes.data?.list || []
    expenseData.value = expenseRes.data || []
    nonProjectStatsData.value = nonProjectRes.data?.data || []
    
    // 默认选中第一个项目
    if (expenseData.value.length > 0) {
      selectedProjectId.value = expenseData.value[0].project_id
    }
    
    // 渲染图表
    await nextTick()
    initExpenseChart()
    initNonProjectChart()
  } catch (error) {
    console.error('获取数据失败:', error)
  }
}

const updateExpenseChart = () => {
  initExpenseChart()
}

const initExpenseChart = () => {
  if (!expenseChartRef.value || !selectedProjectId.value) return
  
  // 找到选中的项目数据
  const selectedProject = expenseData.value.find(item => item.project_id === selectedProjectId.value)
  if (!selectedProject) return
  
  // 如果图表已存在，先销毁
  if (expenseChart) {
    expenseChart.dispose()
  }
  
  expenseChart = echarts.init(expenseChartRef.value)
  
  // 准备图表数据 - 单个项目的各类费用
  const categories = ['人工费用', '直接投入费用', '委托研发费用', '其他费用']
  const budgetData = [
    selectedProject.labor_budget || 0,
    selectedProject.direct_budget || 0,
    selectedProject.outsourcing_budget || 0,
    selectedProject.other_budget || 0
  ]
  const actualInclTaxData = [
    selectedProject.labor_actual_incl_tax || 0,
    selectedProject.direct_actual_incl_tax || 0,
    selectedProject.outsourcing_actual_incl_tax || 0,
    selectedProject.other_actual_incl_tax || 0
  ]
  const actualExclTaxData = [
    selectedProject.labor_actual_excl_tax || 0,
    selectedProject.direct_actual_excl_tax || 0,
    selectedProject.outsourcing_actual_excl_tax || 0,
    selectedProject.other_actual_excl_tax || 0
  ]
  
  const option = {
    title: {
      text: selectedProject.project_name + ' - 费用执行情况',
      left: 'center',
      top: 10
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      },
      formatter: function(params) {
        let result = params[0].axisValue + '<br/>'
        params.forEach(item => {
          const valueInWan = (item.value / 10000).toFixed(2)
          result += item.marker + item.seriesName + ': ' + valueInWan + '万元<br/>'
        })
        return result
      }
    },
    legend: {
      data: ['预算', '实际(含税)', '实际(不含税)'],
      top: 45
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: 100,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: categories,
      axisLabel: {
        interval: 0,
        fontSize: 13
      }
    },
    yAxis: {
      type: 'value',
      name: '金额(万元)',
      axisLabel: {
        formatter: function(value) {
          return (value / 10000).toFixed(1)
        }
      }
    },
    series: [
      {
        name: '预算',
        type: 'bar',
        data: budgetData,
        itemStyle: { 
          color: '#409EFF',
          borderRadius: [4, 4, 0, 0]
        },
        barWidth: '20%',
        label: {
          show: true,
          position: 'top',
          formatter: function(params) {
            return (params.value / 10000).toFixed(1) + '万'
          },
          fontSize: 11
        }
      },
      {
        name: '实际(含税)',
        type: 'bar',
        data: actualInclTaxData,
        itemStyle: { 
          color: '#67C23A',
          borderRadius: [4, 4, 0, 0]
        },
        barWidth: '20%',
        label: {
          show: true,
          position: 'top',
          formatter: function(params) {
            return (params.value / 10000).toFixed(1) + '万'
          },
          fontSize: 11
        }
      },
      {
        name: '实际(不含税)',
        type: 'bar',
        data: actualExclTaxData,
        itemStyle: { 
          color: '#E6A23C',
          borderRadius: [4, 4, 0, 0]
        },
        barWidth: '20%',
        label: {
          show: true,
          position: 'top',
          formatter: function(params) {
            return (params.value / 10000).toFixed(1) + '万'
          },
          fontSize: 11
        }
      }
    ]
  }
  
  expenseChart.setOption(option)
  
  // 响应式调整
  window.addEventListener('resize', () => {
    expenseChart?.resize()
  })
}

const initNonProjectChart = () => {
  if (!nonProjectChartRef.value || nonProjectStatsData.value.length === 0) return
  
  // 如果图表已存在，先销毁
  if (nonProjectChart) {
    nonProjectChart.dispose()
  }
  
  nonProjectChart = echarts.init(nonProjectChartRef.value)
  
  // 准备图表数据
  const businessScenes = nonProjectStatsData.value.map(item => item.business_scene || '(未填写)')
  const amountsInclTax = nonProjectStatsData.value.map(item => item.total_incl_tax || 0)
  const amountsExclTax = nonProjectStatsData.value.map(item => item.total_excl_tax || 0)
  
  // 计算总计
  const grandTotalInclTax = amountsInclTax.reduce((sum, val) => sum + val, 0)
  const grandTotalExclTax = amountsExclTax.reduce((sum, val) => sum + val, 0)
  
  const option = {
    title: {
      text: '非研发项目费用分布',
      subtext: `含税总计: ${(grandTotalInclTax / 10000).toFixed(2)}万元 | 不含税总计: ${(grandTotalExclTax / 10000).toFixed(2)}万元`,
      left: 'center',
      top: 10
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      },
      formatter: function(params) {
        let result = params[0].axisValue + '<br/>'
        params.forEach(item => {
          const valueInWan = (item.value / 10000).toFixed(2)
          result += item.marker + item.seriesName + ': ' + valueInWan + '万元<br/>'
        })
        return result
      }
    },
    legend: {
      data: ['含税费用', '不含税费用'],
      top: 50
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: 100,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: businessScenes,
      axisLabel: {
        interval: 0,
        rotate: businessScenes.length > 5 ? 30 : 0,
        fontSize: 12
      }
    },
    yAxis: {
      type: 'value',
      name: '金额(万元)',
      axisLabel: {
        formatter: function(value) {
          return (value / 10000).toFixed(1)
        }
      }
    },
    series: [
      {
        name: '含税费用',
        type: 'bar',
        data: amountsInclTax,
        itemStyle: { 
          color: '#E6A23C',
          borderRadius: [4, 4, 0, 0]
        },
        barWidth: '35%',
        label: {
          show: true,
          position: 'top',
          formatter: function(params) {
            return (params.value / 10000).toFixed(1) + '万'
          },
          fontSize: 11
        }
      },
      {
        name: '不含税费用',
        type: 'bar',
        data: amountsExclTax,
        itemStyle: { 
          color: '#67C23A',
          borderRadius: [4, 4, 0, 0]
        },
        barWidth: '35%',
        label: {
          show: true,
          position: 'top',
          formatter: function(params) {
            return (params.value / 10000).toFixed(1) + '万'
          },
          fontSize: 11
        }
      }
    ]
  }
  
  nonProjectChart.setOption(option)
  
  // 响应式调整
  window.addEventListener('resize', () => {
    nonProjectChart?.resize()
  })
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.dashboard {
  max-width: 1400px;
}

.stat-cards {
  margin-bottom: 20px;
}

.stat-card {
  border-radius: 8px;
}

.stat-header {
  padding: 10px 15px;
  border-bottom: 1px solid #eee;
}

.stat-title {
  font-size: 16px;
  font-weight: 500;
}

.stat-content-wrapper {
  padding: 10px 15px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 15px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #333;
}

.stat-label {
  font-size: 14px;
  color: #999;
  margin-top: 4px;
}

.content-row {
  margin-top: 20px;
}

.content-card {
  border-radius: 8px;
}

.fixed-height-card {
  height: 400px;
  display: flex;
  flex-direction: column;
}

.fixed-height-card :deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.table-container {
  height: 100%;
  overflow-y: auto;
  padding: 0 20px 20px 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-size: 16px;
  font-weight: 500;
}

.empty-tip {
  padding: 20px 0;
}

.task-list {
  max-height: 350px;
  overflow-y: auto;
}

.task-item {
  padding: 12px;
  border-bottom: 1px solid #eee;
  cursor: pointer;
  transition: background 0.2s;
}

.task-item:hover {
  background: #f5f7fa;
}

.task-item:last-child {
  border-bottom: none;
}

.task-name {
  font-size: 14px;
  color: #333;
  margin-bottom: 6px;
}

.task-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.project-name {
  font-size: 12px;
  color: #999;
}
</style>
