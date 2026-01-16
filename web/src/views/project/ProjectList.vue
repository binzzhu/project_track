<template>
  <div class="project-list">
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="项目名称/编号" clearable @keyup.enter="handleSearch" style="width: 220px;" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 150px;">
            <el-option label="未开始" value="not_started" />
            <el-option label="进行中" value="in_progress" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="阶段">
          <el-select v-model="searchForm.phase" placeholder="全部" clearable style="width: 160px;">
            <el-option v-for="(label, key) in phaseLabels" :key="key" :label="label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="项目负责人">
          <el-select v-model="searchForm.manager_id" placeholder="请选择项目负责人" clearable filterable style="width: 180px;">
            <el-option v-for="user in users" :key="user.id" :label="user.name" :value="user.id" />
          </el-select>
        </el-form-item>
        <el-form-item class="search-button-item">
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button v-if="userStore.canCreateProject" type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon> 创建项目
      </el-button>
    </div>

    <!-- 项目列表 -->
    <el-card>
      <el-table v-loading="loading" :data="projects" stripe>
        <el-table-column prop="name" label="项目名称" min-width="200" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="$router.push(`/projects/${row.id}`)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="project_type" label="项目类型" width="100" header-align="center" align="center" />
        <el-table-column prop="current_phase" label="当前阶段" width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag>{{ phaseLabels[row.current_phase] || row.current_phase }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTypes[row.status]">{{ statusLabels[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="manager" label="项目负责人" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ row.manager?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="contract_no" label="合同编号" min-width="260" header-align="center" align="center">
          <template #default="{ row }">
            <div v-if="row.contract_no">
              <div v-for="(no, index) in splitContractNos(row.contract_no)" :key="index">{{ no }}</div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="budget_code" label="预算编码" min-width="260" header-align="center" align="center">
          <template #default="{ row }">{{ row.budget_code || '-' }}</template>
        </el-table-column>
        <el-table-column prop="innovation_code" label="创新项目编码" min-width="260" header-align="center" align="center">
          <template #default="{ row }">{{ row.innovation_code || '-' }}</template>
        </el-table-column>
        <el-table-column prop="initiation_date" label="立项日期" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ formatDate(row.initiation_date) }}</template>
        </el-table-column>
        <el-table-column prop="closing_date" label="结项日期" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ formatDate(row.closing_date) }}</template>
        </el-table-column>
        <el-table-column prop="labor_cost" label="人工费用" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ row.labor_cost ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="direct_cost" label="直接投入费用" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ row.direct_cost ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="outsourcing_cost" label="委托研发费用" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ row.outsourcing_cost ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="other_cost" label="其他费用" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ row.other_cost ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" header-align="center" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="$router.push(`/projects/${row.id}`)">查看</el-button>
            <el-button v-if="canEditProject(row)" type="primary" link @click="showEditDialog(row)">编辑</el-button>
            <el-popconfirm v-if="canEditProject(row)" title="确定删除该项目吗？" width="200" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button type="danger" link>删除</el-button>
              </template>
            </el-popconfirm>
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
        @size-change="fetchProjects"
        @current-change="fetchProjects"
      />
    </el-card>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑项目' : '创建项目'" width="700px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-form-item label="项目名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入项目名称" />
        </el-form-item>
        <el-form-item label="项目编号">
          <el-input v-model="form.project_no" placeholder="留空自动生成" />
        </el-form-item>
        <el-form-item label="项目类型" prop="project_type">
          <el-select v-model="form.project_type" placeholder="请选择项目类型">
            <el-option label="成本性" value="成本性" />
            <el-option label="资本性" value="资本性" />
          </el-select>
        </el-form-item>
        <el-form-item label="项目负责人" prop="manager_id">
          <el-select v-model="form.manager_id" placeholder="请选择负责人" filterable>
            <el-option v-for="user in users" :key="user.id" :label="user.name" :value="user.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="合同编号">
          <el-input v-model="form.contract_no" placeholder="多个编号请用英文逗号分隔，立项结束后补充（非必填）" />
        </el-form-item>
        <el-form-item label="预算编码">
          <el-input v-model="form.budget_code" placeholder="立项结束后补充（非必填）" />
        </el-form-item>
        <el-form-item label="创新项目编码">
          <el-input v-model="form.innovation_code" placeholder="立项结束后补充（非必填）" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="立项日期" prop="initiation_date">
              <el-date-picker v-model="form.initiation_date" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结项日期" prop="closing_date">
              <el-date-picker v-model="form.closing_date" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="人工费用" prop="labor_cost">
              <el-input-number v-model="form.labor_cost" :min="0" :precision="2" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="直接投入费用" prop="direct_cost">
              <el-input-number v-model="form.direct_cost" :min="0" :precision="2" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="委托研发费用" prop="outsourcing_cost">
              <el-input-number v-model="form.outsourcing_cost" :min="0" :precision="2" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="其他费用" prop="other_cost">
              <el-input-number v-model="form.other_cost" :min="0" :precision="2" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getProjects, createProject, updateProject, deleteProject } from '@/api/project'
import { getUsers } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const userStore = useUserStore()

const loading = ref(false)
const projects = ref([])
const users = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref(null)
const formRef = ref(null)

const searchForm = reactive({
  keyword: '',
  status: '',
  phase: '',
  manager_id: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  name: '',
  project_no: '',
  project_type: '',
  manager_id: null,
  contract_no: '',
  budget_code: '',
  innovation_code: '',
  initiation_date: '',
  closing_date: '',
  labor_cost: 0,
  direct_cost: 0,
  outsourcing_cost: 0,
  other_cost: 0
})

const rules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  project_type: [{ required: true, message: '请选择项目类型', trigger: 'change' }],
  manager_id: [{ required: true, message: '请选择项目负责人', trigger: 'change' }],
  initiation_date: [{ required: true, message: '请选择立项日期', trigger: 'change' }],
  closing_date: [{ required: true, message: '请选择结项日期', trigger: 'change' }],
  labor_cost: [{ required: true, message: '请输入人工费用', trigger: 'blur' }],
  direct_cost: [{ required: true, message: '请输入直接投入费用', trigger: 'blur' }],
  outsourcing_cost: [{ required: true, message: '请输入委托研发费用', trigger: 'blur' }],
  other_cost: [{ required: true, message: '请输入其他费用', trigger: 'blur' }]
}

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
  completed: '已完成'
}

const statusTypes = {
  not_started: 'info',
  in_progress: 'primary',
  completed: 'success'
}

const splitContractNos = (val) => {
  if (!val) return []
  return val.split(',').map(item => item.trim()).filter(item => item)
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

// 检查是否可以编辑项目（管理员或项目负责人）
const canEditProject = (project) => {
  if (userStore.isAdmin) return true
  return project.manager_id === userStore.userId
}

const fetchProjects = async () => {
  loading.value = true
  try {
    const res = await getProjects({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword,
      status: searchForm.status,
      phase: searchForm.phase,
      manager_id: searchForm.manager_id
    })
    projects.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('获取项目列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchUsers = async () => {
  try {
    const res = await getUsers({ page: 1, page_size: 100 })
    users.value = res.data.list || []
  } catch (error) {
    console.error('获取用户列表失败:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchProjects()
}

const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.status = ''
  searchForm.phase = ''
  searchForm.manager_id = ''
  handleSearch()
}

const showCreateDialog = () => {
  isEdit.value = false
  editId.value = null
  Object.assign(form, {
    name: '', project_no: '', project_type: '', manager_id: null,
    contract_no: '', budget_code: '', innovation_code: '',
    initiation_date: '', closing_date: '', labor_cost: 0, direct_cost: 0, outsourcing_cost: 0, other_cost: 0
  })
  dialogVisible.value = true
}

const showEditDialog = (row) => {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, {
    name: row.name,
    project_no: row.project_no,
    project_type: row.project_type,
    manager_id: row.manager_id,
    contract_no: row.contract_no || '',
    budget_code: row.budget_code || '',
    innovation_code: row.innovation_code || '',
    initiation_date: row.initiation_date?.split('T')[0] || '',
    closing_date: row.closing_date?.split('T')[0] || '',
    labor_cost: row.labor_cost,
    direct_cost: row.direct_cost,
    outsourcing_cost: row.outsourcing_cost,
    other_cost: row.other_cost
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updateProject(editId.value, form)
        ElMessage.success('更新成功')
      } else {
        await createProject(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchProjects()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (id) => {
  try {
    await deleteProject(id)
    ElMessage.success('删除成功')
    fetchProjects()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

onMounted(() => {
  fetchProjects()
  fetchUsers()
})
</script>

<style scoped>
.search-card {
  margin-bottom: 20px;
}
.search-form {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
}
.search-form .el-form-item {
  margin-bottom: 0;
}
.search-button-item {
  margin-left: auto;
  margin-bottom: 0 !important;
}
.action-bar {
  margin-bottom: 15px;
}
.pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
