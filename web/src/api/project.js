import request from '@/utils/request'

// 获取项目列表
export function getProjects(params) {
  return request.get('/projects', { params })
}

// 获取项目统计
export function getProjectStatistics() {
  return request.get('/projects/statistics')
}

// 创建项目
export function createProject(data) {
  return request.post('/projects', data)
}

// 获取项目详情
export function getProject(id) {
  return request.get(`/projects/${id}`)
}

// 更新项目
export function updateProject(id, data) {
  return request.put(`/projects/${id}`, data)
}

// 删除项目
export function deleteProject(id) {
  return request.delete(`/projects/${id}`)
}

// 获取项目阶段
export function getProjectPhases(projectId) {
  return request.get(`/projects/${projectId}/phases`)
}

// 添加项目阶段
export function addProjectPhase(projectId, data) {
  return request.post(`/projects/${projectId}/phases`, data)
}

// 更新项目阶段
export function updateProjectPhase(projectId, phaseId, data) {
  return request.put(`/projects/${projectId}/phases/${phaseId}`, data)
}

// 删除项目阶段
export function deleteProjectPhase(projectId, phaseId) {
  return request.delete(`/projects/${projectId}/phases/${phaseId}`)
}

// 获取项目成员
export function getProjectMembers(projectId) {
  return request.get(`/projects/${projectId}/members`)
}

// 添加项目成员
export function addProjectMember(projectId, data) {
  return request.post(`/projects/${projectId}/members`, data)
}

// 移除项目成员
export function removeProjectMember(projectId, memberId) {
  return request.delete(`/projects/${projectId}/members/${memberId}`)
}
