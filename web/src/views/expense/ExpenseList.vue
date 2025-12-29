<template>
  <div class="expense-list">
    <!-- 费用执行情况报表 -->
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
        <el-table-column label="项目名称" min-width="200" align="center" fixed>
          <template #default="{ row }">
            <span v-if="row.innovation_code">{{ row.innovation_code }} - {{ row.project_name }}</span>
            <span v-else>{{ row.project_name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="人工费用" align="center">
          <el-table-column prop="labor_budget" label="预算(元)" width="120" align="right">
            <template #default="{ row }">{{ row.labor_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="labor_actual" label="实际(元)" width="120" align="right">
            <template #default="{ row }">{{ row.labor_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="labor_rate" label="执行率(%)" width="100" align="right">
            <template #default="{ row }">{{ row.labor_rate?.toFixed(1) || '0.0' }}%</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="直接投入费用" align="center">
          <el-table-column prop="direct_budget" label="预算(元)" width="120" align="right">
            <template #default="{ row }">{{ row.direct_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="direct_actual" label="实际(元)" width="120" align="right">
            <template #default="{ row }">{{ row.direct_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="direct_rate" label="执行率(%)" width="100" align="right">
            <template #default="{ row }">{{ row.direct_rate?.toFixed(1) || '0.0' }}%</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="委托研发费用" align="center">
          <el-table-column prop="outsourcing_budget" label="预算(元)" width="120" align="right">
            <template #default="{ row }">{{ row.outsourcing_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="outsourcing_actual" label="实际(元)" width="120" align="right">
            <template #default="{ row }">{{ row.outsourcing_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="outsourcing_rate" label="执行率(%)" width="100" align="right">
            <template #default="{ row }">{{ row.outsourcing_rate?.toFixed(1) || '0.0' }}%</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="其他费用" align="center">
          <el-table-column prop="other_budget" label="预算(元)" width="120" align="right">
            <template #default="{ row }">{{ row.other_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="other_actual" label="实际(元)" width="120" align="right">
            <template #default="{ row }">{{ row.other_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="other_rate" label="执行率(%)" width="100" align="right">
            <template #default="{ row }">{{ row.other_rate?.toFixed(1) || '0.0' }}%</template>
          </el-table-column>
        </el-table-column>
        <el-table-column label="合计" align="center" fixed="right">
          <el-table-column prop="total_budget" label="预算(元)" width="130" align="right">
            <template #default="{ row }">{{ row.total_budget?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="total_actual" label="实际(元)" width="130" align="right">
            <template #default="{ row }">{{ row.total_actual?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="total_rate" label="执行率(%)" width="100" align="right">
            <template #default="{ row }">{{ row.total_rate?.toFixed(1) || '0.0' }}%</template>
          </el-table-column>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="项目编码">
          <el-select v-model="searchForm.project_code" placeholder="请选择项目" clearable filterable style="width: 280px;">
            <el-option 
              v-for="project in projects" 
              :key="project.innovation_code" 
              :label="`${project.innovation_code} - ${project.name}`" 
              :value="project.innovation_code" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="报账人">
          <el-input v-model="searchForm.reimbursed_person_name" placeholder="请输入报账人姓名" clearable style="width: 150px;" />
        </el-form-item>
        <el-form-item label="是否项目开支">
          <el-select v-model="searchForm.is_project_expense" placeholder="全部" clearable style="width: 130px;">
            <el-option label="是" value="是" />
            <el-option label="否" value="否" />
          </el-select>
        </el-form-item>
        <el-form-item label="单据编号">
          <el-input v-model="searchForm.document_no" placeholder="请输入单据编号" clearable style="width: 180px;" />
        </el-form-item>
        <el-form-item label="归类状态">
          <el-select v-model="searchForm.is_classified" placeholder="全部" clearable style="width: 120px;">
            <el-option label="已归类" :value="true" />
            <el-option label="未归类" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item class="search-button-item">
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <div class="action-bar">
      <el-button type="success" @click="showImportDialog">
        <el-icon><Upload /></el-icon> 导入Excel
      </el-button>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon> 添加费用记录
      </el-button>
      <el-button v-if="userStore.isAdmin" type="danger" @click="handleDeleteAll">
        <el-icon><Delete /></el-icon> 一键删除所有记录
      </el-button>
    </div>

    <el-card>
      <el-table v-loading="loading" :data="expenses" stripe border>
        <el-table-column label="关联项目" width="280" header-align="center" align="center" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.project && row.project.innovation_code">
              {{ row.project.innovation_code }} - {{ row.project.name }}
            </span>
            <span v-else-if="row.project_code">
              {{ row.project_code }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="费用类型" width="130" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.expense_type === 'labor'" type="success" size="small">人工费用</el-tag>
            <el-tag v-else-if="row.expense_type === 'direct'" type="primary" size="small">直接投入费用</el-tag>
            <el-tag v-else-if="row.expense_type === 'outsourcing'" type="warning" size="small">委托研发费用</el-tag>
            <el-tag v-else-if="row.expense_type === 'other'" type="info" size="small">其他费用</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="unit_code" label="单位编号" width="100" header-align="center" align="center" />
        <el-table-column prop="unit_name" label="单位名称" width="200" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="document_no" label="单据编号" width="180" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="business_scene" label="业务场景" width="120" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="cross_industry_code" label="跨行业业务编码" width="150" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="summary" label="摘要" min-width="200" header-align="center" align="left" show-overflow-tooltip />
        <el-table-column prop="department_name" label="部门名称" width="120" header-align="center" align="center" />
        <el-table-column prop="reimbursed_person_name" label="报账人" width="100" header-align="center" align="center" />
        <el-table-column prop="document_status" label="单据状态" width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.document_status" size="small">{{ row.document_status }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="frozen_status" label="冻结状态" width="100" header-align="center" align="center" />
        <el-table-column prop="reimbursement_amount" label="报账金额" width="110" header-align="center" align="right">
          <template #default="{ row }">{{ row.reimbursement_amount ? row.reimbursement_amount.toFixed(2) : '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="payment_amount" label="支付金额" width="110" header-align="center" align="right">
          <template #default="{ row }">{{ row.payment_amount ? row.payment_amount.toFixed(2) : '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="write_off_amount" label="核销金额" width="110" header-align="center" align="right">
          <template #default="{ row }">{{ row.write_off_amount ? row.write_off_amount.toFixed(2) : '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="invoice_amount_excl_tax" label="发票不含税金额" width="140" header-align="center" align="right">
          <template #default="{ row }">{{ row.invoice_amount_excl_tax ? row.invoice_amount_excl_tax.toFixed(2) : '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="invoice_amount_incl_tax" label="发票含税金额" width="140" header-align="center" align="right">
          <template #default="{ row }">{{ row.invoice_amount_incl_tax ? row.invoice_amount_incl_tax.toFixed(2) : '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="current_process" label="当前处理环节" width="120" header-align="center" align="center" />
        <el-table-column prop="current_processor" label="当前处理人" width="120" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="physical_status" label="实物状态" width="100" header-align="center" align="center" />
        <el-table-column prop="physical_location" label="实物位置" width="120" header-align="center" align="center" />
        <el-table-column prop="document_type" label="单据类型" width="100" header-align="center" align="center" />
        <el-table-column prop="document_type_name" label="单据类型名称" width="140" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="supplier_code" label="供应商编号" width="160" header-align="center" align="center" />
        <el-table-column prop="supplier_name" label="供应商名称" width="180" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="create_doc_time" label="制单时间" width="160" header-align="center" align="center">
          <template #default="{ row }">{{ row.create_doc_time ? new Date(row.create_doc_time).toLocaleString('zh-CN') : '-' }}</template>
        </el-table-column>
        <el-table-column prop="submit_time" label="提交时间" width="160" header-align="center" align="center">
          <template #default="{ row }">{{ row.submit_time ? new Date(row.submit_time).toLocaleString('zh-CN') : '-' }}</template>
        </el-table-column>
        <el-table-column prop="internal_code" label="内码" width="300" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="shared_process" label="共享处理环节" width="120" header-align="center" align="center" />
        <el-table-column prop="shared_processor" label="共享处理人" width="150" header-align="center" align="center" show-overflow-tooltip />
        <el-table-column prop="payment_account" label="付款账号" width="150" header-align="center" align="center" />
        <el-table-column label="操作" width="150" header-align="center" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canEdit(row)" type="primary" link @click="showEditDialog(row)">编辑</el-button>
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
          <el-select v-model="form.project_id" placeholder="请选择项目（可选）" filterable clearable style="width: 100%;" @change="handleProjectSelect">
            <el-option 
              v-for="project in projects" 
              :key="project.id" 
              :label="`${project.innovation_code} - ${project.name}`" 
              :value="project.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="单据编号" prop="document_no">
          <el-input v-model="form.document_no" placeholder="请输入单据编号" />
        </el-form-item>
        <el-form-item label="费用类型" prop="expense_type">
          <el-select v-model="form.expense_type" placeholder="请选择费用类型" clearable style="width: 100%;">
            <el-option label="人工费用" value="labor" />
            <el-option label="直接投入费用" value="direct" />
            <el-option label="委托研发费用" value="outsourcing" />
            <el-option label="其他费用" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="报账金额" prop="reimbursement_amount">
          <el-input-number v-model="form.reimbursement_amount" :min="0" :precision="2" :controls="false" placeholder="请输入金额" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="发票不含税金额" prop="invoice_amount_excl_tax">
          <el-input-number v-model="form.invoice_amount_excl_tax" :min="0" :precision="2" :controls="false" placeholder="请输入不含税金额" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="发票含税金额" prop="invoice_amount_incl_tax">
          <el-input-number v-model="form.invoice_amount_incl_tax" :min="0" :precision="2" :controls="false" placeholder="请输入含税金额" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="摘要" prop="summary">
          <el-input v-model="form.summary" type="textarea" :rows="3" placeholder="请输入费用摘要" />
        </el-form-item>
        <el-form-item label="业务场景">
          <el-input v-model="form.business_scene" placeholder="业务场景" />
        </el-form-item>
        <el-form-item label="部门名称">
          <el-input v-model="form.department_name" placeholder="部门名称" />
        </el-form-item>
        <el-form-item label="单据状态">
          <el-input v-model="form.document_status" placeholder="单据状态" />
        </el-form-item>
        <el-form-item label="供应商名称">
          <el-input v-model="form.supplier_name" placeholder="供应商名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 导入Excel弹窗 -->
    <el-dialog v-model="importDialogVisible" title="导入费用记录" width="500px">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :on-change="handleFileChange"
        :limit="1"
        accept=".xlsx,.xls"
        drag
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">拖拽Excel文件到此处或<em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">仅支持.xlsx或.xls格式的Excel文件，需按财务模板格式</div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getExpenses, createExpense, updateExpense, deleteExpense, getProjectComparison, importExpenses, deleteAllExpenses } from '@/api/expense'
import { getProjects } from '@/api/project'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'

const userStore = useUserStore()

const loading = ref(false)
const statsLoading = ref(false)
const expenses = ref([])
const projects = ref([])
const comparisonData = ref([])
const dialogVisible = ref(false)
const importDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const importing = ref(false)
const editId = ref(null)
const formRef = ref(null)
const uploadRef = ref(null)
const selectedFile = ref(null)

const searchForm = reactive({ 
  project_code: '', 
  reimbursed_person_name: '', 
  is_project_expense: '', 
  document_no: '', 
  is_classified: null 
})
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const form = reactive({ 
  project_id: null, 
  project_code: '',
  document_no: '', 
  expense_type: '',
  reimbursement_amount: 0, 
  invoice_amount_excl_tax: 0,
  invoice_amount_incl_tax: 0,
  summary: '', 
  business_scene: '',
  department_name: '',
  document_status: '',
  supplier_name: ''
})

const rules = {
  document_no: [{ required: true, message: '请输入单据编号', trigger: 'blur' }],
  reimbursement_amount: [{ required: true, message: '请输入报账金额', trigger: 'blur' }],
  invoice_amount_excl_tax: [{ required: true, message: '请输入发票不含税金额', trigger: 'blur' }],
  invoice_amount_incl_tax: [{ required: true, message: '请输入发票含税金额', trigger: 'blur' }],
  summary: [{ required: true, message: '请输入费用摘要', trigger: 'blur' }]
}

const canEdit = (row) => {
  if (userStore.isAdmin) return true
  return row.reimbursed_by === userStore.userId
}

const canDelete = (row) => {
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
const resetSearch = () => { 
  Object.assign(searchForm, { 
    project_code: '', 
    reimbursed_person_name: '', 
    is_project_expense: '', 
    document_no: '', 
    is_classified: null 
  }); 
  handleSearch() 
}

const showCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, { 
    project_id: null, 
    project_code: '',
    document_no: '', 
    expense_type: '',
    reimbursement_amount: 0, 
    invoice_amount_excl_tax: 0,
    invoice_amount_incl_tax: 0,
    summary: '', 
    business_scene: '',
    department_name: '',
    document_status: '',
    supplier_name: ''
  })
  dialogVisible.value = true
}

const showEditDialog = (row) => {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, {
    project_id: row.project_id,
    project_code: row.project_code || '',
    document_no: row.document_no,
    expense_type: row.expense_type,
    reimbursement_amount: row.reimbursement_amount,
    invoice_amount_excl_tax: row.invoice_amount_excl_tax || 0,
    invoice_amount_incl_tax: row.invoice_amount_incl_tax || 0,
    summary: row.summary,
    business_scene: row.business_scene,
    department_name: row.department_name,
    document_status: row.document_status,
    supplier_name: row.supplier_name
  })
  dialogVisible.value = true
}

// 项目选择变化处理，自动填充创新项目编码
const handleProjectSelect = (projectId) => {
  if (projectId) {
    const selectedProject = projects.value.find(p => p.id === projectId)
    if (selectedProject && selectedProject.innovation_code) {
      form.project_code = selectedProject.innovation_code
    }
  } else {
    // 清空项目编码
    form.project_code = ''
  }
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
      fetchComparison()
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
    fetchComparison()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

const showImportDialog = () => {
  selectedFile.value = null
  importDialogVisible.value = true
}

const handleFileChange = (file) => {
  selectedFile.value = file
}

const handleImport = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请选择Excel文件')
    return
  }

  importing.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value.raw)

    const res = await importExpenses(formData)
    const result = res.data
    
    // 更详细的提示信息
    if (result.error_count > 0) {
      // 构建失败信息
      let errorMsg = `导入完成：新增${result.create_count || 0}条，更新${result.update_count || 0}条，失败${result.error_count}条`
      
      // 如果有错误信息，展示失败的单据编号
      if (result.errors && result.errors.length > 0) {
        errorMsg += '\n失败记录: ' + result.errors.join('、')
      }
      
      ElMessage.warning({
        message: errorMsg,
        duration: 5000,
        showClose: true
      })
    } else {
      ElMessage.success(
        result.message || `成功导入：新增${result.create_count || 0}条，更新${result.update_count || 0}条`
      )
    }
    
    importDialogVisible.value = false
    fetchExpenses()
    fetchComparison()
  } catch (error) {
    console.error('导入失败:', error)
    ElMessage.error('导入失败')
  } finally {
    importing.value = false
  }
}

const handleDeleteAll = async () => {
  ElMessageBox.confirm(
    '确定要删除所有费用记录吗？此操作不可恢复！',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(async () => {
      try {
        const res = await deleteAllExpenses()
        ElMessage.success(res.message || '删除成功')
        fetchExpenses()
        fetchComparison()
      } catch (error) {
        console.error('删除失败:', error)
        ElMessage.error('删除失败')
      }
    })
    .catch(() => {
      // 用户取消
    })
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
  margin-bottom: 20px; 
}

.pagination { 
  margin-top: 20px; 
  justify-content: flex-end; 
}
</style>
