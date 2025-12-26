<template>
  <div class="profile">
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card class="profile-card">
          <div class="profile-header">
            <el-avatar :size="80" icon="UserFilled" />
            <h2>{{ userStore.user.name }}</h2>
            <el-tag>{{ userStore.user.role?.name }}</el-tag>
          </div>
          <el-descriptions :column="1" border class="profile-info">
            <el-descriptions-item label="用户名">{{ userStore.user.username }}</el-descriptions-item>
            <el-descriptions-item label="部门">{{ userStore.user.department || '-' }}</el-descriptions-item>
            <el-descriptions-item label="邮箱">{{ userStore.user.email || '-' }}</el-descriptions-item>
            <el-descriptions-item label="手机">{{ userStore.user.phone || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :span="16">
        <el-card>
          <template #header>
            <span>修改密码</span>
          </template>
          <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" style="max-width: 400px;">
            <el-form-item label="原密码" prop="old_password">
              <el-input v-model="form.old_password" type="password" placeholder="请输入原密码" show-password />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="form.new_password" type="password" placeholder="至少8位，包含大小写字母、数字和特殊符号" show-password />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input v-model="form.confirm_password" type="password" placeholder="请再次输入新密码" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="submitting" @click="handleSubmit">确认修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { changePassword } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const userStore = useUserStore()
const formRef = ref(null)
const submitting = ref(false)

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const validateConfirm = (rule, value, callback) => {
  if (value !== form.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码至少8位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' }
  ]
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      await changePassword({
        old_password: form.old_password,
        new_password: form.new_password
      })
      ElMessage.success('密码修改成功')
      form.old_password = ''
      form.new_password = ''
      form.confirm_password = ''
    } catch (error) {
      console.error('修改密码失败:', error)
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.profile { max-width: 1000px; }
.profile-card { text-align: center; }
.profile-header { padding: 20px 0; }
.profile-header h2 { margin: 15px 0 10px; }
.profile-info { margin-top: 20px; }
</style>
