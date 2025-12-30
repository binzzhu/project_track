import request from '@/utils/request'

// 获取费用记录列表
export function getExpenses(params) {
  return request.get('/expenses', { params })
}

// 获取费用记录详情
export function getExpense(id) {
  return request.get(`/expenses/${id}`)
}

// 创建费用记录
export function createExpense(data) {
  return request.post('/expenses', data)
}

// 更新费用记录
export function updateExpense(id, data) {
  return request.put(`/expenses/${id}`, data)
}

// 删除费用记录
export function deleteExpense(id) {
  return request.delete(`/expenses/${id}`)
}

// 获取费用统计
export function getExpenseStatistics(projectId) {
  return request.get('/expenses/statistics', {
    params: { project_id: projectId }
  })
}

// 获取项目费用对比
export function getProjectComparison() {
  return request.get('/expenses/comparison')
}

// 获取非研发项目费用统计
export function getNonProjectExpenseStats() {
  return request.get('/expenses/non-project-stats')
}

// 导入Excel费用记录
export function importExpenses(formData) {
  return request.post('/expenses/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 一键删除所有费用记录（仅管理员）
export function deleteAllExpenses() {
  return request.delete('/expenses/all')
}
