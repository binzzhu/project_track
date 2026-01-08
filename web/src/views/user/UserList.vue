<template>
  <div class="user-list">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="用户名/姓名" clearable @keyup.enter="handleSearch" style="width: 220px;" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="searchForm.role_id" placeholder="全部" clearable style="width: 220px;">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 220px;">
            <el-option label="启用" value="1" />
            <el-option label="禁用" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item class="search-button-item">
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <div class="action-bar">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon> 创建用户
      </el-button>
    </div>

    <el-card>
      <el-table v-loading="loading" :data="users" stripe>
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">{{ row.role?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="department" label="部门" width="150" />
        <el-table-column prop="email" label="邮箱" width="180" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="showEditDialog(row)">编辑</el-button>
            <el-button type="warning" link @click="handleResetPassword(row)">重置密码</el-button>
            <el-popconfirm v-if="row.username !== 'admin'" title="确定删除该用户吗？" width="200" @confirm="handleDelete(row.id)">
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
        @size-change="fetchUsers"
        @current-change="fetchUsers"
      />
    </el-card>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '创建用户'" width="500px" @open="handleDialogOpen">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEdit" placeholder="请输入用户名" autocomplete="off" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="至少8位，包含大小写字母、数字和特殊符号" show-password autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="角色" prop="role_id">
          <el-select v-model="form.role_id" placeholder="请选择角色">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="部门" prop="department">
          <el-select v-model="form.department" placeholder="请选择部门">
            <el-option label="BMS研发部" value="BMS研发部" />
            <el-option label="PACK研发部" value="PACK研发部" />
            <el-option label="综合部" value="综合部" />
            <el-option label="外部单位" value="外部单位" />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item v-if="isEdit" label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
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
import { getUsers, createUser, updateUser, deleteUser, resetPassword } from '@/api/user'
import { getRoles } from '@/api/auth'
import { ElMessage, ElMessageBox } from 'element-plus'

const loading = ref(false)
const users = ref([])
const roles = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref(null)
const formRef = ref(null)

const searchForm = reactive({ keyword: '', role_id: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const form = reactive({ username: '', password: '', name: '', role_id: null, department: '', email: '', phone: '', status: 1 })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  role_id: [{ required: true, message: '请选择角色', trigger: 'change' }],
  department: [{ required: true, message: '请选择部门', trigger: 'change' }]
}

const formatDateTime = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '-'

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await getUsers({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchForm.keyword,
      role_id: searchForm.role_id,
      status: searchForm.status
    })
    users.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取用户列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchRoles = async () => {
  const res = await getRoles()
  roles.value = res.data || []
}

const handleSearch = () => { pagination.page = 1; fetchUsers() }
const resetSearch = () => { searchForm.keyword = ''; searchForm.role_id = ''; searchForm.status = ''; handleSearch() }

const showCreateDialog = () => {
  isEdit.value = false
  editId.value = null
  Object.assign(form, { username: '', password: '', name: '', role_id: null, department: '', email: '', phone: '', status: 1 })
  dialogVisible.value = true
}

const showEditDialog = (row) => {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, { username: row.username, password: '', name: row.name, role_id: row.role_id, department: row.department, email: row.email, phone: row.phone, status: row.status })
  dialogVisible.value = true
}

// 对话框打开时清除表单验证状态
const handleDialogOpen = () => {
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        const { password, ...data } = form
        await updateUser(editId.value, data)
        ElMessage.success('更新成功')
      } else {
        await createUser(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchUsers()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (id) => {
  try {
    await deleteUser(id)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

const handleResetPassword = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要重置用户"${row.name}"的密码吗？`, '提示', { type: 'warning' })
    await resetPassword(row.id)
    ElMessage.success('密码已重置为: Reset@123')
  } catch {
    // 取消
  }
}

onMounted(() => {
  fetchRoles()
  fetchUsers()
})
</script>

<style scoped>
.search-card { margin-bottom: 20px; }
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
.action-bar { margin-bottom: 15px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
</style>
