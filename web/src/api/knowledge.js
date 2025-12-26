import request from '@/utils/request'

// 获取知识库列表
export function getKnowledgeList(params) {
  return request.get('/knowledge', { params })
}

// 获取热门资料
export function getHotKnowledge(limit = 10) {
  return request.get('/knowledge/hot', { params: { limit } })
}

// 获取分类列表
export function getCategories() {
  return request.get('/knowledge/categories')
}

// 获取详情
export function getKnowledge(id) {
  return request.get(`/knowledge/${id}`)
}

// 上传资料
export function uploadKnowledge(formData) {
  return request.post('/knowledge/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 更新资料
export function updateKnowledge(id, data) {
  return request.put(`/knowledge/${id}`, data)
}

// 删除资料
export function deleteKnowledge(id) {
  return request.delete(`/knowledge/${id}`)
}

// 上传新版本
export function uploadNewVersion(id, formData) {
  return request.post(`/knowledge/${id}/version`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 获取版本历史
export function getVersions(id) {
  return request.get(`/knowledge/${id}/versions`)
}

// 创建分类
export function createCategory(data) {
  return request.post('/knowledge/categories', data)
}

// 更新分类
export function updateCategory(id, data) {
  return request.put(`/knowledge/categories/${id}`, data)
}

// 删除分类
export function deleteCategory(id) {
  return request.delete(`/knowledge/categories/${id}`)
}

// 下载资料
export function downloadKnowledge(id) {
  return request.get(`/knowledge/${id}/download`, {
    responseType: 'blob'
  })
}
