import request from '@/utils/request'

// 获取文档列表
export function getDocuments(params) {
  return request.get('/documents', { params })
}

// 获取文档详情
export function getDocument(id) {
  return request.get(`/documents/${id}`)
}

// 上传文档
export function uploadDocument(formData) {
  return request.post('/documents/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 更新文档
export function updateDocument(id, data) {
  return request.put(`/documents/${id}`, data)
}

// 删除文档
export function deleteDocument(id) {
  return request.delete(`/documents/${id}`)
}

// 归档文档
export function archiveDocument(id) {
  return request.post(`/documents/${id}/archive`)
}

// 获取下载URL（带token参数）
export function getDownloadUrl(id) {
  const token = localStorage.getItem('token')
  return `http://localhost:8080/api/documents/${id}/download?token=${token}`
}
