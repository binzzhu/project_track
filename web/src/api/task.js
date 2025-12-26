import request from '@/utils/request'

// 获取任务列表
export function getTasks(params) {
  return request.get('/tasks', { params })
}

// 获取我的任务
export function getMyTasks(params) {
  return request.get('/tasks/my', { params })
}

// 创建任务
export function createTask(data) {
  return request.post('/tasks', data)
}

// 批量创建任务
export function batchCreateTasks(data) {
  return request.post('/tasks/batch', data)
}

// 获取任务详情
export function getTask(id) {
  return request.get(`/tasks/${id}`)
}

// 更新任务
export function updateTask(id, data) {
  return request.put(`/tasks/${id}`, data)
}

// 删除任务
export function deleteTask(id) {
  return request.delete(`/tasks/${id}`)
}

// 更新任务状态
export function updateTaskStatus(id, status) {
  return request.put(`/tasks/${id}/status`, { status })
}

// 审核任务
export function reviewTask(id, data) {
  return request.post(`/tasks/${id}/review`, data)
}

// 获取任务统计
export function getTaskStatistics() {
  return request.get('/tasks/statistics')
}
