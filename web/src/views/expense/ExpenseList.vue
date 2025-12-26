<template>
  <div class="expense-list">
    <!-- 费用报表统计 -->
    <el-card class="statistics-card">
      <template #header>
        <div class="card-header">
          <span>费用执行情况报表</span>
          <el-button type="primary" size="small" @click="fetchComparison">
            <el-icon><Refresh /></el-icon> 刷新
          </el-button>
        </div>
      </template>
      <el-table v-loading="statsLoading" :data="comparisonData" border stripe max-height="400">
        <el-table-column prop="project_name" label="项目名称" min-width="180" align="center" fixed />
        <el-table-column label="人工费用" align="center">
          <el-table-column prop="labor_budget" label="预算(元)" width="120" align="center">
            <template #default="{ row }">{{ row.labor_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="labor_actual" label="实际(元)" width="120" align="center">
            <template #default="{ row }">{{ row.labor_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="直接投入费用" align="center">
          <el-table-column prop="direct_budget" label="预算(元)" width="120" align="center">
            <template #default="{ row }">{{ row.direct_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="direct_actual" label="实际(元)" width="120" align="center">
            <template #default="{ row }">{{ row.direct_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="委托研发费用" align="center">
          <el-table-column prop="outsourcing_budget" label="预算(元)" width="120" align="center">
            <template #default="{ row }">{{ row.outsourcing_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="outsourcing_actual" label="实际(元)" width="120" align="center">
            <template #default="{ row }">{{ row.outsourcing_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="其他费用" align="center">
          <el-table-column prop="other_budget" label="预算(元)" width="120" align="center">
            <template #default="{ row }">{{ row.other_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="other_actual" label="实际(元)" width="120" align="center">
            <template #default="{ row }">{{ row.other_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="合计" align="center" fixed="right">
          <el-table-column prop="total_budget" label="预算(元)" width="130" align="center">
            <template #default="{ row }">{{ row.total_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="total_actual" label="实际(元)" width="130" align="center">
            <template #default="{ row }">{{ row.total_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="项目">
          <el-select v-model="searchForm.project_id" placeholder="全部项目" clearable filterable style="width: 200px;">
            <el-option v-for="project in projects" :key="project.id" :label="project.name" :value="project.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="费用类型">
          <el-select v-model="searchForm.expense_type" placeholder="全部" clearable style="width: 150px;">
            <el-option label="人工费用" value="labor" />
            <el-option label="直接投入费用" value="direct" />
            <el-option label="委托研发费用" value="outsourcing" />
            <el-option label="其他费用" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <div class="action-bar">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon> 添加费用记录
      </el-button>
    </div>

    <el-card>
      <el-table v-loading="loading" :data="expenses" stripe :expand-row-keys="expandedRows" row-key="id">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div v-if="hasVoucher(row)" class="voucher-expand">
              <div class="voucher-title">凭证文件：</div>
              <div class="voucher-files">
                <div 
                  v-for="(voucher, index) in getVoucherList(row)" 
                  :key="index" 
                  class="voucher-item"
                >
                  <el-link 
                    type="primary" 
                    @click="handleDownloadVoucher(row.id, index)"
                    class="voucher-file-link"
                  >
                    <el-icon><Document /></el-icon>
                    {{ voucher.name }}
                  </el-link>
                  <el-button 
                    v-if="canDeleteVoucher(row)"
                    type="danger" 
                    link 
                    size="small"
                    @click="handleDeleteVoucher(row, index)"
                    class="voucher-delete-btn"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="project" label="项目名称" min-width="180" header-align="center" align="center">
          <template #default="{ row }">{{ row.project?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="expense_type" label="费用类型" width="130" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag>{{ expenseTypeLabels[row.expense_type] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="金额(元)" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ row.amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="expense_date" label="费用日期" width="120" header-align="center" align="center">
          <template #default="{ row }">{{ formatDate(row.expense_date) }}</template>
        </el-table-column>
        <el-table-column prop="description" label="费用说明" min-width="150" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="reimbursed_user" label="报账人" width="100" header-align="center" align="center">
          <template #default="{ row }">{{ row.reimbursed_user?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="voucher_path" label="凭证" width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-button 
              v-if="hasVoucher(row)" 
              type="success" 
              link 
              @click="toggleExpand(row)"
            >
              {{ getVoucherCount(row) }}个文件
            </el-button>
            <el-tag v-else type="info" size="small">未上传</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" header-align="center" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canEdit(row)" type="primary" link @click="showEditDialog(row)">编辑</el-button>
            <el-button v-if="canUploadVoucher(row)" type="success" link @click="showVoucherDialog(row)">上传凭证</el-button>
            <el-popconfirm v-if="canDelete(row)" title="确定删除该费用记录吗？" @confirm="handleDelete(row.id)">
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
        @size-change="fetchExpenses"
        @current-change="fetchExpenses"
      />
    </el-card>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑费用记录' : '添加费用记录'" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-form-item label="关联项目" prop="project_id">
          <el-select v-model="form.project_id" placeholder="请选择项目" filterable style="width: 100%;">
            <el-option v-for="project in projects" :key="project.id" :label="project.name" :value="project.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="费用类型" prop="expense_type">
          <el-select v-model="form.expense_type" placeholder="请选择费用类型" style="width: 100%;">
            <el-option label="人工费用" value="labor" />
            <el-option label="直接投入费用" value="direct" />
            <el-option label="委托研发费用" value="outsourcing" />
            <el-option label="其他费用" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="费用金额" prop="amount">
          <el-input-number v-model="form.amount" :min="0" :precision="2" :controls="false" placeholder="请输入金额" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="费用日期" prop="expense_date">
          <el-date-picker v-model="form.expense_date" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="费用说明" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入费用说明" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="其他备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 上传凭证弹窗 -->
    <el-dialog v-model="voucherDialogVisible" title="上传凭证" width="500px">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :on-change="handleFileChange"
        multiple
        drag
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处或<em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">可上传发票、行程单等凭证文件，支持多个文件</div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="voucherDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUploadVoucher">上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getExpenses, createExpense, updateExpense, deleteExpense, uploadVoucher, downloadVoucher, deleteVoucher, getProjectComparison } from '@/api/expense'
import { getProjects } from '@/api/project'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { ElMessageBox } from 'element-plus'

const userStore = useUserStore()

const loading = ref(false)
const statsLoading = ref(false)
const expenses = ref([])
const projects = ref([])
const comparisonData = ref([])
const dialogVisible = ref(false)
const voucherDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const uploading = ref(false)
const editId = ref(null)
const currentExpense = ref(null)
const expandedRows = ref([])
const formRef = ref(null)
const uploadRef = ref(null)
const selectedFiles = ref([])

const searchForm = reactive({ project_id: '', expense_type: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const form = reactive({ project_id: null, expense_type: '', amount: 0, expense_date: '', description: '', remark: '' })

const expenseTypeLabels = { labor: '人工费用', direct: '直接投入费用', outsourcing: '委托研发费用', other: '其他费用' }

const rules = {
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }],
  expense_type: [{ required: true, message: '请选择费用类型', trigger: 'change' }],
  amount: [{ required: true, message: '请输入费用金额', trigger: 'blur' }],
  expense_date: [{ required: true, message: '请选择费用日期', trigger: 'change' }],
  description: [{ required: true, message: '请输入费用说明', trigger: 'blur' }]
}

const formatDate = (dateStr) => dateStr ? dateStr.split('T')[0] : '-'
const formatDateTime = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'

const hasVoucher = (row) => {
  return row.voucher_path && row.voucher_path.trim() !== ''
}

const getVoucherCount = (row) => {
  if (!row.voucher_path) return 0
  return row.voucher_path.split(',').filter(p => p.trim()).length
}

const getVoucherPaths = (row) => {
  if (!row.voucher_path) return []
  return row.voucher_path.split(',').filter(p => p.trim())
}

const getVoucherList = (row) => {
  const paths = getVoucherPaths(row)
  return paths.map((path, index) => ({
    index,
    name: getVoucherFileName(path, index),
    path
  }))
}

const getVoucherFileName = (path, index) => {
  if (!path) return `凭证${index + 1}`
  // 从路径中提取文件名
  const parts = path.split(/[\\\/]/)
  const filename = parts[parts.length - 1]
  
  // 文件名格式：时间戳_原始文件名.ext
  // 移除时间戳前缀，只保留原始文件名
  const match = filename.match(/^\d+_(.+)$/)
  if (match && match[1]) {
    return match[1]  // 返回原始文件名
  }
  
  // 如果不匹配，返回整个文件名
  return filename
}

const toggleExpand = (row) => {
  const index = expandedRows.value.indexOf(row.id)
  if (index > -1) {
    expandedRows.value.splice(index, 1)
  } else {
    expandedRows.value = [row.id]
  }
}

const getRateType = (rate) => {
  if (!rate || rate === 0) return 'info'
  if (rate < 50) return 'danger'
  if (rate < 80) return 'warning'
  if (rate <= 100) return 'success'
  return 'danger' // 超预算
}

const canEdit = (row) => {
  // 只有系统管理员和费用记录创建者有权限编辑
  if (userStore.isAdmin) return true
  return row.reimbursed_by === userStore.userId
}

const canDelete = (row) => {
  // 只有系统管理员和费用记录创建者有权限删除
  if (userStore.isAdmin) return true
  return row.reimbursed_by === userStore.userId
}

const canUploadVoucher = (row) => {
  // 只有系统管理员和费用记录创建者有权限上传凭证
  if (userStore.isAdmin) return true
  return row.reimbursed_by === userStore.userId
}

const canDeleteVoucher = (row) => {
  // 只有系统管理员和费用记录创建者有权限删除凭证
  if (userStore.isAdmin) return true
  return row.reimbursed_by === userStore.userId
}

const fetchComparison = async () => {
  statsLoading.value = true
  try {
    const res = await getProjectComparison()
    comparisonData.value = res.data || []
  } catch (error) {
    console.error('获取费用统计失败:', error)
  } finally {
    statsLoading.value = false
  }
}

const handleDownloadVoucher = async (expenseId, index) => {
  try {
    const response = await downloadVoucher(expenseId, index)
    const blob = new Blob([response.data])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    
    // 使用友好的文件名
    const expense = expenses.value.find(e => e.id === expenseId)
    const voucherList = expense ? getVoucherList(expense) : []
    const filename = voucherList.find(v => v.index === index)?.name || `凭证_${index + 1}`
    link.download = filename
    
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('下载成功')
  } catch (error) {
    console.error('下载凭证失败:', error)
    ElMessage.error('下载失败')
  }
}

const handleDeleteVoucher = async (row, index) => {
  try {
    await ElMessageBox.confirm('确定要删除该凭证文件吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteVoucher(row.id, index)
    ElMessage.success('删除成功')
    
    // 刷新列表
    await fetchExpenses()
    
    // 如果该费用记录还有凭证，保持展开状态
    const expense = expenses.value.find(e => e.id === row.id)
    if (expense && hasVoucher(expense)) {
      expandedRows.value = [row.id]
    } else {
      expandedRows.value = []
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除凭证失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const fetchProjects = async () => {
  try {
    const res = await getProjects({ page: 1, page_size: 1000 })
    projects.value = res.data?.list || []
  } catch (error) {
    console.error('获取项目列表失败:', error)
  }
}

const fetchExpenses = async () => {
  loading.value = true
  try {
    const res = await getExpenses({
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
    })
    expenses.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取费用列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => { pagination.page = 1; fetchExpenses() }
const resetSearch = () => { Object.assign(searchForm, { project_id: '', expense_type: '' }); handleSearch() }

const showCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, { project_id: null, expense_type: '', amount: 0, expense_date: '', description: '', remark: '' })
  dialogVisible.value = true
}

const showEditDialog = (row) => {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, {
    project_id: row.project_id,
    expense_type: row.expense_type,
    amount: row.amount,
    expense_date: formatDate(row.expense_date),
    description: row.description,
    remark: row.remark
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
        await updateExpense(editId.value, form)
        ElMessage.success('更新成功')
      } else {
        await createExpense(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchExpenses()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (id) => {
  try {
    await deleteExpense(id)
    ElMessage.success('删除成功')
    fetchExpenses()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

const showVoucherDialog = (row) => {
  currentExpense.value = row
  selectedFiles.value = []
  voucherDialogVisible.value = true
}

const handleFileChange = (file, fileList) => {
  selectedFiles.value = fileList
}

const handleUploadVoucher = async () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请选择文件')
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    selectedFiles.value.forEach(file => {
      formData.append('files', file.raw)
    })

    await uploadVoucher(currentExpense.value.id, formData)
    ElMessage.success('上传成功')
    voucherDialogVisible.value = false
    fetchExpenses()
  } catch (error) {
    console.error('上传失败:', error)
  } finally {
    uploading.value = false
  }
}

onMounted(() => {
  fetchProjects()
  fetchExpenses()
  fetchComparison()
})
</script>

<style scoped>
.statistics-card {
  margin-bottom: 20px;
}

.statistics-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-card { margin-bottom: 20px; }
.action-bar { margin-bottom: 20px; }
.pagination { margin-top: 20px; justify-content: flex-end; }

.voucher-expand {
  padding: 10px 50px;
  background-color: #f5f7fa;
}

.voucher-title {
  font-weight: 500;
  color: #606266;
  margin-bottom: 10px;
}

.voucher-files {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.voucher-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.voucher-file-link {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 8px 12px;
  background-color: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  transition: all 0.3s;
}

.voucher-file-link:hover {
  background-color: #ecf5ff;
  border-color: #409eff;
}

.voucher-delete-btn {
  padding: 4px;
  opacity: 0.6;
  transition: opacity 0.3s;
}

.voucher-delete-btn:hover {
  opacity: 1;
}
</style>
